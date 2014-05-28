package view

import (
  "appengine"
  "appengine/datastore"
  "github.com/OwenDurni/loltools/model"
  "net/http"
)

func AdminIndexHandler(
  w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)

  riotApiKey, err := model.GetRiotApiKey(c)
  if HandleError(c, w, err) {
    return
  }

  q := datastore.NewQuery("PlayerGameStats").
    Filter("Saved =", false).
    Filter("NotAvailable =", false).
    KeysOnly()
  gameStatsKeys, err := q.GetAll(c, nil)
  if HandleError(c, w, err) {
    return
  }

  ctx := struct {
    ctxBase
    RiotApiKey            *model.RiotApiKey
    GameStatsBacklogCount int
    RiotRateLimit         string
  }{}
  ctx.ctxBase.init(c)
  ctx.ctxBase.Title = "Admin Console"
  ctx.RiotApiKey = riotApiKey
  ctx.GameStatsBacklogCount = len(gameStatsKeys)
  ctx.RiotRateLimit = model.RiotApiRateLimiter.DebugStr(c)

  err = RenderTemplate(w, "admin.html", "base", ctx)
  if HandleError(c, w, err) {
    return
  }
}

func ApiAdminRiotKeySetHandler(
  w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  apikey := r.FormValue("key")

  err := model.SetRiotApiKey(c, apikey)
  if ApiHandleError(c, w, err) {
    return
  }

  HttpReplyOkEmpty(w)
}
