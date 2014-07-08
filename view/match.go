package view

import (
  "appengine"
  "fmt"
  "github.com/OwenDurni/loltools/model"
  "net/http"
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
