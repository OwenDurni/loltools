package model

import (
  "appengine"
  "appengine/datastore"
)

const (
  RegionNA   = "na"
  RegionEUW  = "euw"
  RegionEUNE = "eune"
)

type RiotApiKey struct {
  Key    string
  Limits []RateLimit
}

func GetRiotApiKey(c appengine.Context) (*RiotApiKey, error) {
  var r = new(RiotApiKey)
  key := datastore.NewKey(c, "RiotApiKey", "dev", 0, nil)
  if err := datastore.Get(c, key, r); err == datastore.ErrNoSuchEntity {
    return nil, nil
  } else if err != nil {
    return nil, err
  }
  return r, nil
}

func SetRiotApiKey(c appengine.Context, apikey string) error {
  r := new(RiotApiKey)
  r.Key = apikey
  r.Limits = []RateLimit{
    RateLimit{10, 10},
    RateLimit{500, 10 * 60},
  }
  key := datastore.NewKey(c, "RiotApiKey", "dev", 0, nil)
  _, err := datastore.Put(c, key, r)
  return err
}
