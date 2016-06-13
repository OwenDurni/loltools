package main

import (
	"fmt"
	"github.com/OwenDurni/loltools/riot"
	"github.com/OwenDurni/loltools/util/web"
	"golang.org/x/net/context"
	"golang.org/x/time/rate"
	"io/ioutil"
	"os"
)

const (
	Region = "na"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Usage:
//   go run tools/ranked_data.go [summoner...]
func main() {
	var riotApiKey string
	{
		contents, err := ioutil.ReadFile("riot-api-key")
		check(err)
		riotApiKey = string(contents)
	}
	fmt.Fprintf(os.Stderr, "Using riot key: %s\n", riotApiKey)

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
		for _, championStats := range rankedStats.Champions {
		  if championStats.ChampionId == riot.ChampionStatsDto_AllChampions {
		    fmt.Printf("%s,%d\n", summoner, championStats.Stats.TotalSessionsPlayed)
		  }
		}
	}
}
