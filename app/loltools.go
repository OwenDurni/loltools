package loltools

import (
  "appengine"
  "appengine/user"
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

var dispatcher *dispatch.Dispatcher

func init() {
  dispatcher = new(dispatch.Dispatcher)

  dispatcher.Add("/", view.HomeHandler)
  dispatcher.Add("/admin", view.AdminIndexHandler)
  dispatcher.Add("/api/admin/riotapikey/set", view.ApiAdminRiotKeySetHandler)
  dispatcher.Add("/api/groups/add-user", view.ApiGroupAddUserHandler)
  dispatcher.Add("/api/groups/create", view.ApiGroupCreateHandler)
  dispatcher.Add("/api/groups/del-user", view.ApiGroupDelUserHandler)
  dispatcher.Add("/api/groups/join", view.ApiGroupJoinHandler)
  dispatcher.Add("/api/leagues/add-team", view.ApiLeagueAddTeamHandler)
  dispatcher.Add("/api/leagues/create", view.ApiLeagueCreateHandler)
  dispatcher.Add("/api/leagues/group-acl-grant", view.ApiLeagueGroupAclGrantHandler)
  dispatcher.Add("/api/leagues/group-acl-revoke", view.ApiLeagueGroupAclRevokeHandler)
  dispatcher.Add("/api/leagues/teams/add-player", view.ApiTeamAddPlayerHandler)
  dispatcher.Add("/api/leagues/teams/del-player", view.ApiTeamDelPlayerHandler)
  dispatcher.Add("/api/matches/create", view.ApiMatchCreateHandler)
  dispatcher.Add("/api/user/add-summoner", view.ApiUserAddSummoner)
  dispatcher.Add("/api/user/set-primary-summoner", view.ApiUserSetPrimarySummoner)
  dispatcher.Add("/api/user/verify-summoner", view.ApiUserVerifySummoner)
  dispatcher.Add("/debug", debugHandler)
  dispatcher.Add("/home", view.HomeHandler)
  dispatcher.Add("/games/<gameId>", view.GameViewHandler)
  dispatcher.Add("/groups", view.GroupIndexHandler)
  dispatcher.Add("/groups/<groupId>", view.GroupViewHandler)
  dispatcher.Add("/leagues", view.LeagueIndexHandler)
  dispatcher.Add("/leagues/<leagueId>", view.LeagueViewHandler)
  dispatcher.Add("/leagues/<leagueId>/games/<gameId>", view.LeagueGameViewHandler)
  dispatcher.Add("/leagues/<leagueId>/matches/create", view.MatchCreateHandler)
  dispatcher.Add("/leagues/<leagueId>/teams/<teamId>", view.TeamViewHandler)
  dispatcher.Add("/leagues/<leagueId>/teams/<teamId>/history", view.TeamGameHistory)
  dispatcher.Add("/task/cron/all-match-sync", task.AllMatchSync)
  dispatcher.Add("/task/cron/all-team-histories", task.AllTeamHistories)
  dispatcher.Add("/task/cron/get-missing-game-stats", task.MissingGameStats)
  dispatcher.Add("/task/riot/get/team/history", task.FetchTeamMatchHistoryHandler)
  dispatcher.Add("/task/match/sync", task.MatchSync)
  dispatcher.Add("/settings", view.SettingsIndexHandler)

  http.HandleFunc("/", dispatcher.RootHandler)

  view.SetTemplateRoot("template/")
  view.AddTemplate("admin.html",
    "form.html", "base.html")
  view.AddTemplate("httperror.html",
    "base.html")
  view.AddTemplate("home.html",
    "base.html")
  view.AddTemplate("games/index.html",
    "games/gamelong.html", "games/champsmall.html", "games/itemsmall.html",
    "games/summonersmall.html", "base.html")
  view.AddTemplate("groups/index.html",
    "form.html", "base.html")
  view.AddTemplate("groups/join.html",
    "form.html", "base.html")
  view.AddTemplate("groups/view.html",
    "form.html", "base.html")
  view.AddTemplate("leagues/create.html",
    "form.html", "base.html")
  view.AddTemplate("leagues/index.html",
    "form.html", "base.html")
  view.AddTemplate("leagues/games/index.html",
    "games/gamelong.html", "games/champsmall.html", "games/itemsmall.html",
    "games/summonersmall.html", "form.html", "base.html")
  view.AddTemplate("leagues/matches/create.html",
    "form.html", "base.html")
  view.AddTemplate("leagues/teams/history.html",
    "games/gamelong.html", "games/champsmall.html", "games/itemsmall.html",
    "games/summonersmall.html", "base.html")
  view.AddTemplate("leagues/teams/view.html",
    "games/gameshort.html", "games/champsmall.html", "form.html", "base.html")
  view.AddTemplate("leagues/view.html",
    "form.html", "types.html", "base.html")
  view.AddTemplate("settings/index.html",
    "common/region_dropdown.html", "form.html", "base.html")
}
