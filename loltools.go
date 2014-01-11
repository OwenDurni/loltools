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

func debugHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  appengineCtx := appengine.NewContext(r)
  user := user.Current(appengineCtx)
  fmt.Fprintf(w, "<html>\n")
  fmt.Fprintf(w, "<body>\n")
  fmt.Fprintf(w, "<h3>Debug Info</h3>\n")
  fmt.Fprintf(w, "<pre>\n")
  fmt.Fprintf(w, "User: %v\n", user)
  fmt.Fprintf(w, "Dispatch Args: %v\n", args)
  fmt.Fprintf(w, "Request: %+v\n", *r)
  fmt.Fprintf(w, "</pre>\n")
  fmt.Fprintf(w, "</body>\n")
  fmt.Fprintf(w, "</html>\n")
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
  dispatcher.Add("/leagues/<leagueKey>", debugHandler)
  dispatcher.Add("/profiles/edit", view.ProfileEditHandler)

  http.HandleFunc("/", dispatcher.RootHandler)
}
