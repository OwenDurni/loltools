package view

import (
  "appengine"
  "github.com/OwenDurni/loltools/model"
  "net/http"
)

func AdminIndexHandler(
  w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)

  riotApiKey, err := model.GetRiotApiKey(c)
  if HandleError(c, w, err) { return }

  ctx := struct {
    ctxBase
    RiotApiKey *model.RiotApiKey
  }{}
  ctx.ctxBase.init(c)
  ctx.ctxBase.Title = "Admin Console"
  ctx.RiotApiKey = riotApiKey

  err = RenderTemplate(w, "admin.html", "base", ctx)
  if HandleError(c, w, err) { return }
}

func ApiAdminRiotKeySetHandler(
  w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  apikey := r.FormValue("key")

  err := model.SetRiotApiKey(c, apikey)
  if ApiHandleError(c, w, err) { return }
  
  HttpReplyOkEmpty(w)
}
