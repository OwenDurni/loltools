package view

import (
  "appengine"
  "appengine/datastore"
  "errors"
  "fmt"
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

type GroupAcl struct {
  Group *Group
  CanView bool
  CanEdit bool
}
func (ga *GroupAcl) Fill(group *Group, perms map[model.Permission]bool) *GroupAcl {
  ga.Group = group
  ga.CanView = perms[model.PermissionView]
  ga.CanEdit = perms[model.PermissionEdit]
  return ga
}

type Member struct {
  Email string
  Owner bool
}
func (m *Member) Fill(u *model.User, membership *model.GroupMembership) *Member {
  m.Email = u.Email
  m.Owner = membership.Owner
  return m
}

func GroupIndexHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)

  // Lookup data from backend.
  _, userKey, err := model.GetUser(c)
  if HandleError(c, w, err) { return }

  groups, memberships, err := model.GetGroupsForUser(c, userKey)
  if HandleError(c, w, err) { return }
  
  // Populate view context.
  ctx := struct {
    ctxBase
    OwnedGroups []*Group
    MemberGroups []*Group
  }{}
  ctx.ctxBase.init(c)
  ctx.ctxBase.Title = "loltools - My Groups"

  for i, m := range memberships {
    g := groups[i]
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
  err = RenderTemplate(w, "groups/index.html", "base", ctx)
  if HandleError(c, w, err) { return }
}

func GroupViewHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  groupId := args["groupId"]
  
  _, userKey, err := model.GetUser(c)
  if HandleError(c, w, err) { return }
  
  group, groupKey, _, err := model.GroupById(c, userKey, groupId)
  if HandleError(c, w, err) { return }
  
  memberships, err := model.GetGroupMemberships(c, groupKey)
  if HandleError(c, w, err) { return }
  
  ctx := struct {
    ctxBase
    Group
    Members []*Member
  }{}
  ctx.ctxBase.init(c)
  ctx.ctxBase.Title = fmt.Sprintf("loltools - %s", group.Name)
  ctx.Group.Fill(group, groupKey)
  ctx.Members = make([]*Member, 0, len(memberships))
  
  for _, m := range memberships {
    user, err := model.GetUserByKey(c, m.UserKey)
    if err != nil {
      ctx.ctxBase.Errors = append(ctx.ctxBase.Errors, err)
      continue
    }
    ctx.Members = append(ctx.Members, new(Member).Fill(user, m))
  }
  
  // Render
  err = RenderTemplate(w, "groups/view.html", "base", ctx)
  if HandleError(c, w, err) { return }
}

func ApiGroupCreateHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  _, groupKey, err := model.CreateGroup(c, r.FormValue("name"))
  if ApiHandleError(c, w, err) { return }
  
  HttpReplyResourceCreated(w, model.GroupUri(groupKey))
}

func ApiGroupAddUserHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  groupId := r.FormValue("group")
  addUserEmail := r.FormValue("email")
  owner := false
  if r.FormValue("owner") == "1" {
    owner = true
  }

  _, userKey, err := model.GetUser(c)
  if ApiHandleError(c, w, err) { return }

  _, groupKey, userMembership, err := model.GroupById(c, userKey, groupId)
  if ApiHandleError(c, w, err) { return }

  // Only owners of a group can add members.
  if !userMembership.Owner {
    HttpReplyError(c, w, http.StatusForbidden, false,
                   errors.New("Can only add members to a group you own."))
    return
  }
  
  _, addUserKey, err := model.GetUserByEmail(c, addUserEmail)
  if ApiHandleError(c, w, err) { return }
  
  err = model.GroupAddMember(c, groupKey, addUserKey, owner)
  if ApiHandleError(c, w, err) { return }
  
  HttpReplyOkEmpty(w)
}

func ApiGroupDelUserHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  groupId := r.FormValue("group")
  delUserEmail := r.FormValue("email")

  _, userKey, err := model.GetUser(c)
  if ApiHandleError(c, w, err) { return }

  _, groupKey, userMembership, err := model.GroupById(c, userKey, groupId)
  if ApiHandleError(c, w, err) { return }

  // Only owners of a group can add members.
  if !userMembership.Owner {
    HttpReplyError(c, w, http.StatusForbidden, false,
                   errors.New("Can only remove members to a group you own."))
    return
  }
  
  _, delUserKey, err := model.GetUserByEmail(c, delUserEmail)
  if ApiHandleError(c, w, err) { return }
  
  err = model.GroupDelMember(c, groupKey, delUserKey)
  if ApiHandleError(c, w, err) { return }
  
  HttpReplyOkEmpty(w)
}