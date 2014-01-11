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

  if err := RenderTemplate(w, "home.html", "base", ctx); err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
}
