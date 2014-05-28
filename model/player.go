package model

import (
  "appengine"
  "appengine/datastore"
  "appengine/memcache"
  "errors"
  "fmt"
  "github.com/OwenDurni/loltools/riot"
  "github.com/OwenDurni/loltools/util/errwrap"
  "strconv"
  "strings"
  "time"
)

// Caches player data within a region.
type PlayerCache struct {
  Region     string
  c          appengine.Context
  byId       map[int64]*Player
  bySummoner map[string]*Player
}

func NewPlayerCache(c appengine.Context, region string) *PlayerCache {
  cache := new(PlayerCache)
  cache.Region = region
  cache.c = c
  cache.byId = make(map[int64]*Player)
  cache.bySummoner = make(map[string]*Player)
  return cache
}
func (cache *PlayerCache) ById(riotId int64) (*Player, error) {
  if p, exists := cache.byId[riotId]; exists {
    return p, nil
  }
  p, _, err := GetPlayerByRiotIdOrStub(cache.c, cache.Region, riotId)
  if err == nil {
    cache.Add(p)
  }
  return p, err
}
func (cache *PlayerCache) BySummoner(summoner string) (*Player, error) {
  if p, exists := cache.bySummoner[summoner]; exists {
    return p, nil
  }
  p, _, err := GetOrCreatePlayerBySummoner(cache.c, cache.Region, summoner)
  if err == nil {
    cache.Add(p)
  }
  return p, err
}
func (cache *PlayerCache) Add(p *Player) {
  cache.byId[p.RiotId] = p
  cache.bySummoner[p.Summoner] = p
}

// ("%s-%s", Region, RiotId) is the key for a player.
type Player struct {
  // The current summoner name for this player.
  Summoner string

  // This is the region for the given player.
  Region string

  // Summoner id as per Riot API.
  RiotId int64

  // This player's in game level.
  Level int

  LastUpdated time.Time
}

func (p *Player) Id() string {
  return MakePlayerId(p.Region, p.RiotId)
}
func (p *Player) Uri() string {
  return fmt.Sprintf("/players/%s", p.Id())
}

func MakePlayerId(region string, riotSummonerId int64) string {
  return fmt.Sprintf("%s-%d", region, riotSummonerId)
}
func KeyForPlayer(c appengine.Context, region string, riotSummonerId int64) *datastore.Key {
  return datastore.NewKey(c, "Player", MakePlayerId(region, riotSummonerId), 0, nil)
}

func SplitPlayerKey(key *datastore.Key) (string, int64, error) {
  parts := strings.Split(key.StringID(), "-")
  if len(parts) != 2 {
    return "", 0, errors.New(fmt.Sprintf("Cannot split malformed PlayerKey: %s",
      key.StringID()))
  }
  region := parts[0]
  id, err := strconv.ParseInt(parts[1], 10, 64)
  if err != nil {
    return "", 0, err
  }
  return region, id, nil
}

// sort.Interface for []*Player
type PlayersBySummoner []*Player

func (a PlayersBySummoner) Len() int {
  return len(a)
}
func (a PlayersBySummoner) Less(i, j int) bool {
  return strings.ToLower(a[i].Summoner) < strings.ToLower(a[j].Summoner)
}
func (a PlayersBySummoner) Swap(i, j int) {
  a[i], a[j] = a[j], a[i]
}

func GetPlayerByRiotIdOrStub(
  c appengine.Context,
  region string,
  riotId int64) (*Player, *datastore.Key, error) {
  player := new(Player)
  playerKey := KeyForPlayer(c, region, riotId)

  // Try memcache first.
  mkey := fmt.Sprintf("Player/%s-%d", region, riotId)
  if _, err := memcache.JSON.Get(c, mkey, player); err == nil {
    return player, playerKey, nil
  }

  err := datastore.Get(c, playerKey, player)
  if err == datastore.ErrNoSuchEntity {
    // Return a stub.
    player.Summoner = fmt.Sprintf("<%s-%d>", region, riotId)
    player.Region = region
    player.RiotId = riotId
    return player, playerKey, nil
  } else if err != nil {
    return nil, nil, err
  }
  return player, playerKey, nil
}

func GetOrCreatePlayerByRiotId(
  c appengine.Context,
  region string,
  riotId int64) (*Player, *datastore.Key, error) {
  player := new(Player)

  // Try memcache first.
  mkey := fmt.Sprintf("Player/%s-%d", region, riotId)
  if _, err := memcache.JSON.Get(c, mkey, player); err == nil {
    return player, KeyForPlayer(c, region, player.RiotId), nil
  }

  playerKey := KeyForPlayer(c, region, riotId)

  for attempt := 0; attempt < 3; attempt++ {
    err := datastore.Get(c, playerKey, player)
    if err == datastore.ErrNoSuchEntity {
      if err := RiotApiRateLimiter.Consume(c, 1); err != nil {
        return nil, nil, errwrap.Wrap(err)
      }
      riotApiKey, err := GetRiotApiKey(c)
      if err != nil {
        return nil, nil, errwrap.Wrap(err)
      }
      riotSummoners, err := riot.SummonersById(c, riotApiKey.Key, region, riotId)
      if err != nil {
        return nil, nil, errwrap.Wrap(err)
      }
      riotSummoner := riotSummoners[0]
      if riotSummoner == nil {
        return nil, nil, errwrap.Wrap(errors.New(fmt.Sprintf("Summoner id not found: %d", riotId)))
      }
      player.Summoner = riotSummoner.Name
      player.Region = region
      player.RiotId = riotSummoner.Id
      player.Level = riotSummoner.SummonerLevel

      if _, err = datastore.Put(c, playerKey, player); err != nil {
        return nil, nil, errwrap.Wrap(err)
      }
      continue
    }
    if err != nil {
      return nil, nil, err
    }
  }

  // Best effort put into memcache.
  memcache.JSON.Set(c, &memcache.Item{Key: mkey, Object: player})
  return player, playerKey, nil
}

func GetOrCreatePlayerBySummoner(
  c appengine.Context,
  region string,
  summoner string) (*Player, *datastore.Key, error) {
  if err := RiotApiRateLimiter.Consume(c, 1); err != nil {
    return nil, nil, errwrap.Wrap(err)
  }
  riotApiKey, err := GetRiotApiKey(c)
  if err != nil {
    return nil, nil, errwrap.Wrap(err)
  }

  riotSummoner, err := riot.SummonerByName(c, riotApiKey.Key, region, summoner)
  if err != nil {
    return nil, nil, errwrap.Wrap(err)
  }

  player := new(Player)
  player.Summoner = riotSummoner.Name
  player.Region = region
  player.RiotId = riotSummoner.Id
  player.Level = riotSummoner.SummonerLevel

  playerKey := KeyForPlayer(c, region, player.RiotId)

  if _, err = datastore.Put(c, playerKey, player); err != nil {
    return nil, nil, errwrap.Wrap(err)
  }

  return player, playerKey, nil
}
