package loltools

import (
  "appengine"
  "appengine/user"
  "errors"
  "fmt"
  "github.com/OwenDurni/loltools/model"
  "github.com/OwenDurni/loltools/task"
  "github.com/OwenDurni/loltools/util/dispatch"
  "github.com/OwenDurni/loltools/view"
  "net/http"
)

func debugHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  appengineUser := user.Current(c)
  user, userKey, userErr := model.GetUser(c)
  fmt.Fprintf(w, "<html>\n")
  fmt.Fprintf(w, "<body>\n")
  fmt.Fprintf(w, "<h3>Debug Info</h3>\n")
  fmt.Fprintf(w, "<pre>\n")
  fmt.Fprintf(w, "Appengine User: %v\n", appengineUser)
  if userErr != nil {
    fmt.Fprintf(w, "User (error): %v\n", userErr.Error())
  } else {
    fmt.Fprintf(w, "User: %+v\n", user)
    fmt.Fprintf(w, "User Key: %v\n", userKey.Encode())
  }
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
  dispatcher.Add("/admin", view.AdminIndexHandler)
  dispatcher.Add("/api/admin/riotapikey/set", view.ApiAdminRiotKeySetHandler)
  dispatcher.Add("/api/groups/create", view.ApiGroupCreateHandler)
  dispatcher.Add("/api/leagues/add-team", view.ApiLeagueAddTeamHandler)
  dispatcher.Add("/api/leagues/create", view.ApiLeagueCreateHandler)
  dispatcher.Add("/api/leagues/teams/add-player", view.ApiTeamAddPlayerHandler)
  dispatcher.Add("/api/leagues/teams/del-player", view.ApiTeamDelPlayerHandler)
  dispatcher.Add("/api/profiles/set", view.ProfileSetHandler)
  dispatcher.Add("/debug", debugHandler)
  dispatcher.Add("/home", view.HomeHandler)
  dispatcher.Add("/groups", view.GroupIndexHandler)
  dispatcher.Add("/leagues", view.LeagueIndexHandler)
  dispatcher.Add("/leagues/<leagueId>", view.LeagueViewHandler)
  dispatcher.Add("/leagues/<leagueId>/teams/<teamId>", view.TeamViewHandler)
  dispatcher.Add("/task/riot/get/team/history", task.FetchTeamMatchHistoryHandler)
  dispatcher.Add("/profiles/edit", view.ProfileEditHandler)

  http.HandleFunc("/", dispatcher.RootHandler)

  view.AddTemplate("admin.html",
                   "form.html", "base.html")
  view.AddTemplate("home.html",
                   "base.html")
  view.AddTemplate("groups/index.html",
                   "form.html", "base.html")
  view.AddTemplate("leagues/create.html",
                   "form.html", "base.html")
  view.AddTemplate("leagues/index.html",
                   "form.html", "base.html")
  view.AddTemplate("leagues/teams/view.html",
                   "games/gameshort.html", "games/champsmall.html", "form.html", "base.html")
  view.AddTemplate("leagues/view.html",
                   "form.html", "base.html")
  view.AddTemplate("profiles/edit.html",
                   "form.html", "base.html")
}
