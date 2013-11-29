package loltools

import (
  "github.com/OwenDurni/loltools/view"
  "net/http"
)

func init() {
  http.HandleFunc("/", view.ProfileEditHandler)
}
