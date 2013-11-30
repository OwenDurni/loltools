package view

import (
  "appengine"
  "github.com/OwenDurni/loltools/model"
  "html/template"
  "net/http"
)

func ProfileEditHandler(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)

  user, err := model.GetUser(c)
  if err != nil {
    httpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  formContents, err := parseTemplate("template/profile/edit.html", user)
  if err != nil {
    httpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  formCtx := new(formCtx)
  formCtx.init()
  formCtx.FormId = "edit-profile"
  formCtx.SubmitUrl = "/profile/update"
  formCtx.FormHTML = template.HTML(formContents)
  formHtml, err := renderForm(formCtx);
  if err != nil {
    httpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  
  pageCtx := new(commonCtx)
  pageCtx.init()
  pageCtx.Title = "Edit Profile"
  pageCtx.ContentHTML = template.HTML(formHtml)
  pageHtml, err := parseTemplate("template/common.html", pageCtx)
  if err != nil {
    httpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  w.Write(pageHtml)
}

func ProfileUpdateHandler(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)
  user, err := model.GetUser(c)
  if err != nil {
    httpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  user.Name = r.FormValue("name")
  user.SummonerName = r.FormValue("summoner")
  if err := user.Save(c); err != nil {
    httpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  httpReplyOkEmpty(w)
}
