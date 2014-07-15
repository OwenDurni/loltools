package task

import (
  "appengine"
  "appengine/datastore"
  "appengine/taskqueue"
  "fmt"
  "github.com/OwenDurni/loltools/model"
  "github.com/OwenDurni/loltools/riot"
  "github.com/OwenDurni/loltools/util/errwrap"
  "net/http"
  "net/url"
  "time"
)

func MissingGameStats(w http.ResponseWriter, r *http.Request, args map[string]string) {
  fmt.Fprintf(w, "<html><body><pre>")
  c := appengine.NewContext(r)

  riotApiKey, err := model.GetRiotApiKey(c)
  if ReportError(c, w, errwrap.Wrap(err)) {
    return
  }

  limit := 20
  q := datastore.NewQuery("PlayerGameStats").
    Filter("Saved =", false).
    Filter("NotAvailable =", false).
    Order("PlayerKey").
    Limit(limit)

  var stats []*model.PlayerGameStats
  statKeys, err := q.GetAll(c, &stats)
  if ReportError(c, w, errwrap.Wrap(err)) {
    return
  }

  fmt.Fprintf(
    w, "Attempting to fill %d PlayerGameStats (limit is %d)\n", len(statKeys), limit)

  playerKeyMap := make(map[string]*datastore.Key)
  gameKeyMap := make(map[string]*datastore.Key)
  for _, stat := range stats {
    playerKeyMap[stat.PlayerKey.Encode()] = stat.PlayerKey
    gameKeyMap[stat.GameKey.Encode()] = stat.GameKey
  }

  players := make([]*model.Player, 0, len(playerKeyMap))
  playerKeys := make([]*datastore.Key, 0, len(playerKeyMap))
  playerMap := make(map[string]*model.Player)
  for _, key := range playerKeyMap {
    region, riotSummonerId, err := model.SplitPlayerKey(key)
    if ReportError(c, w, errwrap.Wrap(err)) {
      return
    }
    _, _, err = model.GetOrCreatePlayerByRiotId(c, region, riotSummonerId)
    if ReportError(c, w, errwrap.Wrap(err)) {
      return
    }
    players = append(players, new(model.Player))
    playerKeys = append(playerKeys, key)
  }
  err = datastore.GetMulti(c, playerKeys, players)
  if ReportError(c, w, errwrap.Wrap(err)) {
    return
  }
  for i := range players {
    playerMap[playerKeys[i].Encode()] = players[i]
    fmt.Fprintf(w, "Including player: %s\n", players[i].Summoner)
  }

  games := make([]*model.Game, 0, len(gameKeyMap))
  gameKeys := make([]*datastore.Key, 0, len(gameKeyMap))
  gameMap := make(map[string]*model.Game)
  for _, key := range gameKeyMap {
    games = append(games, new(model.Game))
    gameKeys = append(gameKeys, key)
  }
  err = datastore.GetMulti(c, gameKeys, games)
  if ReportError(c, w, errwrap.Wrap(err)) {
    return
  }
  for i := range games {
    gameMap[gameKeys[i].Encode()] = games[i]
  }

  collectiveGameStats := new(model.CollectiveGameStats)
  for _, player := range players {
    if err := model.RiotApiRateLimiter.Consume(c, 1); err != nil {
      // Hitting rate limit: break to finish storing what we have already fetched.
      ReportError(c, w, errwrap.Wrap(err))
      return
    }
    recentGamesDto, err := riot.GameStatsForPlayer(
      c, riotApiKey.Key, player.Region, player.RiotId)
    if ReportError(c, w, errwrap.Wrap(err)) {
      return
    }

    for _, gameDto := range recentGamesDto.Games {
      gameId := model.MakeGameId(player.Region, gameDto.GameId)
      collectiveGameStats.Add(gameId, player.RiotId, &gameDto)
    }
  }

  foundCount := 0
  expiredCount := 0

  for i, stat := range stats {
    statKey := statKeys[i]
    game := gameMap[stat.GameKey.Encode()]
    player := playerMap[stat.PlayerKey.Encode()]

    riotData := collectiveGameStats.Lookup(game.Id(), player.RiotId)

    err = datastore.RunInTransaction(c, func(c appengine.Context) error {
      playerGameStats := new(model.PlayerGameStats)
      err := datastore.Get(c, statKey, playerGameStats)
      if err != nil {
        return errwrap.Wrap(err)
      }
      // Only write if the entity hasn't been saved yet.
      if !playerGameStats.Saved {
        if riotData != nil {
          playerGameStats.Saved = true
          playerGameStats.RiotData = riotData.Stats
          playerGameStats.NotAvailable = false
          foundCount++
        } else {
          playerGameStats.Saved = false
          playerGameStats.NotAvailable = true
          expiredCount++
        }
        _, err = datastore.Put(c, statKey, playerGameStats)
        return errwrap.Wrap(err)
      }
      // Nothing to write.
      return nil
    }, nil)
    if ReportError(c, w, errwrap.Wrap(err)) {
      return
    }
  }

  fmt.Fprintf(w, "  Found: %d\n", foundCount)
  fmt.Fprintf(w, "  Not Available: %d\n", expiredCount)
  fmt.Fprintf(w, "</pre></body></html>")
}

func AllTeamHistories(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)

  q := datastore.NewQuery("League")
  var leagues []*model.League
  leagueKeys, err := q.GetAll(c, &leagues)
  if ReportError(c, w, err) {
    return
  }

  teamCount := 0

  for i := range leagues {
    teams, teamKeys, err := model.LeagueAllTeams(c, nil, leagues[i], leagueKeys[i])
    if ReportError(c, w, err) {
      return
    }
    for j := range teams {
      args := &url.Values{}
      args.Add("league", model.EncodeKeyShort(leagueKeys[i]))
      args.Add("team", model.EncodeKeyShort(teamKeys[j]))
      task := taskqueue.NewPOSTTask("/task/riot/get/team/history", *args)
      task.RetryOptions = new(taskqueue.RetryOptions)
      task.RetryOptions.RetryLimit = 1
      taskqueue.Add(c, task, "")
      teamCount++
    }
  }

  fmt.Fprintf(w, "<html><body><pre>")
  fmt.Fprintf(w, "Queueing /task/riot/get/team/history for %d team(s)\n", teamCount)
  fmt.Fprintf(w, "</pre></body></html>")
}

// Note(durni): This is optimized to minimize the number of datastore write ops at
// the cost of potentially increased network ops into the riot api. Datastore write
// ops are expensive (in dollars) relative to network ops.
func FetchTeamMatchHistoryHandler(
  w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  leagueId := r.FormValue("league")
  teamId := r.FormValue("team")

  league, leagueKey, err := model.LeagueById(c, leagueId)
  if ReportError(c, w, err) {
    return
  }

  region := league.Region

  _, teamKey, err := model.TeamById(c, nil, league, leagueKey, teamId)
  if ReportError(c, w, err) {
    return
  }

  players, _, err := model.TeamAllPlayers(
    c, nil, league, leagueKey, teamKey, model.KeysAndEntities)
  if ReportError(c, w, err) {
    return
  }

  riotApiKey, err := model.GetRiotApiKey(c)
  if ReportError(c, w, err) {
    return
  }

  // First gather games from all players on the team.
  collectiveGameStats := new(model.CollectiveGameStats)
  for _, player := range players {
    if err := model.RiotApiRateLimiter.Consume(c, 1); err != nil {
      // Hitting rate limit: break to finish storing what we have already fetched.
      ReportError(c, w, err)
      break
    }
    recentGamesDto, err := riot.GameStatsForPlayer(c, riotApiKey.Key, region, player.RiotId)
    if ReportError(c, w, err) {
      return
    }

    for _, gameDto := range recentGamesDto.Games {
      gameId := model.MakeGameId(region, gameDto.GameId)
      
      var gameDtoCopy riot.GameDto = gameDto
      collectiveGameStats.Add(gameId, player.RiotId, &gameDtoCopy)
    }
  }

  // Filter to only the games that contain at least 3 members of the team.
  collectiveGameStats.FilterToGamesWithPlayersAtLeast(3)

  // Write to datastore.
  collectiveGameStats.ForEachGame(func(
      gameId string,
      gameStats *model.GameStats,
      sampleRiotSummonerId int64,
      sampleStat *riot.GameDto) {
    gameKey := model.KeyForGameId(c, gameId)
    err := model.EnsureGameExists(c, region, gameKey, sampleRiotSummonerId, sampleStat)
    if ReportError(c, w, err) {
      return
    }

    err = model.LeagueAddGameByTeam(c, leagueKey, &model.GameByTeam{
      GameKey: gameKey,
      TeamKey: teamKey,
      DateTime: (time.Time)(sampleStat.CreateDate),
      RiotTeamIds: gameStats.GetTeamsWithCountAtLeast(3),
    })
    if ReportError(c, w, err) {
      return
    }
  })

  collectiveGameStats.ForEachStat(func(gameId string, riotSummonerId int64, stat *riot.GameDto) {
    gameKey := model.KeyForGameId(c, gameId)
    playerKey := model.KeyForPlayer(c, region, riotSummonerId)
    playerId := model.MakePlayerId(region, riotSummonerId)
    statsKey := model.KeyForPlayerGameStatsId(c, gameId, playerId)

    err = datastore.RunInTransaction(c, func(c appengine.Context) error {
      playerGameStats := new(model.PlayerGameStats)
      err := datastore.Get(c, statsKey, playerGameStats)
      if err != nil && err != datastore.ErrNoSuchEntity {
        return err
      }
      // Only write if the entity hasn't been saved yet.
      if !playerGameStats.Saved {
        playerGameStats.GameKey = gameKey
        playerGameStats.PlayerKey = playerKey
        playerGameStats.Saved = (stat != nil)
        playerGameStats.NotAvailable = false
        if stat != nil {
          playerGameStats.RiotData = stat.Stats
        }
        _, err = datastore.Put(c, statsKey, playerGameStats)
        return err
      }
      // Nothing to write.
      return nil
    }, nil)
    if ReportError(c, w, err) {
      return
    }
  })

  // Write some debug info to the response.
  fmt.Fprintf(w, "<html><body><pre>")
  fmt.Fprintf(
    w, "Found %d games with at least 3 players from:\n", collectiveGameStats.Size())
  for _, player := range players {
    fmt.Fprintf(w, "  %s (%d)\n", player.Summoner, player.RiotId)
  }
  fmt.Fprintf(w, collectiveGameStats.DebugString())
  fmt.Fprintf(w, "</pre></body></html>")
}
