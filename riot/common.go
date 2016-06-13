package riot

import (
  "fmt"
  "net/url"
  "strings"
)

const (
  BlueTeamId   = 100
  PurpleTeamId = 200
)

const (
  baseUrl = "https://na.api.pvp.net"
)

func BaseUrl() (url *url.URL) {
  url, err := url.Parse(baseUrl)
  if err != nil {
    panic(err)
  }
  return
}

func ComposeUrl(riotApiKey string, path string, args *url.Values) string {
  u := BaseUrl()
  u.Path += path
  args.Add("api_key", riotApiKey)
  u.RawQuery = args.Encode()
  return u.String()
}

func stripArgs(loc string) string {
  return strings.Split(loc, "?")[0]
}

type ErrRiotRestApi struct {
  Url            string
  HttpStatusCode int
}

func NewErrRiotRestApi(loc string, httpStatusCode int) ErrRiotRestApi {
  var e ErrRiotRestApi
  e.Url = stripArgs(loc)
  e.HttpStatusCode = httpStatusCode
  return e
}
func (e ErrRiotRestApi) Error() string {
  return fmt.Sprintf("%s returned ErrRiotRestApi(%d)", e.Url, e.HttpStatusCode)
}

