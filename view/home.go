package view

import (
  "appengine"
  "github.com/OwenDurni/loltools/model"
  "html/template"
  "net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)

  _, _, err := model.GetUser(c)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  homeHtml, err := parseTemplate("template/home.html", nil)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  pageCtx := new(commonCtx).init(c)
  pageCtx.Title = "Home"
  pageCtx.ContentHTML = template.HTML(homeHtml)
  pageHtml, err := parseTemplate("template/common.html", pageCtx)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  w.Write(pageHtml)
}
