package model

import (
  "appengine"
  "appengine/datastore"
  "fmt"
  "github.com/OwenDurni/loltools/riot"
  "time"
)

type Game struct {
  Region string
  RiotId int64
  
  // The UTC datetime the game started.
  StartDateTime time.Time
}
func (g *Game) Id() string {
  return fmt.Sprintf("%s-%d", g.Region, g.RiotId)
}
func (g *Game) Uri() string {
  return fmt.Sprintf("/games/%s", g.Id())
}
// Returns whether g was changed and needs to be written back to datastore.
func (g *Game) UpdateLocalFromPlayerGameStats(stats *PlayerGameStats) (changed bool) {
  if g.StartDateTime != stats.GameStartDateTime {
    g.StartDateTime = stats.GameStartDateTime
    changed = true
  }
  return
}

func KeyForGame(c appengine.Context, region string, riotGameId int64) *datastore.Key {
  return datastore.NewKey(c, "Game", fmt.Sprintf("%s-%d", region, riotGameId), 0, nil)
}

// The tuple (GameKey, PlayerKey) are unique per entity.
type PlayerGameStats struct {
  GameKey   *datastore.Key
  PlayerKey *datastore.Key
  
  // The time the game started.
  GameStartDateTime time.Time
  
  // Stats for games that have expired out of recent player history on the Riot side
  // before we get around to looking for them may be lost forever. This field is set
  // to true if we know the game stats are no longer available and we don't have a
  // copy yet.
  NotAvailable bool
  
  // Set to true when we have captured the stats for this player and game already.
  Saved bool
  
  // The raw stats fetched from riot.
  RiotData riot.GameDto
}
func (p *PlayerGameStats) OtherPlayers(
  c appengine.Context, region string) ([]*Player, []*datastore.Key, error) {
  players := make([]*Player, 0, 12)
  playerKeys := make([]*datastore.Key, 0, 12)
  for _, riotPlayer := range p.RiotData.FellowPlayers {
    player, playerKey, err := GetOrCreatePlayerByRiotId(c, region, riotPlayer.SummonerId)
    if err != nil {
      return nil, nil, err
    }
    players = append(players, player)
    playerKeys = append(playerKeys, playerKey)
  }
  return players, playerKeys, nil
}

func KeyForPlayerGameStats(c appengine.Context, game *Game, player *Player) *datastore.Key {
  return datastore.NewKey(
    c, "PlayerGameStats", fmt.Sprintf("%s:%s", game.Id(), player.Id()), 0, nil)
}

func GetOrCreateGame(
  c appengine.Context, region string, riotGameId int64) (*Game, *datastore.Key, error) {
  game := new(Game)
  gameKey := KeyForGame(c, region, riotGameId)
  err := datastore.RunInTransaction(c, func(c appengine.Context) error {
    err := datastore.Get(c, gameKey, game)
    if err == datastore.ErrNoSuchEntity {
      game.Region = region
      game.RiotId = riotGameId
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