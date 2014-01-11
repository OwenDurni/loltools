package view

import (
  "appengine"
  "github.com/OwenDurni/loltools/model"
  "net/http"
)

func ProfileEditHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)

  user, _, err := model.GetUser(c)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  ctx := struct {
    ctxBase
    User *model.User
  }{}
  ctx.ctxBase.init(c)
  ctx.ctxBase.Title = "Edit Profile"
  ctx.User = user

  if err := RenderTemplate(w, "profiles/edit.html", "base", ctx); err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
}

func ProfileSetHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  user, _, err := model.GetUser(c)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  user.Name = r.FormValue("name")
  user.SummonerName = r.FormValue("summoner")
  if err := user.Save(c); err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  HttpReplyOkEmpty(w)
}
