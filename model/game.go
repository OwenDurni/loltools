package model

import (
  "appengine"
  "appengine/datastore"
  "fmt"
  "github.com/OwenDurni/loltools/riot"
  "github.com/OwenDurni/loltools/util/errwrap"
  "io"
  "strings"
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

func (game *Game) FormatTime() string {
    return game.StartDateTime.Format(time.Stamp)
}

func (game *Game) FormatGameType() string {
    gameType := game.GameType
    subType := game.SubType

    // Check if custom, ignore tutorial and matched.
    custom := gameType == "CUSTOM_GAME"

    var gameModeString string;

    switch subType {
      case "NONE": gameModeString = "None"
      case "NORMAL": gameModeString = "Normal"
      case "BOT": gameModeString = "Normal (Bot)"
      case "RANKED_SOLO_5x5": gameModeString = "Ranked Solo"
      case "RANKED_PREMADE_3x3": gameModeString = "Ranked Duo 3v3"
      case "RANKED_PREMADE_5x5": gameModeString = "Ranked Duo"
      case "ODIN_UNRANKED": gameModeString = "Dominion"
      case "RANKED_TEAM_3x3": gameModeString = "Ranked Team 3v3"
      case "RANKED_TEAM_5x5": gameModeString = "Ranked Team 5v5"
      case "NORMAL_3x3": gameModeString = "Twisted Treeline"
      case "BOT_3x3": gameModeString = "Twisted Treeline (Bot)"
      case "CAP_5x5": gameModeString = "Team Builder"
      case "ARAM_UNRANKED_5x5": gameModeString = "ARAM"
      case "ONEFORALL_5x5": gameModeString = "One For All (Mirror Mode)"
      case "FIRSTBLOOD_1x1": gameModeString = "Showdown 1v1"
      case "FIRSTBLOOD_2x2": gameModeString = "Showdown 2v2"
      case "SR_6x6": gameModeString = "Hexakill"
      case "URF": gameModeString = "URF"
      case "URF_BOT": gameModeString = "URF (Bot)"
      default: gameModeString = subType
    }

    if custom {
      return fmt.Sprintf("Custom %s", gameModeString)
    } else {
      return gameModeString
    }
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

func RegionForGameId(id string) string {
  return strings.Split(id, "-")[0]
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
  Game *Game

  BlueTeam   *GameTeamInfo
  PurpleTeam *GameTeamInfo

  // Aliases for BlueTeam/PurpleTeam based on which side has more summoners that are players
  // of the team someone is inspecting in the application.
  ThisTeam  *GameTeamInfo
  OtherTeam *GameTeamInfo

  // Set of summoner ids for members of the team someone is inspecting in the application.
  appTeamSummonerIdSet map[int64]bool

  // Map from {riot team id}
  // to {number of summoners on that team that are in appTeamSummonerIdSet}
  riotTeamPlayerCounts map[int]int
}
type GameTeamInfo struct {
  IsBlueTeam        bool
  IsPurpleTeam      bool
  Players           []*GamePlayerInfo
  PlayerStats       []*GamePlayerStatsInfo

  // stats
  ChampionsKilled   int
  NumDeaths         int
  Assists           int
  GoldEarned        int
}
type GamePlayerInfo struct {
  Player     *Player
  ChampionId int
}
type GamePlayerStatsInfo struct {
  Saved        bool
  NotAvailable bool
  IsOnAppTeam  bool
  Player       *Player
  Stats        *PlayerGameStats
}

func NewGameInfo() *GameInfo {
  info := new(GameInfo)
  info.Game = nil
  info.BlueTeam = NewGameTeamInfo()
  info.BlueTeam.IsBlueTeam = true
  info.PurpleTeam = NewGameTeamInfo()
  info.PurpleTeam.IsPurpleTeam = true
  info.ThisTeam = info.BlueTeam
  info.OtherTeam = info.PurpleTeam
  info.appTeamSummonerIdSet = make(map[int64]bool)
  info.riotTeamPlayerCounts = make(map[int]int)
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
func NewGamePlayerStatsInfo(
  player *Player, stats *PlayerGameStats, isOnAppTeam bool) *GamePlayerStatsInfo {
  info := new(GamePlayerStatsInfo)
  if stats != nil {
    info.Saved = stats.Saved
    info.NotAvailable = stats.NotAvailable
  }
  info.IsOnAppTeam = isOnAppTeam
  info.Player = player
  info.Stats = stats
  return info
}

// When this game was looked up in the context of a team registered with the application,
// this adds that player to the GameInfo.
func (ginfo *GameInfo) AddAppTeamPlayer(p *Player) {
  ginfo.appTeamSummonerIdSet[p.RiotId] = true
}

// Adds a player and their in-game-stats to the GameInfo.
func (ginfo *GameInfo) AddGamePlayer(
  teamId int, p *Player, championId int, pstats *PlayerGameStats) {

  isOnAppTeam := ginfo.appTeamSummonerIdSet[p.RiotId]

  pinfo := NewGamePlayerInfo(p, championId)
  psinfo := NewGamePlayerStatsInfo(p, pstats, isOnAppTeam)

  var gtinfo *GameTeamInfo
  if teamId == riot.BlueTeamId {
    gtinfo = ginfo.BlueTeam
  } else if teamId == riot.PurpleTeamId {
    gtinfo = ginfo.PurpleTeam
  } else {
    // We don't know which team to add the player to...
    return
  }

  if isOnAppTeam {
    ginfo.riotTeamPlayerCounts[teamId]++
  }

  if gtinfo != nil {
    gtinfo.Players = append(gtinfo.Players, pinfo)
    gtinfo.PlayerStats = append(gtinfo.PlayerStats, psinfo)
  }

  gtinfo.ChampionsKilled += pstats.RiotData.ChampionsKilled
  gtinfo.NumDeaths += pstats.RiotData.NumDeaths
  gtinfo.Assists += pstats.RiotData.Assists

  gtinfo.GoldEarned += pstats.RiotData.GoldEarned
}

// Fills fields of GameInfo that are derived from other fields.
func (ginfo *GameInfo) computeDerivedData() {
  if ginfo.riotTeamPlayerCounts[riot.PurpleTeamId] > ginfo.riotTeamPlayerCounts[riot.BlueTeamId] {
    ginfo.ThisTeam = ginfo.PurpleTeam
    ginfo.OtherTeam = ginfo.BlueTeam
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
      TeamId:     dto.TeamId,
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
  userAcls *RequestorAclCache,
  n int,
  playerCache *PlayerCache,
  league *League,
  leagueKey *datastore.Key,
  teamKey *datastore.Key) ([]*GameInfo, []error) {
  infos := make([]*GameInfo, 0, n)
  errors := make([]error, 0)

  players, _, err := TeamAllPlayers(c, userAcls, league, leagueKey, teamKey, KeysAndEntities)
  if err != nil {
    errors = append(errors, errwrap.Wrap(err))
    return nil, errors
  }

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
    if game == nil {
      continue
    }
    info := NewGameInfo()
    infos = append(infos, info)
    info.Game = game

    for _, p := range players {
      info.AddAppTeamPlayer(p)
    }

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
      info.AddGamePlayer(playerDto.TeamId, player, playerDto.ChampionId, pstats)
    }
    info.computeDerivedData()
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
func (c *CollectiveGameStats) Lookup(gameId string, riotSummonerId int64) *riot.GameDto {
  if c.games == nil {
    return nil
  }
  playerMap, exists := c.games[gameId]
  if !exists {
    return nil
  }
  stats, exists := playerMap[riotSummonerId]
  if !exists {
    return nil
  }
  return stats
}

// Fills a stub with this player's stats. If no stub exists for these stats this is a no-op.
func (c *CollectiveGameStats) FillStub(
  region string, gameId string, riotSummonerId int64, stats *riot.GameDto) {
  if c.games == nil {
    return
  }
  playerMap, exists := c.games[gameId]
  if !exists {
    return
  }
  oldStats, exists := playerMap[riotSummonerId]
  if !exists {
    return
  }
  if oldStats != nil {
    return
  }
  playerMap[riotSummonerId] = stats
}

// Filters the stats collection to games that have at least n players from the specified
// list appearing on the same team.
func (s *CollectiveGameStats) FilterToGamesWithAtLeast(n int, players []*Player) {
  if s.games == nil {
    return
  }
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
  if s.games == nil {
    return
  }
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
  if s.games == nil {
    return
  }
  for gameId, playerMap := range s.games {
    for playerId, stat := range playerMap {
      fn(gameId, playerId, stat)
    }
  }
}
func (s *CollectiveGameStats) Size() int {
  if s.games == nil {
    return 0
  }
  return len(s.games)
}
func (s *CollectiveGameStats) WriteDebugStringTo(w io.Writer) {
  if s.games == nil {
    return
  }
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
