package view

import (
  "appengine"
  "github.com/OwenDurni/loltools/model"
  "net/http"
)

func LeagueCreateHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)

  _, _, err := model.GetUser(c)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  ctx := struct {
    ctxBase
  }{}
  ctx.ctxBase.init(c)
  ctx.ctxBase.Title = "Create League"

  if err := RenderTemplate(w, "leagues/create.html", "base", ctx); err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
}

func ApiLeagueCreateHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  _, leagueKey, err := model.CreateLeague(c, r.FormValue("name"))
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  HttpReplyResourceCreated(w, model.LeagueUri(leagueKey))
}
