package view

import (
  "appengine"
  "github.com/OwenDurni/loltools/model"
  "net/http"
)

func EventList(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)

  _, _, err := model.GetUser(c)
  if err != nil {
    httpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  return
}
