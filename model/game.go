package model

import (
  "appengine"
  "appengine/datastore"
  "fmt"
  "github.com/OwenDurni/loltools/riot"
  "github.com/OwenDurni/loltools/util/errwrap"
  "io"
  "time"
)

type Game struct {
  Region string
  RiotId int64
  
  // Fields below are populated from riot player stats.
  HasRiotData   bool
  StartDateTime time.Time
  MapId         int
  GameMode      string
  GameType      string
  SubType       string
  Players       []riot.PlayerDto
  Invalid       bool
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

func KeyForGame(c appengine.Context, region string, riotGameId int64) *datastore.Key {
  return KeyForGameId(c, MakeGameId(region, riotGameId))
}
func KeyForGameId(c appengine.Context, gameId string) *datastore.Key {
  return datastore.NewKey(c, "Game", gameId, 0, nil)
}

// The tuple (GameKey, PlayerKey) are unique per entity.
type PlayerGameStats struct {
  GameKey   *datastore.Key
  PlayerKey *datastore.Key
  
  // Stats for games that have expired out of recent player history on the Riot side
  // before we get around to looking for them may be lost forever. This field is set
  // to true if we know the game stats are no longer available and we don't have a
  // copy yet.
  NotAvailable bool
  
  // Set to true when we have captured the stats for this player and game already.
  Saved bool
  
  // The raw stats fetched from riot.
  RiotData riot.RawStatsDto
}

func KeyForPlayerGameStats(c appengine.Context, game *Game, player *Player) *datastore.Key {
  return KeyForPlayerGameStatsId(c, game.Id(), player.Id())
}
func KeyForPlayerGameStatsId(
  c appengine.Context, gameId string, playerId string) *datastore.Key {
  return datastore.NewKey(
    c, "PlayerGameStats", fmt.Sprintf("%s/%s", gameId, playerId), 0, nil)
}


type GameInfo struct {
  Game       *Game
  BlueTeam   *GameTeamInfo
  PurpleTeam *GameTeamInfo
  OtherTeam  *GameTeamInfo  // Only used if reported TeamId not recognized.
}
type GameTeamInfo struct {
  Players []*GamePlayerInfo
  PlayerStats []*GamePlayerStatsInfo
}
type GamePlayerInfo struct {
  Player *Player
  ChampionId int
}
type GamePlayerStatsInfo struct {
  Exists bool
  Player *Player
  Stats  *PlayerGameStats
}
func NewGameInfo() *GameInfo {
  info := new(GameInfo)
  info.Game = nil
  info.BlueTeam = NewGameTeamInfo()
  info.PurpleTeam = NewGameTeamInfo()
  info.OtherTeam = NewGameTeamInfo()
  return info
}
func NewGameTeamInfo() *GameTeamInfo {
  info := new(GameTeamInfo)
  info.Players = make([]*GamePlayerInfo, 0, 5)
  info.PlayerStats = make([]*GamePlayerStatsInfo, 0, 5)
  return info
}
func NewGamePlayerInfo(p *Player, championId int) *GamePlayerInfo {
  info := new(GamePlayerInfo)
  info.Player = p
  info.ChampionId = championId
  return info
}
func NewGamePlayerStatsInfo(player *Player, stats *PlayerGameStats) *GamePlayerStatsInfo {
  info := new(GamePlayerStatsInfo)
  info.Exists = (stats != nil)
  info.Player = player
  info.Stats = stats
  return info
}
func (ginfo *GameInfo) AddPlayer(teamId int, p *Player, championId int, pstats *PlayerGameStats) {
  pinfo := NewGamePlayerInfo(p, championId)
  psinfo := NewGamePlayerStatsInfo(p, pstats)
  if teamId == riot.BlueTeamId {
    ginfo.BlueTeam.Players = append(ginfo.BlueTeam.Players, pinfo)
    ginfo.BlueTeam.PlayerStats = append(ginfo.BlueTeam.PlayerStats, psinfo)
  } else if teamId == riot.PurpleTeamId {
    ginfo.PurpleTeam.Players = append(ginfo.PurpleTeam.Players, pinfo)
    ginfo.PurpleTeam.PlayerStats = append(ginfo.PurpleTeam.PlayerStats, psinfo)
  } else {
    ginfo.OtherTeam.Players = append(ginfo.OtherTeam.Players, pinfo)
    ginfo.OtherTeam.PlayerStats = append(ginfo.OtherTeam.PlayerStats, psinfo)
  }
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
    return errwrap.Wrap(err)
  }, nil)
  return game, gameKey, errwrap.Wrap(err)
}

func EnsureGameExists(
  c appengine.Context,
  region string,
  gameKey *datastore.Key,
  riotSummonerId int64,
  dto *riot.GameDto) error {
  game := new(Game)
  return datastore.RunInTransaction(c, func(c appengine.Context) error {
    err := datastore.Get(c, gameKey, game)
    if err != nil && err != datastore.ErrNoSuchEntity {
      return errwrap.Wrap(err)
    }
    if game.HasRiotData {
      return nil
    }
    game.Region = region
    game.RiotId = dto.GameId
    game.HasRiotData = true
    game.StartDateTime = (time.Time)(dto.CreateDate)
    game.MapId = dto.MapId
    game.GameMode = dto.GameMode
    game.GameType = dto.GameType
    game.SubType = dto.SubType
    players := make([]riot.PlayerDto, 0, len(dto.FellowPlayers)+1)
    players = append(players, riot.PlayerDto{
      ChampionId: dto.ChampionId,
      SummonerId: riotSummonerId,
      TeamId: dto.TeamId,
    })
    for _, playerDto := range dto.FellowPlayers {
      players = append(players, playerDto)
    }
    game.Players = players
    game.Invalid = dto.Invalid
    _, err = datastore.Put(c, gameKey, game)
    return errwrap.Wrap(err)
  }, nil)
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
    return nil, nil, errwrap.Wrap(err)
  } else if len(games) == 0 {
    return nil, gameKeys[0], errwrap.Wrap(err)
  }
  return games[0], gameKeys[0], errwrap.Wrap(err)
}

// Note that sometimes partial results are returned even if there is an error.
func TeamRecentGameInfo(
  c appengine.Context,
  n int,
  playerCache *PlayerCache,
  league *League,
  leagueKey *datastore.Key,
  teamKey *datastore.Key) ([]*GameInfo, []error) {
  infos := make([]*GameInfo, 0, n)
  errors := make([]error, 0)
  
  var gameKeys []*datastore.Key
  {
    var gamesByTeam []*GameByTeam
    q := datastore.NewQuery("GameByTeam").Ancestor(leagueKey).
           Project("GameKey").
           Filter("TeamKey =", teamKey).
           Order("-DateTime").
           Limit(n)
    if _, err := q.GetAll(c, &gamesByTeam); err != nil {
      errors = append(errors, errwrap.Wrap(err))
      return infos, errors
    }
    for _, g := range gamesByTeam {
      gameKeys = append(gameKeys, g.GameKey)
    }
  }
  
  games := make([]*Game, len(gameKeys))
  for i := range games {
    games[i] = new(Game)
  }
  if err := datastore.GetMulti(c, gameKeys, games); err != nil {
    errors = append(errors, errwrap.Wrap(err))
    if me, ok := err.(appengine.MultiError); ok {
      for i, merr := range me {
        if merr == datastore.ErrNoSuchEntity {
          games[i] = nil
          errors = append(errors, errwrap.Wrap(err))
        }
      }
    }
  }
  
  for _, game := range games {
    if game == nil { continue }
    info := NewGameInfo()
    infos = append(infos, info)
    info.Game = game
    
    for _, playerDto := range game.Players {
      summonerId := playerDto.SummonerId
      statKey := KeyForPlayerGameStatsId(c, game.Id(), MakePlayerId(game.Region, summonerId))
      player, err := playerCache.ById(summonerId)
      if err != nil {
        errors = append(errors, errwrap.Wrap(err))
      }
      pstats := new(PlayerGameStats)
      err = datastore.Get(c, statKey, pstats)
      if err != nil {
        if err == datastore.ErrNoSuchEntity {
          pstats = nil
        } else {
          errors = append(errors, errwrap.Wrap(err))
        }
      }
      info.AddPlayer(playerDto.TeamId, player, playerDto.ChampionId, pstats)
    }
  }
  
  return infos, errors
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
// Calls the provided function once for each game with a sample player's stat for that game.
func (s *CollectiveGameStats) ForEachGame(fn func(string, int64, *riot.GameDto)) {
  if s.games == nil { return }
  for gameId, playerMap := range s.games {
    for riotSummonerId, stat := range playerMap {
      if stat != nil {
        fn(gameId, riotSummonerId, stat)
        break
      }
    }
  }
}
// Calls the provided function once for each player's stat.
func (s *CollectiveGameStats) ForEachStat(fn func(string, int64, *riot.GameDto)) {
  if s.games == nil { return }
  for gameId, playerMap := range s.games {
    for playerId, stat := range playerMap {
      fn(gameId, playerId, stat)
    }
  }
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