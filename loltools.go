package loltools

import (
  "fmt"
  "loltools/view"
  "net/http"
)

func init() {
  http.HandleFunc("/", view.ProfileEditHandler)
}
