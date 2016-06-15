package riot

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// match-v2.2: https://developer.riotgames.com/api/methods#!/1064
type MatchDetail struct {
	MapId                 int                    `json:"mapId"`
	MatchCreation         int64                  `json:"matchCreation"`
	MatchDuration         int64                  `json:"matchDuration"`
	MatchId               int64                  `json:"matchId"`
	MatchMode             string                 `json:"matchMode"`
	MatchType             string                 `json:"matchType"`
	MatchVersion          string                 `json:"matchVersion"`
	ParticipantIdentities []*ParticipantIdentity `json:"participantIdentities"`
	Participants          []*Participant         `json:"participants"`
	PlatformId            string                 `json:"platformId"`
	QueueType             string                 `json:"queueType"`
	Region                string                 `json:"region"`
	Season                string                 `json:"season"`
	Teams                 []*Team                `json:"teams"`
	Timeline              *Timeline              `json:"timeline"`
}

// match-v2.2: https://developer.riotgames.com/api/methods#!/1064
type Participant struct {
	ChampionId                int                  `json:"championId"`
	HighestAchievedSeasonTier string               `json:"highestAchievedSeasonTier"`
	Masteries                 []*Mastery           `json:"masteries"`
	ParticipantId             int                  `json:"participantId"`
	Runes                     []*Rune              `json:"runes"`
	Summoner1                 int                  `json:"spell1Id"`
	Summoner2                 int                  `json:"spell2Id"`
	Stats                     *ParticipantStats    `json:"stats"`
	TeamId                    int                  `json:"teamId"`
	Timeline                  *ParticipantTimeline `json:"timeline"`
}

// match-v2.2: https://developer.riotgames.com/api/methods#!/1064
type ParticipantIdentity struct {
	ParticipantId int     `json:"participantId"`
	Player        *Player `json:"player"`
}

// match-v2.2: https://developer.riotgames.com/api/methods#!/1064
type Team struct {
	Bans                 []*BannedChampion `json:"bans"`
	BaronKills           int               `json:"baronKills"`
	DominionVictoryScore int64             `json:"dominionVictoryScore"`
	DragonKills          int               `json:"dragonKills"`
	FirstBaron           bool              `json:"firstBaron"`
	FirstBlood           bool              `json:"firstBlood"`
	FirstDragon          bool              `json:"firstDragon"`
	FirstInhibitor       bool              `json:"firstInhibitor"`
	FirstRiftHerald      bool              `json:"firstRiftHerald"`
	FirstTower           bool              `json:"firstTower"`
	InhibitorKills       int               `json:"inhibitorKills"`
	RiftHeraldKills      int               `json:"riftHeraldKills"`
	TeamId               int               `json:"teamId"`
	TowerKills           int               `json:"towerKills"`
	VilemawKills         int               `json:"vilemawKills"`
	Winner               bool              `json:"winner"`
}

// match-v2.2: https://developer.riotgames.com/api/methods#!/1064
type Timeline struct {
	FrameInterval int64    `json:"frameInterval"`
	Frames        []*Frame `json:"frames"`
}

// match-v2.2: https://developer.riotgames.com/api/methods#!/1064
type Mastery struct {
	MasteryId int64 `json:"masteryId"`
	Rank      int64 `json:"rank"`
}

// match-v2.2: https://developer.riotgames.com/api/methods#!/1064
type ParticipantStats struct {
	Assists                         int64 `json:"assists"`
	ChampLevel                      int64 `json:"champLevel"`
	CombatPlayerScore               int64 `json:"combatPlayerScore"`
	Deaths                          int64 `json:"deaths"`
	DoubleKills                     int64 `json:"doubleKills"`
	FirstBloodAssist                bool  `json:"firstBloodAssist"`
	FirstBloodKill                  bool  `json:"firstBloodKill"`
	FirstInhibitorAssist            bool  `json:"firstInhibitorAssist"`
	FirstInhibitorKill              bool  `json:"firstInhibitorKill"`
	FirstTowerAssist                bool  `json:"firstTowerAssist"`
	FirstTowerKill                  bool  `json:"firstTowerKill"`
	GoldEarned                      int64 `json:"goldEarned"`
	GoldSpent                       int64 `json:"goldSpent"`
	InhibitorKills                  int64 `json:"inhibitorKills"`
	Item0                           int64 `json:"item0"`
	Item1                           int64 `json:"item1"`
	Item2                           int64 `json:"item2"`
	Item3                           int64 `json:"item3"`
	Item4                           int64 `json:"item4"`
	Item5                           int64 `json:"item5"`
	Item6                           int64 `json:"item6"`
	KillingSprees                   int64 `json:"killingSprees"`
	Kills                           int64 `json:"kills"`
	LargestCriticalStrike           int64 `json:"largestCriticalStrike"`
	LargestKillingSpree             int64 `json:"largestKillingSpree"`
	LargestMultiKill                int64 `json:"largestMultiKill"`
	MagicDamageDealt                int64 `json:"magicDamageDealt"`
	MagicDamageDealtToChampions     int64 `json:"magicDamageDealtToChampions"`
	MagicDamageTaken                int64 `json:"magicDamageTaken"`
	MinionsKilled                   int64 `json:"minionsKilled"`
	NeutralMinionsKilled            int64 `json:"neutralMinionsKilled"`
	NeutralMinionsKilledEnemyJungle int64 `json:"neutralMinionsKilledEnemyJungle"`
	NeutralMinionsKilledTeamJungle  int64 `json:"neutralMinionsKilledTeamJungle"`
	NodeCapture                     int64 `json:"nodeCapture"`
	NodeCaptureAssist               int64 `json:"nodeCaptureAssist"`
	NodeNeutralize                  int64 `json:"nodeNeutralize"`
	NodeNeutralizeAssist            int64 `json:"nodeNeutralizeAssist"`
	ObjectivePlayerScore            int64 `json:"objectivePlayerScore"`
	PentaKills                      int64 `json:"pentaKills"`
	PhysicalDamageDealt             int64 `json:"physicalDamageDealt"`
	PhysicalDamageDealtToChampions  int64 `json:"physicalDamageDealtToChampions"`
	PhysicalDamageTaken             int64 `json:"physicalDamageTaken"`
	QuadraKills                     int64 `json:"quadraKills"`
	SightWardsBoughtInGame          int64 `json:"sightWardsBoughtInGame"`
	TeamObjective                   int64 `json:"teamObjective"`
	TotalDamageDealt                int64 `json:"totalDamageDealt"`
	TotalDamageDealtToChampions     int64 `json:"totalDamageDealtToChampions"`
	TotalDamageTaken                int64 `json:"totalDamageTaken"`
	TotalHeal                       int64 `json:"totalHeal"`
	TotalPlayerScore                int64 `json:"totalPlayerScore"`
	TotalScoreRank                  int64 `json:"totalScoreRank"`
	TotalTimeCrowdControlDealt      int64 `json:"totalTimeCrowdControlDealt"`
	TotalUnitsHealed                int64 `json:"totalUnitsHealed"`
	TowerKills                      int64 `json:"towerKills"`
	TripleKills                     int64 `json:"tripleKills"`
	TrueDamageDealt                 int64 `json:"trueDamageDealt"`
	TrueDamageDealtToChampions      int64 `json:"trueDamageDealtToChampions"`
	TrueDamageTaken                 int64 `json:"trueDamageTaken"`
	UnrealKills                     int64 `json:"unrealKills"`
	VisionWardsBoughtInGame         int64 `json:"visionWardsBoughtInGame"`
	WardsKilled                     int64 `json:"wardsKilled"`
	WardsPlaced                     int64 `json:"wardsPlaced"`
	Winner                          bool  `json:"winner"`
}

// match-v2.2: https://developer.riotgames.com/api/methods#!/1064
type ParticipantTimeline struct {
	AncientGolemAssistsPerMinCounts *ParticipantTimelineData `json:"ancientGolemAssistsPerMinCounts"`
	AncientGolemKillsPerMinCounts   *ParticipantTimelineData `json:"ancientGolemKillsPerMinCounts"`
	AssistedLaneDeathsPerMinDeltas  *ParticipantTimelineData `json:"assistedLaneDeathsPerMinDeltas"`
	AssistedLaneKillsPerMinDeltas   *ParticipantTimelineData `json:"assistedLaneKillsPerMinDeltas"`
	BaronAssistsPerMinCounts        *ParticipantTimelineData `json:"baronAssistsPerMinCounts"`
	BaronKillsPerMinCounts          *ParticipantTimelineData `json:"baronKillsPerMinCounts"`
	CreepsPerMinDeltas              *ParticipantTimelineData `json:"creepsPerMinDeltas"`
	CsDiffPerMinDeltas              *ParticipantTimelineData `json:"csDiffPerMinDeltas"`
	DamageTakenDiffPerMinDeltas     *ParticipantTimelineData `json:"damageTakenDiffPerMinDeltas"`
	DamageTakenPerMinDeltas         *ParticipantTimelineData `json:"damageTakenPerMinDeltas"`
	DragonAssistsPerMinCounts       *ParticipantTimelineData `json:"dragonAssistsPerMinCounts"`
	DragonKillsPerMinCounts         *ParticipantTimelineData `json:"dragonKillsPerMinCounts"`
	ElderLizardAssistsPerMinCounts  *ParticipantTimelineData `json:"elderLizardAssistsPerMinCounts"`
	ElderLizardKillsPerMinCounts    *ParticipantTimelineData `json:"elderLizardKillsPerMinCounts"`
	GoldPerMinDeltas                *ParticipantTimelineData `json:"goldPerMinDeltas"`
	InhibitorAssistsPerMinCounts    *ParticipantTimelineData `json:"inhibitorAssistsPerMinCounts"`
	InhibitorKillsPerMinCounts      *ParticipantTimelineData `json:"inhibitorKillsPerMinCounts"`
	Lane                            string                   `json:"lane"`
	Role                            string                   `json:"role"`
	TowerAssistsPerMinCounts        *ParticipantTimelineData `json:"towerAssistsPerMinCounts"`
	TowerKillsPerMinCounts          *ParticipantTimelineData `json:"towerKillsPerMinCounts"`
	TowerKillsPerMinDeltas          *ParticipantTimelineData `json:"towerKillsPerMinDeltas"`
	VilemawAssistsPerMinCounts      *ParticipantTimelineData `json:"vilemawAssistsPerMinCounts"`
	VilemawKillsPerMinCounts        *ParticipantTimelineData `json:"vilemawKillsPerMinCounts"`
	WardsPerMinDeltas               *ParticipantTimelineData `json:"wardsPerMinDeltas"`
	XpDiffPerMinDeltas              *ParticipantTimelineData `json:"xpDiffPerMinDeltas"`
	XpPerMinDeltas                  *ParticipantTimelineData `json:"xpPerMinDeltas"`
}

// match-v2.2: https://developer.riotgames.com/api/methods#!/1064
type Rune struct {
	Rank   int64 `json:"rank"`
	RuneId int64 `json:"runeId"`
}

// match-v2.2: https://developer.riotgames.com/api/methods#!/1064
type Player struct {
	MatchHistoryUri string `json:"matchHistoryUri"`
	ProfileIcon     int    `json:"profileIcon"`
	SummonerId      int64  `json:"summonerId"`
	SummonerName    string `json:"summonerName"`
}

// match-v2.2: https://developer.riotgames.com/api/methods#!/1064
type BannedChampion struct {
	ChampionId int `json:"championId"`
	PickTurn   int `json:"pickTurn"`
}

// match-v2.2: https://developer.riotgames.com/api/methods#!/1064
type Frame struct {
	Events            []*Event                    `json:"events"`
	ParticipantFrames map[string]ParticipantFrame `json:"participantFrames"`
	Timestamp         RiotTime                    `json:"timestamp"`
}

// match-v2.2: https://developer.riotgames.com/api/methods#!/1064
type ParticipantTimelineData struct {
	ZeroToTen      float64 `json:"zeroToTen"`
	TenToTwenty    float64 `json:"tenToTwenty"`
	TwentyToThirty float64 `json:"twentyToThirty"`
	ThirtyToEnd    float64 `json:"thirtyToEnd"`
}

// match-v2.2: https://developer.riotgames.com/api/methods#!/1064
type Event struct {
	AscendedType            string   `json:"ascendedType"`
	AssistingParticipantIds []int    `json:"assistingParticipantIds"`
	BuildingType            string   `json:"buildingType"`
	CreatorId               int      `json:"creatorId"`
	EventType               string   `json:"eventType"`
	ItemAfter               int      `json:"itemAfter"`
	ItemBefore              int      `json:"itemBefore"`
	ItemId                  int      `json:"itemId"`
	KillerId                int      `json:"killerId"`
	LaneType                string   `json:"laneType"`
	LevelUpType             string   `json:"levelUpType"`
	MonsterType             string   `json:"monsterType"`
	ParticipantId           int      `json:"participantId"`
	PointCaptured           string   `json:"pointCaptured"`
	Position                Position `json:"position"`
	SkillSlot               int      `json:"skillSlot"`
	TeamId                  int      `json:"teamId"`
	Timestamp               int64    `json:"timestamp"`
	TowerType               string   `json:"towerType"`
	VictimId                int      `json:"victimId"`
	WardType                string   `json:"wardType"`
}

// match-v2.2: https://developer.riotgames.com/api/methods#!/1064
type ParticipantFrame struct {
	CurrentGold         int       `json:"currentGold"`
	DominionScore       int       `json:"dominionScore"`
	JungleMinionsKilled int       `json:"jungleMinionsKilled"`
	Level               int       `json:"level"`
	MinionsKilled       int       `json:"minionsKilled"`
	ParticipantId       int       `json:"participantId"`
	Position            *Position `json:"position"`
	TeamScore           int       `json:"teamScore"`
	TotalGold           int       `json:"totalGold"`
	Xp                  int       `json:"xp"`
}

// match-v2.2: https://developer.riotgames.com/api/methods#!/1064
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func LookupMatch(
	urlFetcher func(string) ([]byte, int, error),
	rateLimiter func(),
	riotApiKey string,
	region string,
	matchId int64) (*MatchDetail, error) {
	loc := ComposeUrl(
		riotApiKey,
		fmt.Sprintf("/api/lol/%s/v2.2/match/%d", region, matchId),
		&url.Values{})
	rateLimiter()
	jsonData, httpStatus, err := urlFetcher(loc)
	if err != nil {
		return nil, err
	}
	if httpStatus == 404 {
		return nil, NotFound{}
	}
	match := new(MatchDetail)
	err = json.Unmarshal(jsonData, &match)
	if err != nil {
		return nil, err
	}
	return match, nil
}
