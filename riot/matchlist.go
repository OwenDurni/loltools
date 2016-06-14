package riot

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// matchlist-v2.2: https://developer.riotgames.com/api/methods#!/1069
type MatchList struct {
	EndIndex   int               `json:"endIndex"`
	Matches    []*MatchReference `json:"matches'`
	StartIndex int               `json:"startIndex"`
	TotalGames int               `json:"totalGames"`
}

// matchlist-v2.2: https://developer.riotgames.com/api/methods#!/1069
type MatchReference struct {
	Champion   int64    `json:"champion"`
	Lane       string   `json:"lane"`
	MatchId    int64    `json:"matchId"`
	PlatformId string   `json:"platformId"`
	Queue      string   `json:"queue"`
	Region     string   `json:"region"`
	Role       string   `json:"role"`
	Season     string   `json:"season"`
	Timestamp  RiotTime `json:"timestamp"`
}

func RankedGameHistoryBySummonerIdSince(
	urlFetcher func(string) ([]byte, error),
	rateLimiter func(),
	riotApiKey string,
	region string,
	summonerId int64,
	startDateTime RiotTime) (*MatchList, error) {
	loc := ComposeUrl(
		riotApiKey,
		fmt.Sprintf("/api/lol/%s/v2.2/matchlist/by-summoner/%d",
			region, summonerId),
		&url.Values{
			"beginTime": []string{startDateTime.UnixMillisString()},
		})
	rateLimiter()
	jsonData, err := urlFetcher(loc)
	if err != nil {
		return nil, err
	}
	mlist := new(MatchList)
	err = json.Unmarshal(jsonData, &mlist)
	return mlist, err
}
