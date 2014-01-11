package view

import (
  "appengine"
  "errors"
  "fmt"
  "github.com/OwenDurni/loltools/model"
  "net/http"
)

type PlayerInfo struct {
  Id  string
  Uri string

  Summoner string
  Wins     int
  Losses   int
}

func (p *PlayerInfo) Fill(m *model.Player) {
  p.Id = m.Id()
  p.Uri = m.Uri()
  p.Summoner = m.Summoner
}

func TeamViewHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  leagueId := args["leagueId"]
  teamId := args["teamId"]

  _, userKey, err := model.GetUser(c)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  league, leagueKey, err := model.LeagueById(c, userKey, leagueId)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  team, teamKey, err := model.TeamById(c, userKey, leagueKey, teamId)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  players /*playerKeys*/, _, err := model.TeamAllPlayers(
    c, userKey, leagueKey, teamKey, model.KeysAndEntities)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  // Populate view context.
  ctx := struct {
    ctxBase
    League
    Team
    Players []*PlayerInfo
  }{}
  ctx.ctxBase.init(c)
  ctx.ctxBase.Title = fmt.Sprintf("loltools - %s - %s", league.Name, team.Name)
  ctx.League.Fill(league, leagueKey)
  ctx.Team.Fill(team, teamKey, leagueKey)

  ctx.Players = make([]*PlayerInfo, len(players))
  for i, p := range players {
    ctx.Players[i].Fill(p)
  }

  // Render
  if err := RenderTemplate(w, "leagues/teams/view.html", "base", ctx); err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
}

func ApiTeamAddPlayerHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  leagueId := r.FormValue("league")
  teamId := r.FormValue("team")
  //region := r.FormValue("region")
  //summoner := r.FormValue("summoner")

  /*user*/ _, userKey, err := model.GetUser(c)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  /*league*/ _, leagueKey, err := model.LeagueById(c, userKey, leagueId)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  /*team*/ _ /*teamKey*/, _, err = model.TeamById(c, userKey, leagueKey, teamId)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  HttpReplyError(w, r, http.StatusInternalServerError,
    errors.New("TODO: implement by-summoner name lookup"))
}
