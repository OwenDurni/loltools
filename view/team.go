package view

import (
  "appengine"
  "fmt"
  "github.com/OwenDurni/loltools/model"
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

  user, userKey, err := model.GetUser(c)
  if HandleError(c, w, err) {
    return
  }
  userAcls := model.NewRequestorAclCache(userKey)

  league, leagueKey, err := model.LeagueById(c, leagueId)
  if HandleError(c, w, err) {
    return
  }

  team, teamKey, err := model.TeamById(c, userAcls, league, leagueKey, teamId)
  if HandleError(c, w, err) {
    return
  }

  players, _, err := model.TeamAllPlayers(
    c, userAcls, league, leagueKey, teamKey, model.KeysAndEntities)
  if HandleError(c, w, err) {
    return
  }

  playerCache := model.NewPlayerCache(c, league.Region)
  for _, p := range players {
    playerCache.Add(p)
  }

  // Sort players by summoner.
  sort.Sort(model.PlayersBySummoner(players))

  // Get recent match history.
  gameInfos, errors := model.TeamRecentGameInfo(c, userAcls, 5, playerCache, league, leagueKey, teamKey)

  // Populate view context.
  ctx := struct {
    ctxBase
    League
    Team
    RecentGames []*model.GameInfo
    Players     []*PlayerInfo
  }{}
  ctx.ctxBase.init(c, user)
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
  err = RenderTemplate(w, "leagues/teams/view.html", "base", ctx)
  if HandleError(c, w, err) {
    return
  }
}

func TeamGameHistory(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  leagueId := args["leagueId"]
  teamId := args["teamId"]

  user, userKey, err := model.GetUser(c)
  if HandleError(c, w, err) {
    return
  }
  userAcls := model.NewRequestorAclCache(userKey)

  league, leagueKey, err := model.LeagueById(c, leagueId)
  if HandleError(c, w, err) {
    return
  }

  team, teamKey, err := model.TeamById(c, userAcls, league, leagueKey, teamId)
  if HandleError(c, w, err) {
    return
  }

  // Get "all" match history.
  // TODO(durni): This should likely be paginated or use infinite scroll. Currently
  // just shows 100 games.
  playerCache := model.NewPlayerCache(c, league.Region)
  gameInfos, errors := model.TeamRecentGameInfo(
    c, userAcls, 100, playerCache, league, leagueKey, teamKey)

  // Populate view context.
  ctx := struct {
    ctxBase
    League
    Team
    RecentGames []*model.GameInfo
  }{}
  ctx.ctxBase.init(c, user)
  ctx.ctxBase.Title = fmt.Sprintf("loltools - %s - %s", league.Name, team.Name)
  ctx.ctxBase.Errors = errors
  ctx.League.Fill(league, leagueKey)
  ctx.Team.Fill(team, teamKey, leagueKey)
  ctx.RecentGames = gameInfos

  // Render
  err = RenderTemplate(w, "leagues/teams/history.html", "base", ctx)
  if HandleError(c, w, err) {
    return
  }
}

func ApiTeamAddPlayerHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  leagueId := r.FormValue("league")
  teamId := r.FormValue("team")
  region := r.FormValue("region")
  summoner := r.FormValue("summoner")

  _, userKey, err := model.GetUser(c)
  if ApiHandleError(c, w, err) {
    return
  }
  userAcls := model.NewRequestorAclCache(userKey)

  league, leagueKey, err := model.LeagueById(c, leagueId)
  if ApiHandleError(c, w, err) {
    return
  }

  _, teamKey, err := model.TeamById(c, userAcls, league, leagueKey, teamId)
  if ApiHandleError(c, w, err) {
    return
  }

  _, playerKey, err := model.GetOrCreatePlayerBySummoner(c, region, summoner)
  if ApiHandleError(c, w, err) {
    return
  }

  err = model.TeamAddPlayer(c, userAcls, league, leagueKey, teamKey, playerKey)
  if ApiHandleError(c, w, err) {
    return
  }

  HttpReplyOkEmpty(w)
}

func ApiTeamDelPlayerHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  leagueId := r.FormValue("league")
  teamId := r.FormValue("team")
  region := r.FormValue("region")
  summoner := r.FormValue("summoner")

  _, userKey, err := model.GetUser(c)
  if ApiHandleError(c, w, err) {
    return
  }
  userAcls := model.NewRequestorAclCache(userKey)

  league, leagueKey, err := model.LeagueById(c, leagueId)
  if ApiHandleError(c, w, err) {
    return
  }

  _, teamKey, err := model.TeamById(c, userAcls, league, leagueKey, teamId)
  if ApiHandleError(c, w, err) {
    return
  }

  _, playerKey, err := model.GetOrCreatePlayerBySummoner(c, region, summoner)
  if ApiHandleError(c, w, err) {
    return
  }

  err = model.TeamDelPlayer(c, userAcls, league, leagueKey, teamKey, playerKey)
  if ApiHandleError(c, w, err) {
    return
  }

  HttpReplyOkEmpty(w)
}
