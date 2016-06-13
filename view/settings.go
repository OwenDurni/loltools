package view

import (
	"appengine"
	"github.com/OwenDurni/loltools/model"
	"net/http"
	"strconv"
)

func SettingsIndexHandler(
	w http.ResponseWriter, r *http.Request, args map[string]string) {
	c := appengine.NewContext(r)

	user, userKey, err := model.GetUser(c)
	if HandleError(c, w, err) {
		return
	}

	ctx := struct {
		ctxBase
		Summoners []*model.SummonerData
	}{}
	ctx.ctxBase.init(c, user)

	summoners, err := model.GetSummonerDatas(c, userKey)
	if HandleError(c, w, err) {
		return
	}
	ctx.Summoners = summoners

	err = RenderTemplate(w, "settings/index.html", "base", ctx)
	if HandleError(c, w, err) {
		return
	}
}

func ApiUserAddSummoner(
	w http.ResponseWriter, r *http.Request, args map[string]string) {
	c := appengine.NewContext(r)
	region := r.FormValue("region")
	summoner := r.FormValue("summoner")

	_, userKey, err := model.GetUser(c)
	if ApiHandleError(c, w, err) {
		return
	}

	err = model.AddUnverifiedSummoner(c, userKey, region, summoner)
	if ApiHandleError(c, w, err) {
		return
	}

	HttpReplyOkEmpty(w)
}

func ApiUserVerifySummoner(
	w http.ResponseWriter, r *http.Request, args map[string]string) {
	c := appengine.NewContext(r)
	region := r.FormValue("region")
	summonerId, err := strconv.ParseInt(r.FormValue("summonerid"), 10, 64)
	if ApiHandleError(c, w, err) {
		return
	}

	_, userKey, err := model.GetUser(c)
	if ApiHandleError(c, w, err) {
		return
	}

	player, playerKey, err := model.GetOrCreatePlayerByRiotId(c, region, summonerId)
	if ApiHandleError(c, w, err) {
		return
	}

	err = model.VerifySummoner(c, userKey, playerKey, player)
	if ApiHandleError(c, w, err) {
		return
	}

	HttpReplyOkEmpty(w)
}

func ApiUserSetPrimarySummoner(
	w http.ResponseWriter, r *http.Request, args map[string]string) {
	c := appengine.NewContext(r)
	region := r.FormValue("region")
	summonerId, err := strconv.ParseInt(r.FormValue("summonerid"), 10, 64)
	if ApiHandleError(c, w, err) {
		return
	}

	_, userKey, err := model.GetUser(c)
	if ApiHandleError(c, w, err) {
		return
	}

	player, playerKey, err := model.GetOrCreatePlayerByRiotId(c, region, summonerId)
	if ApiHandleError(c, w, err) {
		return
	}

	err = model.SetPrimarySummoner(c, userKey, player, playerKey)
	if ApiHandleError(c, w, err) {
		return
	}

	HttpReplyOkEmpty(w)
}
