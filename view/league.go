package view

import (
  "appengine"
  "appengine/datastore"
  "github.com/OwenDurni/loltools/model"
  "net/http"
)

type League struct {
  Name string
  Owner string
  Id string
  Uri string
}
func (l *League) Fill(m model.League, k *datastore.Key) *League {
  l.Name = m.Name
  l.Id = model.EncodeGlobalKeyShort(k)
  l.Uri = model.LeagueUri(k)
  return l
}

func LeagueIndexHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)

  // Lookup data from backend.
  _, userKey, err := model.GetUser(c)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  
  leagueInfos, err := model.LeaguesForUser(c, userKey)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  
  // Populate view context.
  ctx := struct {
    ctxBase
    MyLeagues []*League
  }{}
  ctx.ctxBase.init(c)
  
  ctx.MyLeagues = make([]*League, len(leagueInfos))
  for i, info := range leagueInfos {
    league := new(League).Fill(info.League, info.LeagueKey)
    if owner, err := model.GetUserByKey(c, info.League.Owner); err == nil {
      league.Owner = owner.Email
    } else {
      league.Owner = err.Error()
    }
    ctx.MyLeagues[i] = league
  }
  
  // Render
  if err := RenderTemplate(w, "leagues/index.html", "base", ctx); err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
}

func LeagueCreateHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)

  _, _, err := model.GetUser(c)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  ctx := struct {
    ctxBase
  }{}
  ctx.ctxBase.init(c)
  ctx.ctxBase.Title = "Create League"

  if err := RenderTemplate(w, "leagues/create.html", "base", ctx); err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
}

func ApiLeagueCreateHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  _, leagueKey, err := model.CreateLeague(c, r.FormValue("name"))
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  HttpReplyResourceCreated(w, model.LeagueUri(leagueKey))
}