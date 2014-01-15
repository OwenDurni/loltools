package model

import (
  "appengine"
  "appengine/datastore"
  "fmt"
  "github.com/OwenDurni/loltools/riot"
  "github.com/OwenDurni/loltools/util/errwrap"
  "strings"
  "time"
)

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

  // The last time we refreshed data for this player (UTC).
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

func GetOrCreatePlayerByRiotId(
  c appengine.Context,
  region string,
  riotId int64) (*Player, *datastore.Key, error) {
  var player *Player = new(Player)
  playerKey := KeyForPlayer(c, region, riotId)

  err := datastore.RunInTransaction(c, func(c appengine.Context) error {
    err := datastore.Get(c, playerKey, player)
    if err == datastore.ErrNoSuchEntity {
      player.Region = region
      player.RiotId = riotId
      playerKey, err = datastore.Put(c, playerKey, player)
    }
    return err
  }, nil)
  if err != nil {
    return nil, nil, err
  }
  return player, playerKey, nil
}

func GetOrCreatePlayerBySummoner(
  c appengine.Context,
  region string,
  summoner string) (*Player, *datastore.Key, error) {
  // Do a first pass check to avoid hitting Riot API if possible.
  q := datastore.NewQuery("Player").
    Filter("Region =", region).
    Filter("Summoner =", summoner).
    Limit(1)
  var players []*Player
  playerKeys, err := q.GetAll(c, &players)
  if err != nil {
    return nil, nil, errwrap.Wrap(err)
  }
  if len(players) > 0 {
    return players[0], playerKeys[0], nil
  }

  // Otherwise we need to fetch some data from Riot.
  if err := RiotApiRateLimiter.TryConsume(c, 1); err != nil {
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
  player.LastUpdated = time.Now().UTC()

  playerKey := KeyForPlayer(c, region, player.RiotId)

  if _, err = datastore.Put(c, playerKey, player); err != nil {
    return nil, nil, errwrap.Wrap(err)
  }
  return player, playerKey, nil
}
