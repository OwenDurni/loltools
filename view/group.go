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
	Id   string
	Uri  string
}

func (g *Group) Fill(m *model.Group, key *datastore.Key) *Group {
	if m != nil {
		g.Name = m.Name
	}
	g.Id = model.GroupId(key)
	g.Uri = model.GroupUri(key)
	return g
}

type GroupAcl struct {
	Group   *Group
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

type ProposedMember struct {
	Email string
	Notes string
}

func (m *ProposedMember) Fill(u *model.User, proposal *model.ProposedGroupMembership) *ProposedMember {
	m.Email = u.Email
	m.Notes = proposal.Notes
	return m
}

func GroupIndexHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
	c := appengine.NewContext(r)

	// Lookup data from backend.
	user, userKey, err := model.GetUser(c)
	if HandleError(c, w, err) {
		return
	}

	groups, memberships, err := model.GetGroupsForUser(c, userKey)
	if HandleError(c, w, err) {
		return
	}

	// Populate view context.
	ctx := struct {
		ctxBase
		OwnedGroups  []*Group
		MemberGroups []*Group
	}{}
	ctx.ctxBase.init(c, user)
	ctx.ctxBase.Title = "loltools > My Groups"

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
	if HandleError(c, w, err) {
		return
	}
}

func GroupViewHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
	c := appengine.NewContext(r)
	groupId := args["groupId"]

	user, userKey, err := model.GetUser(c)
	if HandleError(c, w, err) {
		return
	}

	group, groupKey, _, err := model.GroupById(c, userKey, groupId)
	switch e := err.(type) {
	case model.ErrNotAuthorized:
		GroupViewNotAuthorizedHandler(w, c, user, e.Resource)
		return
	default:
		if HandleError(c, w, e) {
			return
		}
	}

	memberships, err := model.GetGroupMemberships(c, groupKey)
	if HandleError(c, w, err) {
		return
	}

	proposedMemberships, err := model.GetProposedGroupMemberships(c, groupKey)
	if HandleError(c, w, err) {
		return
	}

	ctx := struct {
		ctxBase
		Group
		Members         []*Member
		ProposedMembers []*ProposedMember
	}{}
	ctx.ctxBase.init(c, user)
	ctx.ctxBase.Title = fmt.Sprintf("loltools > %s", group.Name)
	ctx.Group.Fill(group, groupKey)
	ctx.Members = make([]*Member, 0, len(memberships))
	ctx.ProposedMembers = make([]*ProposedMember, 0, len(proposedMemberships))

	for _, m := range memberships {
		user, err := model.GetUserByKey(c, m.UserKey)
		if err != nil {
			ctx.ctxBase.Errors = append(ctx.ctxBase.Errors, err)
			continue
		}
		ctx.Members = append(ctx.Members, new(Member).Fill(user, m))
	}

	for _, m := range proposedMemberships {
		user, err := model.GetUserByKey(c, m.UserKey)
		if err != nil {
			ctx.ctxBase.Errors = append(ctx.ctxBase.Errors, err)
			continue
		}
		ctx.ProposedMembers = append(ctx.ProposedMembers, new(ProposedMember).Fill(user, m))
	}

	// Render
	err = RenderTemplate(w, "groups/view.html", "base", ctx)
	if HandleError(c, w, err) {
		return
	}
}
func GroupViewNotAuthorizedHandler(
	w http.ResponseWriter, c appengine.Context, user *model.User, groupKey *datastore.Key) {

	ctx := struct {
		ctxBase
		Group
	}{}
	ctx.ctxBase.init(c, user)
	ctx.ctxBase.Title = fmt.Sprintf("loltools > group %s", model.GroupId(groupKey))
	ctx.Group.Fill(nil, groupKey)

	err := RenderTemplate(w, "groups/join.html", "base", ctx)
	if HandleError(c, w, err) {
		return
	}
}

func ApiGroupCreateHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
	c := appengine.NewContext(r)
	_, groupKey, err := model.CreateGroup(c, r.FormValue("name"))
	if ApiHandleError(c, w, err) {
		return
	}

	HttpReplyResourceCreated(w, model.GroupUri(groupKey))
}

func ApiGroupJoinHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
	c := appengine.NewContext(r)
	groupId := r.FormValue("group")
	notes := r.FormValue("notes")

	groupKey, err := model.GroupKeyById(c, groupId)
	if ApiHandleError(c, w, err) {
		return
	}

	_, userKey, err := model.GetUser(c)
	if ApiHandleError(c, w, err) {
		return
	}

	err = model.GroupAddProposedMember(c, groupKey, userKey, notes)
	if ApiHandleError(c, w, err) {
		return
	}

	HttpReplyOkEmpty(w)
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
	if ApiHandleError(c, w, err) {
		return
	}

	_, groupKey, userMembership, err := model.GroupById(c, userKey, groupId)
	if ApiHandleError(c, w, err) {
		return
	}

	// Only owners of a group can add members.
	if !userMembership.Owner {
		HttpReplyError(c, w, http.StatusForbidden, false,
			errors.New("Can only add members to a group you own."))
		return
	}

	_, addUserKey, err := model.GetUserByEmail(c, addUserEmail)
	if ApiHandleError(c, w, err) {
		return
	}

	err = model.GroupAddMember(c, groupKey, addUserKey, owner)
	if ApiHandleError(c, w, err) {
		return
	}

	HttpReplyOkEmpty(w)
}

func ApiGroupDelUserHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
	c := appengine.NewContext(r)
	groupId := r.FormValue("group")
	delUserEmail := r.FormValue("email")

	_, userKey, err := model.GetUser(c)
	if ApiHandleError(c, w, err) {
		return
	}

	_, groupKey, userMembership, err := model.GroupById(c, userKey, groupId)
	if ApiHandleError(c, w, err) {
		return
	}

	// Only owners of a group can add members.
	if !userMembership.Owner {
		HttpReplyError(c, w, http.StatusForbidden, false,
			errors.New("Can only remove members to a group you own."))
		return
	}

	_, delUserKey, err := model.GetUserByEmail(c, delUserEmail)
	if ApiHandleError(c, w, err) {
		return
	}

	err = model.GroupDelMember(c, groupKey, delUserKey)
	if ApiHandleError(c, w, err) {
		return
	}

	HttpReplyOkEmpty(w)
}
