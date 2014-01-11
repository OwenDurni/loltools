package loltools

import (
  "appengine"
  "appengine/user"
  "errors"
  "fmt"
  "github.com/OwenDurni/loltools/view"
  "github.com/OwenDurni/loltools/util/dispatch"
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

var dispatcher *dispatch.Dispatcher

func init() {
  dispatcher = new(dispatch.Dispatcher)
  
  dispatcher.Add("/", view.HomeHandler)
  dispatcher.Add("/api/leagues/create", view.ApiLeagueCreateHandler)
  dispatcher.Add("/api/profiles/set", view.ProfileSetHandler)
  dispatcher.Add("/home", view.HomeHandler)
  dispatcher.Add("/leagues/create", view.LeagueCreateHandler)
  //dispatcher.Add("/leagues", view.LeaguesHandler)
  dispatcher.Add("/profiles/edit", view.ProfileEditHandler)

  http.HandleFunc("/", dispatcher.RootHandler)
}
