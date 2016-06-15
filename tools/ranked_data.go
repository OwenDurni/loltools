package main

import (
	"fmt"
	"github.com/OwenDurni/loltools/riot"
	"github.com/OwenDurni/loltools/util/web"
	"golang.org/x/net/context"
	"golang.org/x/time/rate"
	"io/ioutil"
	"os"
	"time"
)

const (
	Region              = "na"
	GamesSinceStartDate = "2016-06-01T00:00:00Z"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Usage:
//   go run tools/ranked_data.go [summoner...] 2> /dev/null
func main() {
	var riotApiKey string
	{
		contents, err := ioutil.ReadFile("riot-api-key")
		check(err)
		riotApiKey = string(contents)
	}
	fmt.Fprintf(os.Stderr, "Using riot key: %s\n", riotApiKey)

	var gamesSinceStartDate riot.RiotTime
	{
		t, err := time.Parse(time.RFC3339, GamesSinceStartDate)
		check(err)
		gamesSinceStartDate = riot.RiotTime(t)
	}

	var rateLimiter func()
	{
		// Create a rate limiter that allows 1 call every 2 seconds with a burst of
		// 10. Wait 10 at the beginning so that consecutive invocations of this tool
		// do not exceed the rate limit (e.g. the bucket should start empty).
		ctx := context.Background()
		lim := rate.NewLimiter(0.5, 10)
		lim.WaitN(ctx, 10)
		rateLimiter = func() {
			lim.Wait(ctx)
		}
	}

	for _, arg := range os.Args[1:] {
		summoner := arg
		summonerData, err := riot.SummonerByName(
			web.FetchUrl, rateLimiter, riotApiKey, Region, summoner)
		check(err)

		summonerId := summonerData.Id

		rankedStats, err := riot.RankedStatsBySummonerId(
			web.FetchUrl, rateLimiter, riotApiKey, Region, summonerId)
		check(err)

		rankedGames, err := riot.RankedGameHistoryBySummonerIdSince(
			web.FetchUrl, rateLimiter, riotApiKey, Region, summonerId,
			gamesSinceStartDate)
		check(err)

		var sampleMatchId *int64 = nil
		if len(rankedGames.Matches) > 0 {
			sampleMatchId = &rankedGames.Matches[0].MatchId
		}

		soloRank, err := riot.SoloQueueRankBySummonerId(
			web.FetchUrl, rateLimiter, riotApiKey, Region, summonerId)
		check(err)

		previousSeasonRank := "Unranked"
		if sampleMatchId != nil {
			match, err := riot.LookupMatch(
				web.FetchUrl, rateLimiter, riotApiKey, Region, *sampleMatchId)
			check(err)
			for _, pid := range match.ParticipantIdentities {
				if pid.Player.SummonerId == summonerId {
					for _, p := range match.Participants {
						if pid.ParticipantId == p.ParticipantId {
							previousSeasonRank = p.HighestAchievedSeasonTier
						}
					}
				}
			}
		}

		totalRankedGames := 0
		totalGamesSinceDate := rankedGames.TotalGames
		for _, championStats := range rankedStats.Champions {
			if championStats.ChampionId == riot.ChampionStatsDto_AllChampions {
				totalRankedGames = championStats.Stats.TotalSessionsPlayed
			}
		}

		for _, rankedGame := range rankedGames.Matches {
			fmt.Fprintln(
				os.Stderr, time.Time(rankedGame.Timestamp).Format(time.RFC3339))
		}

		fmt.Printf("%s,%d,%d,%s,%s\n",
			summoner, totalRankedGames, totalGamesSinceDate, soloRank,
			previousSeasonRank)
	}
}
