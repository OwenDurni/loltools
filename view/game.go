package view

import (
  "appengine"
  "fmt"
  "github.com/OwenDurni/loltools/model"
  "net/http"
)

func GameViewHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  leagueId := args["leagueId"]
  gameId := args["gameId"]
  
  _, _, err := model.GetUser(c)
  if HandleError(c, w, err) {
    return
  }
  //userAcls := model.NewRequestorAclCache(userKey)
  
  league, leagueKey, err := model.LeagueById(c, leagueId)
  if HandleError(c, w, err) {
    return
  }
  
  game, _, err := model.GameById(c, gameId)
  playerCache := model.NewPlayerCache(c, league.Region)
  gameInfo, errs := model.GetGameInfo(c, playerCache, game)
  if HandleError(c, w, errs...) {
    return
  }
  
  ctx := struct {
    ctxBase
    League
    GameInfo *model.GameInfo
  }{}
  ctx.ctxBase.init(c)
  ctx.ctxBase.Title = fmt.Sprintf("loltools - %s - %s", league.Name, gameId)
  ctx.League.Fill(league, leagueKey)
  ctx.GameInfo = gameInfo
  
  err = RenderTemplate(w, "leagues/games/index.html", "base", ctx)
  if HandleError(c, w, err) {
    return
  }
}