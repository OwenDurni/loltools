package view

import (
  "appengine"
  "html/template"
  "net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)

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
