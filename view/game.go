package view

import (
	"appengine"
	"fmt"
	"github.com/OwenDurni/loltools/model"
	"net/http"
)

func GameViewHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
	c := appengine.NewContext(r)
	gameId := args["gameId"]

	user, _, err := model.GetUser(c)
	if HandleError(c, w, err) {
		return
	}

	game, _, err := model.GameById(c, gameId)
	playerCache := model.NewPlayerCache(c, game.Region)
	gameInfo, errs := model.GetGameInfo(c, playerCache, game)
	if HandleError(c, w, errs...) {
		return
	}

	ctx := struct {
		ctxBase
		GameInfo *model.GameInfo
	}{}
	ctx.ctxBase.init(c, user)
	ctx.ctxBase.Title = fmt.Sprintf("loltools > %s", gameId)
	ctx.GameInfo = gameInfo

	err = RenderTemplate(w, "games/index.html", "base", ctx)
	if HandleError(c, w, err) {
		return
	}
}

func LeagueGameViewHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
	c := appengine.NewContext(r)
	leagueId := args["leagueId"]
	gameId := args["gameId"]

	user, userKey, err := model.GetUser(c)
	if HandleError(c, w, err) {
		return
	}
	userAcls := model.NewRequestorAclCache(userKey)

	league, leagueKey, err := model.LeagueById(c, leagueId)
	if HandleError(c, w, err) {
		return
	}

	game, gameKey, err := model.GameById(c, gameId)
	playerCache := model.NewPlayerCache(c, game.Region)
	gameInfo, errs := model.GetGameInfo(c, playerCache, game)
	if HandleError(c, w, errs...) {
		return
	}

	ctx := struct {
		ctxBase
		League
		GameInfo *model.GameInfo
		UserTags []*model.UserGameTag
		Tags     []*model.GameTag
	}{}
	ctx.ctxBase.init(c, user)
	ctx.ctxBase.Title = fmt.Sprintf("loltools > %s > %s", league.Name, gameId)
	ctx.League.Fill(league, leagueKey)
	ctx.GameInfo = gameInfo

	ctx.UserTags, _, err = model.GetUserGameTags(c, userKey, leagueKey, gameKey)
	if HandleError(c, w, err) {
		return
	}

	ctx.Tags, _, err = model.GetGameTags(c, userAcls, leagueKey, gameKey)
	if HandleError(c, w, err) {
		return
	}

	err = RenderTemplate(w, "leagues/games/index.html", "base", ctx)
	if HandleError(c, w, err) {
		return
	}
}
