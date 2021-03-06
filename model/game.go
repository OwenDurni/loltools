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
func (g *Game) Key(c appengine.Context) *datastore.Key {
	return KeyForGameId(c, g.Id())
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

	var gameModeString string

	switch subType {
	case "NONE":
		gameModeString = ""
	case "NORMAL":
		gameModeString = "Normal"
	case "BOT":
		gameModeString = "Normal (Bot)"
	case "RANKED_SOLO_5x5":
		gameModeString = "Ranked Solo"
	case "RANKED_PREMADE_3x3":
		gameModeString = "Ranked Duo 3v3"
	case "RANKED_PREMADE_5x5":
		gameModeString = "Ranked Duo"
	case "ODIN_UNRANKED":
		gameModeString = "Dominion"
	case "RANKED_TEAM_3x3":
		gameModeString = "Ranked Team 3v3"
	case "RANKED_TEAM_5x5":
		gameModeString = "Ranked Team 5v5"
	case "NORMAL_3x3":
		gameModeString = "Twisted Treeline"
	case "BOT_3x3":
		gameModeString = "Twisted Treeline (Bot)"
	case "CAP_5x5":
		gameModeString = "Team Builder"
	case "ARAM_UNRANKED_5x5":
		gameModeString = "ARAM"
	case "ONEFORALL_5x5":
		gameModeString = "One For All (Mirror Mode)"
	case "FIRSTBLOOD_1x1":
		gameModeString = "Showdown 1v1"
	case "FIRSTBLOOD_2x2":
		gameModeString = "Showdown 2v2"
	case "SR_6x6":
		gameModeString = "Hexakill"
	case "URF":
		gameModeString = "URF"
	case "URF_BOT":
		gameModeString = "URF (Bot)"
	default:
		gameModeString = subType
	}

	if custom {
		return fmt.Sprintf("Custom %s", gameModeString)
	} else {
		return gameModeString
	}
}

type GameByTime []*Game

func (a GameByTime) Len() int      { return len(a) }
func (a GameByTime) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a GameByTime) Less(i, j int) bool {
	return a[i].StartDateTime.Unix() < a[j].StartDateTime.Unix()
}

func GameUri(gameKey *datastore.Key) string {
	return fmt.Sprintf("/games/%s", gameKey.StringID())
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

	// Non-empty when viewed in the context of a league.
	LeagueId string

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
	IsBlueTeam   bool
	IsPurpleTeam bool
	Players      []*GamePlayerInfo
	PlayerStats  []*GamePlayerStatsInfo

	// stats
	IsWinner        bool
	ChampionsKilled int
	NumDeaths       int
	Assists         int
	GoldEarned      int
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

	gtinfo.Players = append(gtinfo.Players, pinfo)
	gtinfo.PlayerStats = append(gtinfo.PlayerStats, psinfo)
	gtinfo.IsWinner = pstats.RiotData.Win
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

func GameById(
	c appengine.Context,
	gameId string) (*Game, *datastore.Key, error) {
	game := new(Game)
	gameKey := KeyForGameId(c, gameId)
	err := datastore.Get(c, gameKey, game)
	if err != nil {
		return nil, nil, err
	}
	return game, gameKey, nil
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

func GetGameInfo(
	c appengine.Context,
	playerCache *PlayerCache,
	game *Game) (*GameInfo, []error) {
	info := NewGameInfo()
	info.Game = game
	errors := make([]error, 0, 8)

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
	return info, errors
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
		info, errs := GetGameInfo(c, playerCache, game)
		infos = append(infos, info)
		errors = append(errors, errs...)

		info.LeagueId = EncodeKeyShort(leagueKey)
		for _, p := range players {
			info.AddAppTeamPlayer(p)
		}
		info.computeDerivedData()
	}

	return infos, errors
}

type CollectiveGameStats struct {
	// map[GameId]map[RiotSummonerId]*riot.GameDto
	games map[string]*GameStats

	// The players we are interested in.
	playerSet map[int64]struct{}
}
type GameStats struct {
	// map[RiotSummonerId]*riot.GameDto
	stats map[int64]*riot.GameDto

	// map[RiotTeamId]{number of players on that team in the 'players' slice}
	teamCount map[int]int
}

func (g *GameStats) AddStats(riotSummonerId int64, stats *riot.GameDto) {
	if g.stats == nil {
		g.stats = make(map[int64]*riot.GameDto)
	}
	g.stats[riotSummonerId] = stats
	g.IncrementTeamCount(stats.TeamId)

	for _, riotOtherPlayer := range stats.FellowPlayers {
		g.stub(riotOtherPlayer.SummonerId)
	}
}
func (g *GameStats) GetTeamsWithCountAtLeast(target int) []int {
	var ret []int = make([]int, 0, 2)
	if g.teamCount == nil {
		return ret
	}
	for teamId, count := range g.teamCount {
		if count >= target {
			ret = append(ret, teamId)
		}
	}
	return ret
}
func (g *GameStats) IncrementTeamCount(riotTeamId int) {
	if g.teamCount == nil {
		g.teamCount = make(map[int]int)
	}
	g.teamCount[riotTeamId] = g.teamCount[riotTeamId] + 1
}
func (g *GameStats) Lookup(riotSummonerId int64) *riot.GameDto {
	if g.stats == nil {
		return nil
	}
	return g.stats[riotSummonerId]
}
func (g *GameStats) stub(riotSummonerId int64) {
	if g.stats == nil {
		g.stats = make(map[int64]*riot.GameDto)
	}
	_, exists := g.stats[riotSummonerId]
	if !exists {
		g.stats[riotSummonerId] = nil
	}
}

// Adds a player's stats to the collective stats and creates entries with nil stats
// for players in this game that have not been added to the map yet.
func (c *CollectiveGameStats) Add(
	gameId string, riotSummonerId int64, stats *riot.GameDto) {
	if c.games == nil {
		c.games = make(map[string]*GameStats)
	}
	gameStats, exists := c.games[gameId]
	if !exists {
		gameStats = new(GameStats)
		c.games[gameId] = gameStats
	}
	gameStats.AddStats(riotSummonerId, stats)
}
func (c *CollectiveGameStats) Lookup(gameId string, riotSummonerId int64) *riot.GameDto {
	if c.games == nil {
		return nil
	}
	gameStats, exists := c.games[gameId]
	if !exists {
		return nil
	}
	return gameStats.Lookup(riotSummonerId)
}

// Filters the stats collection to games that have at least n players from the specified
// list appearing on the same team.
func (s *CollectiveGameStats) FilterToGamesWithPlayersAtLeast(n int) {
	if s.games == nil {
		return
	}
	filteredGames := make(map[string]*GameStats)

	for gameId, gameStats := range s.games {
		if len(gameStats.GetTeamsWithCountAtLeast(n)) > 0 {
			filteredGames[gameId] = gameStats
		}
	}
	s.games = filteredGames
}

// Calls the provided function once for each game with a sample player's stat for that game.
func (s *CollectiveGameStats) ForEachGame(fn func(string, *GameStats, int64, *riot.GameDto)) {
	if s.games == nil {
		return
	}
	for gameId, gameStats := range s.games {
		for riotSummonerId, stat := range gameStats.stats {
			if stat != nil {
				fn(gameId, gameStats, riotSummonerId, stat)
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
	for gameId, gameStats := range s.games {
		if gameStats.stats == nil {
			continue
		}
		for playerId, stat := range gameStats.stats {
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
func (s *CollectiveGameStats) DebugString() string {
	ret := ""
	s.ForEachGame(func(
		gameId string,
		gameStats *GameStats,
		sampleSummonerId int64,
		sampleStats *riot.GameDto) {
		ret += fmt.Sprintf("Game %s: %+v\n", gameId, gameStats)
		for summonerId, stat := range gameStats.stats {
			ret += fmt.Sprintf("  Summoner %d: %+v\n", summonerId, stat)
		}
	})
	return ret
}
