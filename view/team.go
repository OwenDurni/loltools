package view

import (
  "appengine"
  "fmt"
  "github.com/OwenDurni/loltools/model"
  "github.com/OwenDurni/loltools/util/errwrap"
  "net/http"
  "sort"
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
    HttpReplyError(w, r, http.StatusInternalServerError, errwrap.Wrap(err))
    return
  }

  league, leagueKey, err := model.LeagueById(c, userKey, leagueId)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, errwrap.Wrap(err))
    return
  }

  team, teamKey, err := model.TeamById(c, userKey, leagueKey, teamId)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, errwrap.Wrap(err))
    return
  }

  players /*playerKeys*/, _, err := model.TeamAllPlayers(
    c, userKey, leagueKey, teamKey, model.KeysAndEntities)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, errwrap.Wrap(err))
    return
  }

  // Sort players by summoner.
  sort.Sort(model.PlayersBySummoner(players))

  // Get recent match history.
  gameInfos, errors := model.TeamRecentGameInfo(c, 5, leagueKey, teamKey)
  
  // Populate view context.
  ctx := struct {
    ctxBase
    League
    Team
    RecentGames []*model.GameInfo
    Players     []*PlayerInfo
  }{}
  ctx.ctxBase.init(c)
  ctx.ctxBase.Title = fmt.Sprintf("loltools - %s - %s", league.Name, team.Name)
  ctx.ctxBase.Errors = errors
  ctx.League.Fill(league, leagueKey)
  ctx.Team.Fill(team, teamKey, leagueKey)
  ctx.RecentGames = gameInfos

  ctx.Players = make([]*PlayerInfo, len(players))
  for i, p := range players {
    ctx.Players[i] = new(PlayerInfo)
    ctx.Players[i].Fill(p)
  }

  // Render
  if err := RenderTemplate(w, "leagues/teams/view.html", "base", ctx); err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, errwrap.Wrap(err))
    return
  }
}

func ApiTeamAddPlayerHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  leagueId := r.FormValue("league")
  teamId := r.FormValue("team")
  region := r.FormValue("region")
  summoner := r.FormValue("summoner")

  /*user*/ _, userKey, err := model.GetUser(c)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, errwrap.Wrap(err))
    return
  }

  _, leagueKey, err := model.LeagueById(c, userKey, leagueId)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, errwrap.Wrap(err))
    return
  }

  _, teamKey, err := model.TeamById(c, userKey, leagueKey, teamId)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, errwrap.Wrap(err))
    return
  }

  _, playerKey, err := model.GetOrCreatePlayerBySummoner(
    c, region, summoner)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, errwrap.Wrap(err))
    return
  }

  err = model.TeamAddPlayer(c, userKey, leagueKey, teamKey, playerKey)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, errwrap.Wrap(err))
    return
  }
  HttpReplyOkEmpty(w)
}
