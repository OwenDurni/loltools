package task

import (
  "appengine"
  "appengine/datastore"
  "errors"
  "fmt"
  "github.com/OwenDurni/loltools/model"
  "github.com/OwenDurni/loltools/riot"
  "net/http"
  "strconv"
  "time"
)

func FetchMatchHistoryHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  region := r.FormValue("region")
  riotSummonerId, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
  if err != nil {
    ReportPermanentError(c, w, errors.New(fmt.Sprintf("Could not parse id param: %v", err)))
    return
  }
  
  riotApiKey, err := model.GetRiotApiKey(c)
  if err != nil {
    ReportPermanentError(c, w, errors.New(fmt.Sprintf("Could not get riotApiKey: %v", err)))
    return
  }
  
  riotData, err := riot.GameStatsForPlayer(c, riotApiKey.Key, region, riotSummonerId)
  if err != nil {
    ReportPermanentError(
      c, w, errors.New(fmt.Sprintf("Error in riot.GameStatsForPlayer(): %v", err)))
    return
  }
  
  if len(riotData.Games) == 0 {
    ReportPermanentError(
      c, w, errors.New(fmt.Sprintf(
        "No matches in history for summoner in region %s with id %d",
        region, riotSummonerId)))
    return
  }
  
  player, playerKey, err := model.GetOrCreatePlayerByRiotId(c, region, riotSummonerId)
  if err != nil {
    ReportPermanentError(
      c, w, errors.New(fmt.Sprintf(
        "Error creating Player(%s-%d) in datastore: %s",
        region, riotSummonerId, err.Error())))
    return
  }
  
  // If we have no issues fetching all the game entities then we should mark game stats
  // we know are no longer available as such. The map is a set of gameKey.Encode()
  // for each gameKey in the fetched match history.
  updateStatsNotAvailable := true
  encodedAvailableGameKeySet := make(map[string]struct{})
  
  // Cache each game stats in datastore.
  for _, gameData := range riotData.Games {
    game, gameKey, err := model.GetOrCreateGame(c, region, gameData.GameId)
    if err != nil {
      updateStatsNotAvailable = false
      ReportPermanentError(
        c, w, errors.New(fmt.Sprintf(
          "Error creating Game(%s-%d) in datastore: %s",
          region, gameData.GameId, err.Error())))
      continue
    }
    encodedAvailableGameKeySet[gameKey.Encode()] = struct{}{}
    
    playerGameStats := new(model.PlayerGameStats)
    playerGameStats.GameKey = gameKey
    playerGameStats.PlayerKey = playerKey
    playerGameStats.GameStartDateTime = (time.Time)(gameData.CreateDate)
    playerGameStats.NotAvailable = false
    playerGameStats.Saved = true
    playerGameStats.RiotData = gameData
    
    // Possibly also update some stats for the game itself.
    if game.UpdateLocalFromPlayerGameStats(playerGameStats) {
      if _, err = datastore.Put(c, gameKey, game); err != nil {
        ReportPermanentError(
          c, w, errors.New(fmt.Sprintf(
            "Error writing updated Game(%s) in datastore: %s",
            game.Id(), err.Error())))
        continue
      }
    }
    
    // Save the stats.
    playerGameStatsKey := model.KeyForPlayerGameStats(c, game, player)
    if _, err = datastore.Put(c, playerGameStatsKey, playerGameStats); err != nil {
      ReportPermanentError(
        c, w, errors.New(fmt.Sprintf(
          "Error writing PlayerGameStats(%v) in datastore: %s",
          playerGameStatsKey, err.Error())))
      continue
    }
    
    // TODO(durni): Create stub PlayerGameStats for all the other players in the
    // game if the entities don't exist yet.
  }
  
  // If we have any incomplete PlayerGameStats entities that are not in
  // this match history and are at least 1 day old, mark them as expired.
  if updateStatsNotAvailable {
    now := time.Now().UTC()
    q := datastore.NewQuery("PlayerGameStats").
           Filter("NotAvailable =", false).
           KeysOnly()
    playerGameStatsKeys, err := q.GetAll(c, nil)
    if err != nil {
      ReportPermanentError(
        c, w, errors.New(fmt.Sprintf(
          "Error updating PlayerGameStats.NotAvailable: %s", err.Error())))
      // Intentionally just continue as if the error didn't happen.
    }
    for i := range playerGameStatsKeys {
      err = datastore.RunInTransaction(c, func(c appengine.Context) error {
        var playerGameStats model.PlayerGameStats
        err := datastore.Get(c, playerGameStatsKeys[i], &playerGameStats)
        if err != nil {
          return err
        }
        // Don't do anything for games that aren't at least a day old.
        if now.Sub(playerGameStats.GameStartDateTime).Hours() < 24 {
          return nil
        }
        // If this game isn't in the available game set, mark it as unavailable.
        _, exists := encodedAvailableGameKeySet[playerGameStatsKeys[i].Encode()]
        if !exists {
          playerGameStats.NotAvailable = true
          _, err = datastore.Put(c, playerGameStatsKeys[i], playerGameStats)
        }
        return err
      }, nil)
      if err != nil {
        ReportPermanentError(
          c, w, errors.New(fmt.Sprintf(
            "Error updating PlayerGameStats.NotAvailable: %s", err.Error())))
        // Intentionally just continue as if the error didn't happen.
      }
    }
  }
  
  // Output the raw data to http for debugging purposes.
  fmt.Fprintf(w, "%+v", riotData)
}