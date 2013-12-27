package loltools

import (
  "appengine"
  "appengine/user"
  "errors"
  "fmt"
  "github.com/OwenDurni/loltools/view"
  "net/http"
)

func debugHandler(w http.ResponseWriter, r *http.Request) {
  appengineCtx := appengine.NewContext(r)
  user := user.Current(appengineCtx)
  fmt.Println(w, "User: ", user)
}

func apiNotImplemented(w http.ResponseWriter, r *http.Request) {
  err := errors.New("Not implemented")
  view.HttpReplyError(w, r, http.StatusInternalServerError, err)
  return
}

func init() {
  http.HandleFunc("/", view.HomeHandler)
  http.HandleFunc("/home", view.HomeHandler)
  http.HandleFunc("/profile/edit", view.ProfileEditHandler)
  http.HandleFunc("/league/create", view.LeagueCreateHandler)

  http.HandleFunc("/api/profile/set", view.ProfileSetHandler)
  http.HandleFunc("/api/league/create", apiNotImplemented)
}
