package view

import (
  "appengine"
  "appengine/datastore"
  "fmt"
  "github.com/OwenDurni/loltools/model"
  "net/http"
)

type League struct {
  Name   string
  Owner  string
  Id     string
  Uri    string
  Region string
}

func (l *League) Fill(m *model.League, k *datastore.Key) *League {
  l.Name = m.Name
  l.Id = model.EncodeKeyShort(k)
  l.Uri = model.LeagueUri(k)
  l.Region = model.RegionNA
  if m.Region != "" {
    l.Region = m.Region
  }
  return l
}

type Team struct {
  Name string
  Id   string
  Uri  string

  Wins   int
  Losses int
}

func (t *Team) Fill(
  team *model.Team, teamKey *datastore.Key, leagueKey *datastore.Key) *Team {
  t.Name = team.Name
  t.Id = model.EncodeKeyShort(teamKey)
  t.Uri = model.LeagueTeamUri(leagueKey, teamKey)
  return t
}

func LeagueIndexHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)

  // Lookup data from backend.
  _, userKey, err := model.GetUser(c)
  if HandleError(c, w, err) { return }
  userAcls := model.NewRequestorAclCache(userKey)

  leagues, leagueKeys, err := model.LeaguesForUser(c, userAcls)
  if HandleError(c, w, err) { return }

  // Populate view context.
  ctx := struct {
    ctxBase
    MyLeagues []*League
  }{}
  ctx.ctxBase.init(c)

  ctx.MyLeagues = make([]*League, len(leagues))
  for i := range leagues {
    league := new(League).Fill(leagues[i], leagueKeys[i])
    if owner, err := model.GetUserByKey(c, leagues[i].Owner); err == nil {
      league.Owner = owner.Email
    } else {
      league.Owner = err.Error()
    }
    ctx.MyLeagues[i] = league
  }

  // Render
  err = RenderTemplate(w, "leagues/index.html", "base", ctx)
  if HandleError(c, w, err) { return }
}

func LeagueViewHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  leagueId := args["leagueId"]

  _, userKey, err := model.GetUser(c)
  if HandleError(c, w, err) { return }
  userAcls := model.NewRequestorAclCache(userKey)

  league, leagueKey, err := model.LeagueById(c, leagueId)
  if HandleError(c, w, err) { return }

  teams, teamKeys, err := model.LeagueAllTeams(c, userAcls, league, leagueKey)
  if HandleError(c, w, err) { return }
  
  // Populate view context.
  ctx := struct {
    ctxBase
    League
    Teams []Team
    GroupAcls []GroupAcl
  }{}
  ctx.ctxBase.init(c)
  ctx.ctxBase.Title = fmt.Sprintf("loltools - %s", league.Name)

  ctx.League.Fill(league, leagueKey)

  ctx.Teams = make([]Team, len(teams))
  for i, t := range teams {
    ctx.Teams[i].Fill(t, teamKeys[i], leagueKey)
  }

  if *league.Owner == *userKey {
    groupKeys, groups, perms, err := userAcls.PermissionMapFor(c, leagueKey)
    if HandleError(c, w, err) { return }
    
    ctx.GroupAcls = make([]GroupAcl, len(groups))
    for i := range groups {      
      vg := new(Group).Fill(groups[i], groupKeys[i])
      ctx.GroupAcls[i].Fill(vg, perms[i])
    }
  }
  
  // Render
  err = RenderTemplate(w, "leagues/view.html", "base", ctx)
  if HandleError(c, w, err) { return }
}

func ApiLeagueCreateHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  _, leagueKey, err := model.CreateLeague(c, r.FormValue("name"))
  if HandleError(c, w, err) { return }
  
  HttpReplyResourceCreated(w, model.LeagueUri(leagueKey))
}

func ApiLeagueAddTeamHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  leagueId := r.FormValue("league")
  teamName := r.FormValue("team")

  _, userKey, err := model.GetUser(c)
  if HandleError(c, w, err) { return }
  
  userAcls := model.NewRequestorAclCache(userKey)

  _, leagueKey, err := model.LeagueById(c, leagueId)
  if HandleError(c, w, err) { return }

  _, teamKey, err := model.LeagueAddTeam(c, userAcls, leagueId, teamName)
  if HandleError(c, w, err) { return }

  HttpReplyResourceCreated(w, model.LeagueTeamUri(leagueKey, teamKey))
}
