package riot

import (
  "appengine"
  "appengine/urlfetch"
  "errors"
  "fmt"
  "io/ioutil"
  "net/url"
)

const (
  baseUrl = "https://prod.api.pvp.net"
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

func Fetch(c appengine.Context, loc string) ([]byte, error) {
  client := urlfetch.Client(c)
  resp, err := client.Get(loc)
  if err != nil {
    return nil, err
  }
  if resp.StatusCode < 200 || resp.StatusCode > 299 {
    return nil, errors.New(
      fmt.Sprintf("Fetch failed with status %d: %s", resp.StatusCode, loc))
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return nil, err
  }
  return body, nil
}
