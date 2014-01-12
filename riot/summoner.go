package riot

import (
  "appengine"
  "encoding/json"
  "fmt"
  "net/url"
)

// v1.2: http://developer.riotgames.com/api/methods#!/334/1132
type SummonerDto struct {
  Id            int64    `json:"id"`
  Name          string   `json:"name"`
  ProfileIconId int      `json:"profileIconId"`
  RevisionDate  RiotTime `json:"revisionDate"`
  SummonerLevel int      `json:"summonerLevel"`
}

func SummonerByName(
  c appengine.Context,
  riotApiKey string,
  region string,
  name string) (*SummonerDto, error) {
  loc := ComposeUrl(
    riotApiKey,
    fmt.Sprintf("/api/lol/%s/v1.2/summoner/by-name/%s", region, name),
    &url.Values{})
  jsonData, err := Fetch(c, loc)
  if err != nil {
    return nil, err
  }

  s := new(SummonerDto)
  err = json.Unmarshal(jsonData, s)
  return s, err
}
