package view

import (
  "appengine"
  "github.com/OwenDurni/loltools/model"
  "net/http"
)

func SettingsIndexHandler(
  w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)

  _, _, err := model.GetUser(c)
  if HandleError(c, w, err) {
    return
  }
}
