package view

import (
  "appengine"
  "net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)

  ctx := struct {
    ctxBase
  }{}
  ctx.ctxBase.init(c)

  err := RenderTemplate(w, "home.html", "base", ctx)
  if HandleError(c, w, err) {
    return
  }
}
