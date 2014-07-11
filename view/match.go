package view

import (
  "appengine"
  "appengine/datastore"
  "errors"
  "fmt"
  "github.com/OwenDurni/loltools/model"
  "net/http"
  "strconv"
)

func MatchCreateHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  leagueId := args["leagueId"]
  
  _, userKey, err := model.GetUser(c)
  if HandleError(c, w, err) {
    return
  }
  userAcls := model.NewRequestorAclCache(userKey)
  
  league, leagueKey, err := model.LeagueById(c, leagueId)
  if HandleError(c, w, err) {
    return
  }

  teams, teamKeys, err := model.LeagueAllTeams(c, userAcls, league, leagueKey)
  if HandleError(c, w, err) {
    return
  }
  
  ctx := struct {
    ctxBase
    League
    Teams     []Team
    GroupAcls []GroupAcl
  }{}
  ctx.ctxBase.init(c)
  ctx.ctxBase.Title = fmt.Sprintf("loltools - %s - Create a Match", league.Name)

  ctx.League.Fill(league, leagueKey)

  ctx.Teams = make([]Team, len(teams))
  for i, t := range teams {
    ctx.Teams[i].Fill(t, teamKeys[i], leagueKey)
  }
  
  // Render
  err = RenderTemplate(w, "leagues/matches/create.html", "base", ctx)
  if HandleError(c, w, err) {
    return
  }
}

func ApiMatchCreateHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  
  var err error
  
  // Extract and validate parameters.
  leagueId := r.FormValue("league")
  tz := r.FormValue("tz")
  summary := r.FormValue("summary")
  description := r.FormValue("description")
  
  primaryTag := r.FormValue("primary-tag")
  if primaryTag == "" {
    err = errors.New("'primary-tag' must be non-empty")
    ApiHandleError(c, w, err)
    return
  }
  
  numGames, err := strconv.ParseInt(r.FormValue("num-games"), 10, 32)
  if ApiHandleError(c, w, err) {
    return
  }
  if numGames < 0 {
    err = errors.New(fmt.Sprintf("'num-games' must be non-negative: %d", numGames))
    ApiHandleError(c, w, err)
    return
  }
  
  officialDatetime, err := parseDatetime(r.FormValue("official-date"), r.FormValue("official-time"), tz)
  if ApiHandleError(c, w, err) {
    return
  }
  startDatetime, err := parseDatetime(r.FormValue("start-date"), r.FormValue("start-time"), tz)
  if ApiHandleError(c, w, err) {
    return
  }
  endDatetime, err := parseDatetime(r.FormValue("end-date"), r.FormValue("end-time"), tz)
  if ApiHandleError(c, w, err) {
    return
  }

  var homeTeams []string = r.PostForm["home-team"]
  var awayTeams []string = r.PostForm["away-team"]
  if len(homeTeams) != len(awayTeams) {
    err = errors.New(fmt.Sprintf(
      "length of home-team (%d) must equal length of away-team (%d)", len(homeTeams), len(awayTeams)))
    ApiHandleError(c, w, err)
    return
  }
  for i := range homeTeams {
    if homeTeams[i] == awayTeams[i] {
      err = errors.New("A team cannot be matched against itself.")
      ApiHandleError(c, w, err)
      return
    }
  }
  
  _, userKey, err := model.GetUser(c)
  if HandleError(c, w, err) {
    return
  }
  userAcls := model.NewRequestorAclCache(userKey)
  
  // Lookup league.
  league, leagueKey, err := model.LeagueById(c, leagueId)
  if ApiHandleError(c, w, err) {
    return
  }
  
  // Lookup all the teams involved.
  var homeTeamKeys []*datastore.Key = make([]*datastore.Key, len(homeTeams))
  var awayTeamKeys []*datastore.Key = make([]*datastore.Key, len(awayTeams))
  {
    type A struct {
      Ids  []string
      Keys []*datastore.Key
    }
    as := []*A{
      &A{Ids: homeTeams, Keys: homeTeamKeys},
      &A{Ids: awayTeams, Keys: awayTeamKeys},
    }
    for _, a := range as {
      for i, id := range a.Ids {
        _, a.Keys[i], err = model.TeamById(c, userAcls, league, leagueKey, id)
        if ApiHandleError(c, w, err) {
          return
        }
      }
    }
  }
  
  for i := range homeTeamKeys {
    match := &model.ScheduledMatch{
      Summary: summary,
      Description: description,
      PrimaryTag: primaryTag,
      TeamKeys: []*datastore.Key{homeTeamKeys[i], awayTeamKeys[i]},
      NumGames: int(numGames),
      OfficialDatetime: officialDatetime,
      DateEarliest: startDatetime,
      DateLatest: endDatetime,
    }
    
    err = model.CreateScheduledMatch(c, userAcls, league, leagueKey, match)
    if ApiHandleError(c, w, err) {
      return
    }
  }
  
  HttpReplyOkEmpty(w)
}
