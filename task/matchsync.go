package task

import (
  "appengine"
  "appengine/datastore"
  "appengine/taskqueue"
  "fmt"
  "github.com/OwenDurni/loltools/model"
  "github.com/OwenDurni/loltools/model/tags"
  "io"
  "net/http"
  "net/url"
)

func QueueMatchSync(c appengine.Context, matchKey *datastore.Key) {
  args := &url.Values{}
  args.Add("match", matchKey.Encode())
  task := taskqueue.NewPOSTTask("/task/match/sync", *args)
  task.RetryOptions = new(taskqueue.RetryOptions)
  task.RetryOptions.RetryLimit = 1
  taskqueue.Add(c, task, "")
}

func AllMatchSync(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  
  q := datastore.NewQuery("ScheduledMatch").KeysOnly()
  matchKeys, err := q.GetAll(c, nil)
  if ReportError(c, w, err) {
    return
  }
  
  for _, matchKey := range matchKeys {
    QueueMatchSync(c, matchKey)
  }
  
  fmt.Fprintf(w, "<html><body><pre>")
  fmt.Fprintf(w, "Queueing /task/match/sync for %d match(es)\n", len(matchKeys))
  fmt.Fprintf(w, "</pre></body></html>")
}

func MatchSync(w http.ResponseWriter, r *http.Request, args map[string]string) {
  fmt.Fprintf(w, "<html><body><pre>")
  c := appengine.NewContext(r)
  
  matchKey, err := datastore.DecodeKey(r.FormValue("match"))
  if ReportError(c, w, err) {
    return
  }
  
  match := new(model.ScheduledMatch)
  err = datastore.Get(c, matchKey, match)
  if ReportError(c, w, err) {
    return
  }
  
  homeTeamKey := match.HomeTeam()
  awayTeamKey := match.AwayTeam()
  
  // Phase 1: Tag games that look like they could be for this match.
  tagGamesInMatchWindow(c, w, homeTeamKey, awayTeamKey, match)
  fmt.Fprintf(w, "</pre></body></html>")
}

func tagGamesInMatchWindow(
  c appengine.Context,
  w io.Writer,
  homeTeamKey *datastore.Key,
  awayTeamKey *datastore.Key,
  match *model.ScheduledMatch) error {
  gameKeys, err := getGamesInMatchWindow(c, homeTeamKey, awayTeamKey, match)
  if err != nil {
    return err
  }
  fmt.Fprintf(w, "Found %d possible match results:\n", len(gameKeys))
  for _, gameKey := range gameKeys {
    uri := model.GameUri(gameKey)
    fmt.Fprintf(w, "  <a href=\"%s\">%s</a>\n", uri, uri)
  }
  fmt.Fprintf(w, "\n")
  
  leagueKey := homeTeamKey.Parent()
  for _, gameKey := range gameKeys {
    err := model.AddGameTag(
      c, nil, leagueKey, gameKey,
      tags.AutomaticallyDetectedMatchResult(match.PrimaryTag),
      tags.ReasonNotApplicable)
    if err != nil {
      return err
    }
  }
  return nil
}

// Gets a list of gameKeys played by the two specified teams within the bounds of
// the specified match.
func getGamesInMatchWindow(
  c appengine.Context,
  homeTeamKey *datastore.Key,
  awayTeamKey *datastore.Key,
  match *model.ScheduledMatch) ([]*datastore.Key, error) {
  homeMatchPossibleGameKeys, err := getGamesInMatchWindowByTeam(c, homeTeamKey, match)
  if err != nil {
    return nil, err
  }
  awayMatchPossibleGameKeys, err := getGamesInMatchWindowByTeam(c, awayTeamKey, match)
  if err != nil {
    return nil, err
  }
  ret := make([]*datastore.Key, 0, 8)
  gameKeySet := make(map[string]struct{})
  for _, k := range homeMatchPossibleGameKeys {
    gameKeySet[k.Encode()] = struct{}{}
  }
  for _, k := range awayMatchPossibleGameKeys {
    if _, exists := gameKeySet[k.Encode()]; exists {
      ret = append(ret, k)
    }
  }
  return ret, nil
}

func getGamesInMatchWindowByTeam(
  c appengine.Context,
  teamKey *datastore.Key,
  match *model.ScheduledMatch) ([]*datastore.Key, error) {
  q := datastore.NewQuery("GameByTeam").
    Ancestor(teamKey.Parent()).
    Filter("TeamKey =", teamKey).
    Filter("DateTime >=", match.DateEarliest).
    Filter("DateTime <=", match.DateLatest).
    Project("GameKey")
  var gameByTeams []*model.GameByTeam
  _, err := q.GetAll(c, &gameByTeams)
  if err != nil {
    return nil, err
  }
  gameKeys := make([]*datastore.Key, len(gameByTeams))
  for i := range gameByTeams {
    gameKeys[i] = gameByTeams[i].GameKey
  }
  return gameKeys, nil
}
