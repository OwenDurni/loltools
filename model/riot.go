package model

import (
  "appengine"
  "appengine/datastore"
  "appengine/memcache"
)

const (
  RegionNA   = "na"
  RegionEUW  = "euw"
  RegionEUNE = "eune"
)

var RiotApiRateLimiter = DistributedRateLimiter{
  Name: "riot-rest-api",
  Limits: RiotDevRateLimits,
}

var RiotDevRateLimits = []RateLimit{
  RateLimit{5, 10},
  RateLimit{250, 10 * 60},
}

type RiotApiKey struct {
  Key string
}

func GetRiotApiKey(c appengine.Context) (*RiotApiKey, error) {
  var r = new(RiotApiKey)
  
  // First try memcache.
  if _, err := memcache.JSON.Get(c, "RiotApiKey/dev", r); err == nil {
    return r, nil
  }
  
  // Next try datastore.
  key := datastore.NewKey(c, "RiotApiKey", "dev", 0, nil)
  if err := datastore.Get(c, key, r); err == datastore.ErrNoSuchEntity {
    return nil, nil
  } else if err != nil {
    return nil, err
  }
  
  // Best effort put into datastore before returning.
  memcache.JSON.Set(c, &memcache.Item{Key: "RiotApiKey/dev", Object: r})
  return r, nil
}

func SetRiotApiKey(c appengine.Context, apikey string) error {
  r := new(RiotApiKey)
  r.Key = apikey
  
  key := datastore.NewKey(c, "RiotApiKey", "dev", 0, nil)
  _, err := datastore.Put(c, key, r)
  if err != nil {
    return err
  }
  // Best effort put in memcache.
  memcache.JSON.Set(c, &memcache.Item{Key: "RiotApiKey/dev", Object: r})
  return nil
}
