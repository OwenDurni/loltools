package task

import (
  "appengine"
  "errors"
  "fmt"
  "github.com/OwenDurni/loltools/model"
  "github.com/OwenDurni/loltools/riot"
  "net/http"
  "strconv"
)

func FetchMatchHistoryHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  region := r.FormValue("region")
  riotSummonerId, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
  if err != nil {
    ReportPermanentError(c, w, errors.New(fmt.Sprintf("Could not parse id param: %v", err)))
    return
  }
  
  riotApiKey, err := model.GetRiotApiKey(c)
  if err != nil {
    ReportPermanentError(c, w, errors.New(fmt.Sprintf("Could not get riotApiKey: %v", err)))
    return
  }
  
  riotData, err := riot.GameStatsForPlayer(c, riotApiKey.Key, region, riotSummonerId)
  if err != nil {
    ReportPermanentError(
      c, w, errors.New(fmt.Sprintf("Could not riot.GameStatsForPlayer: %v", err)))
    return
  }
  
  // TODO(durni): Cache result in datastore.
  
  fmt.Fprintf(w, "%+v", riotData)
}