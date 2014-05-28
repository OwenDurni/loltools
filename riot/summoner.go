package riot

import (
  "appengine"
  "encoding/json"
  "errors"
  "fmt"
  "net/url"
  "strconv"
  "strings"
)

// v1.4: http://developer.riotgames.com/api/methods#!/620/1931
type SummonerDto struct {
  Id            int64    `json:"id"`
  Name          string   `json:"name"`
  ProfileIconId int      `json:"profileIconId"`
  RevisionDate  RiotTime `json:"revisionDate"`
  SummonerLevel int      `json:"summonerLevel"`
}

func min(x int, y int) int {
  if x < y {
    return x
  }
  return y
}

func CanonicalizeSummoner(name string) string {
  // lower case all spaces removed.
  return strings.Replace(strings.ToLower(name), " ", "", -1)
}

func SummonerByName(
  c appengine.Context,
  riotApiKey string,
  region string,
  name string) (*SummonerDto, error) {
  name = CanonicalizeSummoner(name)
    
  loc := ComposeUrl(
    riotApiKey,
    fmt.Sprintf("/api/lol/%s/v1.4/summoner/by-name/%s", region, name),
    &url.Values{})
  jsonData, err := Fetch(c, loc)
  data := make(map[string]*SummonerDto)
  json.Unmarshal(jsonData, &data)
  if err != nil {
    return nil, err
  }

  // jsonData is a map from canonical summoner to json SummonerDto
  if dto, ok := data[name]; ok {
    return dto, nil
  } else {
    return nil, errors.New(fmt.Sprintf("Summoner does not exist: %s", name))
  }
}

// Note that if the summoner id is not found nil gets populated into the output slice.
func SummonersById(
  c appengine.Context,
  riotApiKey string,
  region string,
  ids ...int64) ([]*SummonerDto, error) {
  ret := make([]*SummonerDto, len(ids))
  
  // API supports up to 40 summoners at a time.
  begin := 0
  end := min(len(ret), 40)
  
  idStrings := make([]string, len(ids))
  for i := 0; i < len(ids); i++ {
    idStrings[i] = strconv.FormatInt(ids[i], 10)
  }
  
  for ; ; begin, end = begin+40, min(len(ret), end+40) {
    batch := idStrings[begin:end]
    
    loc := ComposeUrl(
      riotApiKey,
      fmt.Sprintf("/api/lol/%s/v1.4/summoner/%s", region, strings.Join(batch, ",")),
      &url.Values{})
    jsonData, err := Fetch(c, loc)
    data := make(map[string]*SummonerDto)
    err = json.Unmarshal(jsonData, &data)
    if err != nil {
      return nil, err
    }
    
    for i := begin; i < end; i++ {
      if dto, ok := data[idStrings[i]]; ok {
        ret[i] = dto
      }
    }
    if end == len(ret) {
      break;
    }
  }
  return ret, nil
}
