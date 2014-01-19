package view

import (
  "appengine"
  "appengine/datastore"
  "github.com/OwenDurni/loltools/model"
  "net/http"
)

type Group struct {
  Name string
  Id string
  Uri string
}
func (g *Group) Fill(m *model.Group, key *datastore.Key) *Group {
  g.Name = m.Name
  g.Id = model.GroupId(key)
  g.Uri = model.GroupUri(key)
  return g
}

func GroupIndexHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)

  // Lookup data from backend.
  _, userKey, err := model.GetUser(c)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  memberships, err := model.GetGroupsForUser(c, userKey)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  
  // Populate view context.
  ctx := struct {
    ctxBase
    OwnedGroups []*Group
    MemberGroups []*Group
  }{}
  ctx.ctxBase.init(c)
  ctx.ctxBase.Title = "loltools - My Groups"

  for _, m := range memberships {
    g, err := model.GetGroupByKey(c, m.GroupKey)  // TODO: GetMulti
    if err != nil {
      ctx.Errors = append(ctx.Errors, err)
      continue
    }
    vg := new(Group).Fill(g, m.GroupKey)
    if m.Owner {
      ctx.OwnedGroups = append(ctx.OwnedGroups, vg)
    } else {
      ctx.MemberGroups = append(ctx.MemberGroups, vg)
    }
  }

  // Render
  if err := RenderTemplate(w, "groups/index.html", "base", ctx); err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
}

func ApiGroupCreateHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  _, groupKey, err := model.CreateGroup(c, r.FormValue("name"))
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  HttpReplyResourceCreated(w, model.GroupUri(groupKey))
}