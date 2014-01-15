package model

import (
  "appengine"
  "appengine/datastore"
  "fmt"
  "github.com/OwenDurni/loltools/riot"
  "io"
  "time"
)

type Game struct {
  Region string
  RiotId int64
  
  // The UTC datetime the game started.
  StartDateTime time.Time
}
func (g *Game) Id() string {
  return MakeGameId(g.Region, g.RiotId)
}
func (g *Game) Uri() string {
  return fmt.Sprintf("/games/%s", g.Id())
}
func MakeGameId(region string, riotId int64) string {
  return fmt.Sprintf("%s-%d", region, riotId)
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
  return datastore.NewKey(c, "Game", MakeGameId(region, riotGameId), 0, nil)
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

type CollectiveGameStats struct {
  // map[GameId]map[RiotSummonerId]*riot.GameDto
  games map[string]map[int64]*riot.GameDto
}
// Adds a player's stats to the collective stats and creates entries with nil stats
// for players in this game that have not been added to the map yet.
func (c *CollectiveGameStats) Add(
  region string, gameId string, riotSummonerId int64, stats *riot.GameDto) {
  if c.games == nil {
    c.games = make(map[string]map[int64]*riot.GameDto)
  }
  playerMap, exists := c.games[gameId]
  if !exists {
    playerMap = make(map[int64]*riot.GameDto)
    c.games[gameId] = playerMap
  }
  playerMap[riotSummonerId] = stats
  
  for _, riotOtherPlayer := range stats.FellowPlayers {
    c.stub(region, gameId, riotOtherPlayer.SummonerId)
  }
}
// Fills a stub with this player's stats. If no stub exists for these stats this is a no-op.
func (c *CollectiveGameStats) FillStub(
  region string, gameId string, riotSummonerId int64, stats *riot.GameDto) {
  if c.games == nil { return }
  playerMap, exists := c.games[gameId]
  if !exists { return }
  oldStats, exists := playerMap[riotSummonerId]
  if !exists { return }
  if oldStats != nil { return }
  playerMap[riotSummonerId] = stats
}
// Filters the stats collection to games that have at least n players from the specified
// list appearing on the same team.
func (s *CollectiveGameStats) FilterToGamesWithAtLeast(n int, players []*Player) {
  if (s.games == nil) { return }
  filteredGames := make(map[string]map[int64]*riot.GameDto)
  
  playerSet := make(map[int64]struct{})
  for _, player := range players {
    playerSet[player.RiotId] = struct{}{}
  }
  
  for gameId, playerMap := range s.games {
    for riotSummonerId, gameDto := range playerMap {
      if gameDto == nil {
        // Skip stubs
        continue
      }
      // Counts of the number of players matching on each team.
      counts := make(map[int]int)
      matches := false
      
      // Check the current player.
      if _, exists := playerSet[riotSummonerId]; exists {
        counts[gameDto.TeamId]++
      }
      
      // Check each fellow player.
      for _, playerDto := range gameDto.FellowPlayers {
        if _, exists := playerSet[playerDto.SummonerId]; exists {
          counts[playerDto.TeamId]++
          if counts[playerDto.TeamId] >= n {
            matches = true
            break
          }
        }
      }
      if matches {
        filteredGames[gameId] = playerMap
        break
      }
    }
  }
  s.games = filteredGames
}
func (s *CollectiveGameStats) Save(c appengine.Context) {
  
}
func (s *CollectiveGameStats) Size() int {
  if (s.games == nil) { return 0 }
  return len(s.games)
}
func (s *CollectiveGameStats) WriteDebugStringTo(w io.Writer) {
  if s.games == nil { return }
  for gameId, playerMap := range s.games {
    fmt.Fprintf(w, "Game %s:\n", gameId)
    for playerId, stat := range playerMap {
      if stat == nil {
        fmt.Fprintf(w, "  Player %d: <stub>\n", playerId)
      } else {
        fmt.Fprintf(w, "  Player %d: KDA(%d/%d/%d)\n",
                    playerId, stat.Stats.ChampionsKilled, stat.Stats.NumDeaths,
                    stat.Stats.Assists)
      }
    }
  }
}
func (s *CollectiveGameStats) stub(region string, gameId string, riotSummonerId int64) {
  if s.games == nil {
    s.games = make(map[string]map[int64]*riot.GameDto)
  }
  playerMap, exists := s.games[gameId]
  if !exists {
    playerMap = make(map[int64]*riot.GameDto)
    s.games[gameId] = playerMap
  }
  _, exists = playerMap[riotSummonerId]
  if !exists {
    playerMap[riotSummonerId] = nil
  }
}