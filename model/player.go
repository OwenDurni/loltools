package model

import (
  "appengine"
  "appengine/datastore"
  "errors"
  "fmt"
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

  // The last time we refreshed data for this player.
  LastUpdated time.Time
}

func (p *Player) Id() string {
  return fmt.Sprintf("%s-%d", p.Region, p.RiotId)
}
func (p *Player) Uri() string {
  return fmt.Sprintf("/players/%s", p.Id())
}

func GetPlayerByRiotId(
  c appengine.Context,
  region string,
  riotId int64) (*Player, *datastore.Key, error) {
  var player *Player = new(Player)
  playerKey := datastore.NewKey(
    c, "Player", fmt.Sprintf("%s-%d", region, riotId), 0, nil)

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
  return nil, nil, errors.New("not implemented")
}
