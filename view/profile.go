package view

import (
  "appengine"
  "github.com/OwenDurni/loltools/model"
  "net/http"
)

func ProfileEditHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)

  user, _, err := model.GetUser(c)
  if HandleError(c, w, err) { return }

  ctx := struct {
    ctxBase
    User *model.User
  }{}
  ctx.ctxBase.init(c)
  ctx.ctxBase.Title = "Edit Profile"
  ctx.User = user

  err = RenderTemplate(w, "profiles/edit.html", "base", ctx)
  if HandleError(c, w, err) { return }
}

func ProfileSetHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  
  user, _, err := model.GetUser(c)
  if ApiHandleError(c, w, err) { return }
  
  user.Name = r.FormValue("name")
  user.SummonerName = r.FormValue("summoner")
  err = user.Save(c)
  if ApiHandleError(c, w, err) { return }
  
  HttpReplyOkEmpty(w)
}
