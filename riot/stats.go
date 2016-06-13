package riot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
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

// stats-v1.3: https://developer.riotgames.com/api/methods#!/1080
// Lots of fields omitted.
type AggregatedStatsDto struct {
	BotGamesPlayed           int `json:"botGamesPlayed"`
	NormalGamesPlayed        int `json:"normalGamesPlayed"`
	RankedPremadeGamesPlayed int `json:"rankedPremadeGamesPlayed"`
	RankedSoloGamesPlayed    int `json:"rankedSoloGamesPlayed"`
	TotalSessionsPlayed      int `json:"totalSessionsPlayed"`
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
	data := make(map[string]*SummonerDto)
	json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, err
	}

	if dto, ok := data[name]; ok {
		return dto, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Summoner does not exist: %s", name))
	}
}

// Note that if the summoner id is not found nil gets populated into the output slice.
func SummonersById(
	urlFetcher func(string) ([]byte, error),
	riotApiKey string,
	region string,
	ids ...int64) ([]*SummonerDto, error) {
	ret := make([]*SummonerDto, len(ids))

	// API supports up to 40 summoners at a time.
	begin := 0
	end := min(len(ret), 40)

	idStrings := make([]string, len(ids))
	for i := 0; i < len(ids); i++ {
		idStrings[i] = strconv.FormatInt(ids[i], 10)
	}

	for ; ; begin, end = begin+40, min(len(ret), end+40) {
		batch := idStrings[begin:end]

		loc := ComposeUrl(
			riotApiKey,
			fmt.Sprintf("/api/lol/%s/v1.4/summoner/%s", region, strings.Join(batch, ",")),
			&url.Values{})
		jsonData, err := urlFetcher(loc)
		data := make(map[string]*SummonerDto)
		err = json.Unmarshal(jsonData, &data)
		if err != nil {
			return nil, err
		}

		for i := begin; i < end; i++ {
			if dto, ok := data[idStrings[i]]; ok {
				ret[i] = dto
			}
		}
		if end == len(ret) {
			break
		}
	}
	return ret, nil
}

func RunesBySummonerId(
	urlFetcher func(string) ([]byte, error),
	riotApiKey string,
	region string,
	riotId int64) (*RunePagesDto, error) {
	loc := ComposeUrl(
		riotApiKey,
		fmt.Sprintf("/api/lol/%s/v1.4/summoner/%d/runes", region, riotId),
		&url.Values{})
	jsonData, err := urlFetcher(loc)
	data := make(map[string]*RunePagesDto)
	json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, err
	}

	if dto, ok := data[fmt.Sprintf("%d", riotId)]; ok {
		return dto, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Summoner does not exist: %s-%d", region, riotId))
	}
}
