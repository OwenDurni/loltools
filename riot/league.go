package riot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

// league-v2.5: https://developer.riotgames.com/api/methods#!/985
type LeagueDto struct {
	Entries       []*LeagueEntryDto `json:"entries"`
	ParticipantId string            `json:"participantId"`
	Queue         string            `json:"queue"`
	Tier          string            `json:"tier"`
}

// league-v2.5: https://developer.riotgames.com/api/methods#!/985
type LeagueEntryDto struct {
	Division     string `json:"division"`
	LeaguePoints int    `json:"leaguePoints"`
}

type Rank struct {
	Tier         string // BRONZE, SILVER, GOLD, PLATINUM, DIAMOND, MASTER, CHALLENGER
	Division     string // I, II, III, IV, V
	LeaguePoints int
}

func (r *Rank) String() string {
	if r.Division == "" {
		return r.Tier
	}
	return fmt.Sprintf("%s %s %dLP", r.Tier, r.Division, r.LeaguePoints)
}

func LeagueInfoBySummonerId(
	urlFetcher func(string) ([]byte, int, error),
	rateLimiter func(),
	riotApiKey string,
	region string,
	summonerId int64) ([]*LeagueDto, error) {
	loc := ComposeUrl(
		riotApiKey,
		fmt.Sprintf("/api/lol/%s/v2.5/league/by-summoner/%d/entry",
			region, summonerId),
		&url.Values{})
	rateLimiter()
	jsonData, httpStatus, err := urlFetcher(loc)
	if err != nil {
		return nil, err
	}
	if httpStatus == 404 {
		return nil, NotFound{}
	}
	data := make(map[string][]*LeagueDto)
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, err
	}
	if dto, ok := data[strconv.FormatInt(summonerId, 10)]; ok {
		return dto, nil
	}
	return nil, errors.New("Riot data did not contain info for this summoner")
}

func SoloQueueRankBySummonerId(
	urlFetcher func(string) ([]byte, int, error),
	rateLimiter func(),
	riotApiKey string,
	region string,
	summonerId int64) (*Rank, error) {
	leagueDtos, err := LeagueInfoBySummonerId(
		urlFetcher, rateLimiter, riotApiKey, region, summonerId)
	if err != nil {
		if _, ok := err.(NotFound); ok {
			// No league exists for this summoner id.
			return &Rank{"Unranked", "", 0}, nil
		}
		return nil, err
	}
	for _, leagueDto := range leagueDtos {
		if leagueDto.Queue == "RANKED_SOLO_5x5" {
			if len(leagueDto.Entries) != 1 {
				return nil, errors.New("Unexpected response from Riot API")
			}
			leagueDtoEntry := leagueDto.Entries[0]
			rank := &Rank{
				leagueDto.Tier, leagueDtoEntry.Division, leagueDtoEntry.LeaguePoints}
			return rank, nil
		}
	}
	// No soloQ data found, assume unranked.
	return &Rank{"Unranked", "", 0}, nil
}
