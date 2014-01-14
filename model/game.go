package model

import (
  "appengine"
  "appengine/datastore"
  "fmt"
  "github.com/OwenDurni/loltools/riot"
)

type Game struct {
  Region     string
  RiotGameId int64
}

func KeyForGame(c appengine.Context, region string, riotGameId int64) *datastore.Key {
  return datastore.NewKey(c, "Game", fmt.Sprintf("%s-%d", region, riotGameId), 0, nil)
}

// The tuple (GameKey, PlayerKey) are unique per entity.
type PlayerGameStats struct {
  GameKey   *datastore.Key
  PlayerKey *datastore.Key
  
  // This may be nil if we tried to lookup player stats for this game after the game
  // was no longer in the recent history for this player.
  RawStats riot.GameDto
}

func GetOrCreateGame(
  c appengine.Context, region string, riotGameId int64) (*Game, *datastore.Key, error) {
  game := new(Game)
  gameKey := KeyForGame(c, region, riotGameId)
  err := datastore.RunInTransaction(c, func(c appengine.Context) error {
    err := datastore.Get(c, gameKey, game)
    if err == datastore.ErrNoSuchEntity {
      game.Region = region
      game.RiotGameId = riotGameId
      _, err = datastore.Put(c, gameKey, game)
    }
    return err
  }, nil)
  return game, gameKey, err
}

func GetPlayerGameStats(
  c appengine.Context,
  gameKey *datastore.Key,
  playerKey *datastore.Key,
  getKeysOnly bool) (*Game, *datastore.Key, error) {
  q := datastore.NewQuery("PlayerGameStats").
         Filter("GameKey =", gameKey).
         Filter("PlayerKey =", playerKey).
         Limit(1)
  if getKeysOnly {
    q = q.KeysOnly()
  }
  var games []*Game
  gameKeys, err := q.GetAll(c, &games)
  if len(gameKeys) == 0 {
    return nil, nil, err
  } else if len(games) == 0 {
    return nil, gameKeys[0], err
  }
  return games[0], gameKeys[0], err
}