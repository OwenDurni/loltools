package view

import (
  "appengine"
  "github.com/OwenDurni/loltools/model"
  "html/template"
  "net/http"
)

func ProfileEditHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)

  user, _, err := model.GetUser(c)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  formContents, err := parseTemplate("template/profiles/edit.html", user)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  formCtx := new(formCtx)
  formCtx.init()
  formCtx.FormId = "edit-profile"
  formCtx.SubmitUrl = "/api/profiles/set"
  formCtx.FormHTML = template.HTML(formContents)
  formHtml, err := renderForm(formCtx)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  pageCtx := new(commonCtx).init(c)
  pageCtx.Title = "Edit Profile"
  pageCtx.ContentHTML = template.HTML(formHtml)
  pageHtml, err := parseTemplate("template/common.html", pageCtx)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  w.Write(pageHtml)
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
