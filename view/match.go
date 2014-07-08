package view

import (
  "appengine"
  "appengine/datastore"
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
  
  //leagueId := r.FormValue("league")
  tz := r.FormValue("tz")
  summary := r.FormValue("summary")
  description := r.FormValue("description")
  primaryTag := r.FormValue("primary-tag")
  numGames, err := strconv.ParseInt(r.FormValue("num-games"), 10, 32)
  if ApiHandleError(c, w, err) {
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
  
  match := &model.ScheduledMatch{
    Summary: summary,
    Description: description,
    PrimaryTag: primaryTag,
    TeamKeys: make([]*datastore.Key, 2),
    NumGames: int(numGames),
    OfficialDatetime: &officialDatetime,
    DateEarliest: &startDatetime,
    DateLatest: &endDatetime,
  }
  
  c.Debugf("Match: %v", match)
  c.Debugf("Home Teams: %v", homeTeams)
  c.Debugf("Away Teams: %v", awayTeams)
  
  HttpReplyOkEmpty(w)
}
