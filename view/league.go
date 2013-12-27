package view

import (
  "appengine"
  "github.com/OwenDurni/loltools/model"
  "html/template"
  "net/http"
)

func LeagueCreateHandler(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)

  _, _, err := model.GetUser(c)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  formContents, err := parseTemplate("template/league/create.html", nil)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  formCtx := new(formCtx).init()
  formCtx.FormId = "league-create"
  formCtx.SubmitUrl = "/api/league/create"
  formCtx.FormHTML = template.HTML(formContents)
  formHtml, err := renderForm(formCtx)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  pageCtx := new(commonCtx).init(c)
  pageCtx.Title = "Create League"
  pageCtx.ContentHTML = template.HTML(formHtml)
  pageHtml, err := parseTemplate("template/common.html", pageCtx)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  w.Write(pageHtml)
}

func ApiLeagueCreateHandler(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)
  _, leagueKey, err := model.CreateLeague(c, r.FormValue("name"))
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  HttpReplyResourceCreated(w, model.LeagueUri(leagueKey))
}
