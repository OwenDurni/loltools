package riot

import (
	"net/url"
)

const (
	BlueTeamId   = 100
	PurpleTeamId = 200
)

const (
	baseUrl = "https://na.api.pvp.net"
)

type NotFound struct{}

func (e NotFound) Error() string {
	return "Riot API 404"
}

func BaseUrl() (url *url.URL) {
	url, err := url.Parse(baseUrl)
	if err != nil {
		panic(err)
	}
	return
}

func ComposeUrl(riotApiKey string, path string, args *url.Values) string {
	u := BaseUrl()
	u.Path += path
	args.Add("api_key", riotApiKey)
	u.RawQuery = args.Encode()
	return u.String()
}
