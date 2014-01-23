package task

import (
  "appengine"
  "appengine/datastore"
  "fmt"
  "github.com/OwenDurni/loltools/model"
  "github.com/OwenDurni/loltools/riot"
  "net/http"
  "time"
)

// Note(durni): This is optimized to minimize the number of datastore write ops at
// the cost of potentially increased network ops into the riot api. Datastore write
// ops are expensive (in dollars) relative to network ops.
func FetchTeamMatchHistoryHandler(
  w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  leagueId := r.FormValue("league")
  teamId := r.FormValue("team")
  
  league, leagueKey, err := model.LeagueById(c, nil, leagueId)
  if ReportError(c, w, err) { return }
  
  region := league.Region
  
  _, teamKey, err := model.TeamById(c, nil, leagueKey, teamId)
  if ReportError(c, w, err) { return }
  
  players, _, err := model.TeamAllPlayers(
    c, nil, leagueKey, teamKey, model.KeysAndEntities)
  if ReportError(c, w, err) { return }
  
  riotApiKey, err := model.GetRiotApiKey(c)
  if ReportError(c, w, err) { return }
  
  // First gather games from all players on the team.
  collectiveGameStats := new(model.CollectiveGameStats)
  for _, player := range players {
    if err := model.RiotApiRateLimiter.TryConsume(c, 1); err != nil {
      // Hitting rate limit: break to finish storing what we have already fetched.
      ReportError(c, w, err)
      break
    }
    recentGamesDto, err := riot.GameStatsForPlayer(c, riotApiKey.Key, region, player.RiotId)
    if ReportError(c, w, err) { return }
    
    for _, gameDto := range recentGamesDto.Games {
      gameId := model.MakeGameId(region, gameDto.GameId)
      gameDtoCopy := gameDto
      collectiveGameStats.Add(region, gameId, player.RiotId, &gameDtoCopy)
    }
  }
  
  // Filter to only the games that contain at least 3 members of the team.
  collectiveGameStats.FilterToGamesWithAtLeast(3, players)
  
  // Write to datastore.
  collectiveGameStats.ForEachGame(func(gameId string,
                                       sampleRiotSummonerId int64,
                                       sampleStat *riot.GameDto) {
    gameKey := model.KeyForGameId(c, gameId)
    err := model.EnsureGameExists(c, region, gameKey, sampleRiotSummonerId, sampleStat)
    if ReportError(c, w, err) { return }
    
    err = model.LeagueAddGameByTeam(
      c, leagueKey, gameKey, teamKey, (time.Time)(sampleStat.CreateDate))
    if ReportError(c, w, err) { return }
  })
  
  collectiveGameStats.ForEachStat(func(gameId string, riotSummonerId int64, stat *riot.GameDto) {
    if stat == nil { return }
    gameKey := model.KeyForGameId(c, gameId)
    playerKey := model.KeyForPlayer(c, region, riotSummonerId)
    playerId := model.MakePlayerId(region, riotSummonerId)
    statsKey := model.KeyForPlayerGameStatsId(c, gameId, playerId)
    
    err = datastore.RunInTransaction(c, func (c appengine.Context) error {
      playerGameStats := new(model.PlayerGameStats)
      err := datastore.Get(c, statsKey, playerGameStats)
      if err != nil && err != datastore.ErrNoSuchEntity {
        return err
      }
      // Only write if the entity hasn't been saved yet.
      if !playerGameStats.Saved {
        playerGameStats.GameKey = gameKey
        playerGameStats.PlayerKey = playerKey
        playerGameStats.Saved = true
        playerGameStats.NotAvailable = false
        playerGameStats.RiotData = stat.Stats
        _, err = datastore.Put(c, statsKey, playerGameStats)
        return err
      }
      // Nothing to write.
      return nil
    }, nil)
    if ReportError(c, w, err) { return }
  })
  
  // Write some debug info to the response.
  fmt.Fprintf(w, "<html><body><pre>")
  fmt.Fprintf(
    w, "Found %d games with at least 3 players from:\n", collectiveGameStats.Size())
  for _, player := range players {
    fmt.Fprintf(w, "  %s (%d)\n", player.Summoner, player.RiotId)
  }
  collectiveGameStats.WriteDebugStringTo(w)
  fmt.Fprintf(w, "</pre></body></html>")
}