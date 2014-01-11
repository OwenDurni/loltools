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
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  ctx := struct {
    ctxBase
    RiotApiKey *model.RiotApiKey
  }{}
  ctx.ctxBase.init(c)
  ctx.ctxBase.Title = "Admin Console"
  ctx.RiotApiKey = riotApiKey

  if err := RenderTemplate(w, "admin.html", "base", ctx); err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
}

func ApiAdminRiotKeySetHandler(
    w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  apikey := r.FormValue("key")
  
  if err := model.SetRiotApiKey(c, apikey); err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  HttpReplyOkEmpty(w)
}