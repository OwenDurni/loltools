package loltools

import (
  "appengine"
  "appengine/user"
  "fmt"
  "github.com/OwenDurni/loltools/view"
  "net/http"
)

func debugHandler(w http.ResponseWriter, r *http.Request) {
  appengineCtx := appengine.NewContext(r)
  user := user.Current(appengineCtx)
  fmt.Println(w, "User: ", user)
}

func init() {
  http.HandleFunc("/", debugHandler)
  http.HandleFunc("/profile/edit", view.ProfileEditHandler)
  http.HandleFunc("/profile/update", view.ProfileUpdateHandler)
}
