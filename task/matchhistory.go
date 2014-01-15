package task

import (
  "appengine"
  //"appengine/datastore"
  //"appengine/taskqueue"
  "fmt"
  "github.com/OwenDurni/loltools/model"
  "github.com/OwenDurni/loltools/riot"
  "github.com/OwenDurni/loltools/util/errwrap"
  "net/http"
  //"net/url"
  //"strconv"
  //"time"
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
  if err != nil {
    ReportPermanentError(
      c, w, fmt.Sprintf("Failed to lookup league/%s: %v", leagueId, err))
    return
  }
  region := league.Region
  
  _, teamKey, err := model.TeamById(c, nil, leagueKey, teamId)
  if err != nil {
    ReportPermanentError(
      c, w, fmt.Sprintf("Failed to lookup league/%s/team/%s: %v", leagueId, teamId, err))
    return
  }
  
  players, _, err := model.TeamAllPlayers(
    c, nil, leagueKey, teamKey, model.KeysAndEntities)
  if err != nil {
    ReportPermanentError(
      c, w,
      fmt.Sprintf("Failed to lookup league/%s/teas/%s/players: %v", leagueId, teamId, err))
    return
  }
  
  riotApiKey, err := model.GetRiotApiKey(c)
  if err != nil {
    ReportPermanentError(c, w, fmt.Sprintf("Failed to lookup RiotApiKey: %v", err))
    return
  }
  
  // First gather games from all players on the team.
  collectiveGameStats := new(model.CollectiveGameStats)
  for _, player := range players {
    if err := model.RiotApiRateLimiter.TryConsume(c, 1); err != nil {
      // Hitting rate limit: break to finish storing what we have already fetched.
      ReportPermanentError(c, w, fmt.Sprintf("RiotRateLimit: %v", errwrap.Wrap(err)))
      break
    }
    recentGamesDto, err := riot.GameStatsForPlayer(c, riotApiKey.Key, region, player.RiotId)
    if err != nil {
      ReportPermanentError(c, w, fmt.Sprintf("Error in riot.GameStatsForPlayer(): %v", err))
      return
    }
    for _, gameDto := range recentGamesDto.Games {
      gameId := model.MakeGameId(region, gameDto.GameId)
      collectiveGameStats.Add(region, gameId, player.RiotId, gameDto)
    }
  }
  
  // Filter to only the games that contain at least 3 members of the team.
  collectiveGameStats.FilterToGamesWithAtLeast(3, players)
  
  // Write to datastore.
  collectiveGameStats.Save(c)
  
  // Write some debug info to the response.
  fmt.Fprintf(
    w, "Found %d games with at least 3 players from:\n", collectiveGameStats.Size())
  for _, player := range players {
    fmt.Fprintf(w, "  %s (%d)\n", player.Summoner, player.RiotId)
  }
  collectiveGameStats.WriteDebugStringTo(w)
}