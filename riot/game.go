package riot

import (
  "encoding/json"
  "fmt"
  "net/url"
)

// v1.3: http://developer.riotgames.com/api/methods#!/339/1143
type RecentGamesDto struct {
  Games      []GameDto `json:"games"`
  SummonerId int64     `json:"summonerId"`
}

// v1.3: http://developer.riotgames.com/api/methods#!/339/1143
type GameDto struct {
  GameId     int64    `json:"gameId"`
  MapId      int      `json:"mapId"`
  CreateDate RiotTime `json:"createDate"`

  GameMode string `json:"gameMode"`
  GameType string `json:"gameType"`
  SubType  string `json:"subType"`

  TeamId        int         `json:"teamId"`
  FellowPlayers []PlayerDto `json:"fellowPlayers"`

  ChampionId     int         `json:"championId"`
  Level          int         `json:"level"`
  SummonerSpell1 int         `json:"spell1"`
  SummonerSpell2 int         `json:"spell2"`
  Stats          RawStatsDto `json:"stats"`

  Invalid bool `json:"invalid"`
}

// v1.3: http://developer.riotgames.com/api/methods#!/339/1143
type PlayerDto struct {
  ChampionId int   `json:"championId"`
  SummonerId int64 `json:"summonerId"`
  TeamId     int   `json:"teamId"`
}

// v1.3: http://developer.riotgames.com/api/methods#!/339/1143
//
// Fields that are not used in the app are commented out to save datastore ops.
// This is the struct that appears most often in the datastore.
type RawStatsDto struct {
  // Overall Game Stats
  Win        bool `json:"win"`
  TimePlayed int  `json:"timePlayed"`
  //Team       int  `json:"team"`
  ChampionId int /* populated from GameDto */

  // KDA
  ChampionsKilled int `json:"championsKilled"`
  NumDeaths       int `json:"numDeaths"`
  Assists         int `json:"assists"`
  //TurretsKilled   int `json:"turretsKilled"`
  //BarracksKilled  int `json:"barracksKilled"`

  // Gold/XP
  Level int `json:"level"`
  //Gold       int `json:"gold"`
  GoldEarned int `json:"goldEarned"`
  //GoldSpent  int `json:"goldSpent"`

  // CS
  //MinionsDenied                   int `json:"minionsDenied"`
  MinionsKilled                   int `json:"minionsKilled"`
  NeutralMinionsKilled            int `json:"neutralMinionsKilled"`
  NeutralMinionsKilledEnemyJungle int `json:"neutralMinionsKilledEnemyJungle"`
  NeutralMinionsKilledYourJungle  int `json:"neutralMinionsKilledYourJungle"`
  SuperMonstersKilled             int `json:"superMonsterKilled"`

  // Items
  Item0 int `json:"item0"`
  Item1 int `json:"item1"`
  Item2 int `json:"item2"`
  Item3 int `json:"item3"`
  Item4 int `json:"item4"`
  Item5 int `json:"item5"`
  Item6 int `json:"item6"`
  //ItemsPurchased        int `json:"itemsPurchased"`
  //ConsumablesPurchased  int `json:"consumablesPurchased"`
  //NumItemsBought        int `json:"numItemsBought"`
  //LegendaryItemsCreated int `json:"legendaryItemsCreated"`

  // Vision
  WardPlaced        int `json:"wardPlaced"`
  SightWardsBought  int `json:"sightWardsBought"`
  VisionWardsBought int `json:"visionWardsBought"`
  WardKilled        int `json:"wardKilled"`

  // Summoners
  SummonerSpell1 int /* populated from GameDto */
  SummonerSpell2 int /* populated from GameDto */

  // Damage Dealt
  //DamageDealtPlayer              int `json:"damageDealtPlayer"`
  //TotalDamageDealt               int `json:"totalDamageDealt"`
  //TotalDamageDealtToChampions    int `json:"totalDamageDealtToChampions"`
  //PhysicalDamageDealtPlayer      int `json:"physicalDamageDealtPlayer"`
  PhysicalDamageDealtToChampions int `json:"physicalDamageDealtToChampions"`
  //MagicDamageDealtPlayer         int `json:"magicDamageDealtPlayer"`
  MagicDamageDealtToChampions int `json:"magicDamageDealtToChampions"`
  //TrueDamageDealtPlayer          int `json:"trueDamageDealtPlayer"`
  TrueDamageDealtToChampions int `json:"trueDamageDealtToChampions"`

  // Damage Taken
  //TotalDamageTaken    int `json:"totalDamageTaken"`
  PhysicalDamageTaken int `json:"physicalDamageTaken"`
  MagicDamageTaken    int `json:"magicDamageTaken"`
  TrueDamageTaken     int `json:"trueDamageTaken"`

  // Misc Dealt Stats
  TotalTimeCrowdControlDealt int `json:"totalTimeCrowdControlDealt"`
  //TotalHeal                  int `json:"totalHeal"`
  //TotalUnitsHealed           int `json:"totalUnitsHealed"`

  // Vanity stats.
  //KillingSprees         int  `json:"killingSprees"`
  //LargestKillingSpree   int  `json:"largestKillingSpree"`
  LargestMultiKill int `json:"largestMultiKill"`
  //DoubleKills           int  `json:"doubleKills"`
  //TripleKills           int  `json:"tripleKills"`
  //QuadraKills           int  `json:"quadraKills"`
  //PentaKills            int  `json:"pentaKills"`
  //UnrealKills           int  `json:"unrealKills"`
  //NexusKilled           bool `json:"nexusKilled"`
  FirstBlood int `json:"firstBlood"`
  //LargestCriticalStrike int  `json:"largestCriticalStrike"`

  // Number of times various spells were cast.
  //Spell1Cast         int `json:"spell1Cast"`
  //Spell2Cast         int `json:"spell2Cast"`
  //Spell3Cast         int `json:"spell3Cast"`
  //Spell4Cast         int `json:"spell4Cast"`
  //SummonerSpell1Cast int `json:"summonSpell1Cast"`
  //SummonerSpell2Cast int `json:"summonSpell2Cast"`

  // Dominion
  //VictoryPointTotal    int `json:"victoryPointTotal"`
  //NodeCapture          int `json:"nodeCapture"`
  //NodeCaptureAssist    int `json:"nodeCaptureAssist"`
  //NodeNeutralize       int `json:"nodeNeutralize"`
  //NodeNeutralizeAssist int `json:"nodeNeutralizeAssist"`

  // ???
  //TotalPlayerScore     int `json:"totalPlayerScore"`
  //TotalScoreRank       int `json:"totalScoreRank"`
  //CombatPlayerScore    int `json:"combatPlayerScore"`
  //ObjectivePlayerScore int `json:"objectivePlayerScore"`
  //TeamObjective        int `json:"teamObjective"`
}

func GameStatsForPlayer(
  urlFetcher func(string) ([]byte, error),
  rateLimiter func(),
  riotApiKey string,
  region string,
  riotSummonerId int64) (*RecentGamesDto, error) {
  loc := ComposeUrl(
    riotApiKey,
    fmt.Sprintf("/api/lol/%s/v1.3/game/by-summoner/%d/recent", region, riotSummonerId),
    &url.Values{})
  g := new(RecentGamesDto)
  g.SummonerId = riotSummonerId

  rateLimiter()
  jsonData, err := urlFetcher(loc)
  if err != nil {
    if err, ok := err.(ErrRiotRestApi); ok && err.HttpStatusCode == 404 {
      // 404 means no match history for this summoner id.
      return g, nil
    }

    return nil, err
  }

  if err = json.Unmarshal(jsonData, g); err == nil {
    // Do some post-processing.
    for i, _ := range g.Games {
      var gameDto *GameDto = &g.Games[i]
      gameDto.Stats.ChampionId = gameDto.ChampionId
      gameDto.Stats.SummonerSpell1 = gameDto.SummonerSpell1
      gameDto.Stats.SummonerSpell2 = gameDto.SummonerSpell2
    }
  }
  return g, err
}
