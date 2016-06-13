package riot

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// stats-v1.3: https://developer.riotgames.com/api/methods#!/1080
type RankedStatsDto struct {
	Champions []*ChampionStatsDto `json:"champions"`
	//modifyDate
	SummonerId int64 `json:"summonerId"`
}

// stats-v1.3: https://developer.riotgames.com/api/methods#!/1080
type ChampionStatsDto struct {
	ChampionId int                 `json:"id"`
	Stats      *AggregatedStatsDto `json:"stats"`
}

// ChampionStatsDto.ChampionId for aggregated data about all champions.
const ChampionStatsDto_AllChampions = 0

// stats-v1.3: https://developer.riotgames.com/api/methods#!/1080
// Lots of fields omitted.
type AggregatedStatsDto struct {
	//BotGamesPlayed           int `json:"botGamesPlayed"`
	//NormalGamesPlayed        int `json:"normalGamesPlayed"`
	//RankedPremadeGamesPlayed int `json:"rankedPremadeGamesPlayed"`
	//RankedSoloGamesPlayed    int `json:"rankedSoloGamesPlayed"`
	TotalSessionsPlayed int `json:"totalSessionsPlayed"`
}

func RankedStatsBySummonerId(
	urlFetcher func(string) ([]byte, error),
	rateLimiter func(),
	riotApiKey string,
	region string,
	summonerId int64) (*RankedStatsDto, error) {
	loc := ComposeUrl(
		riotApiKey,
		fmt.Sprintf("/api/lol/%s/v1.3/stats/by-summoner/%d/ranked",
			region, summonerId),
		&url.Values{})
	rateLimiter()
	jsonData, err := urlFetcher(loc)
	dto := new(RankedStatsDto)
	json.Unmarshal(jsonData, &dto)
	if err != nil {
		return nil, err
	}

	return dto, err
}