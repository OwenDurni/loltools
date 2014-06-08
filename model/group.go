package model

import (
  "appengine"
  "appengine/datastore"
  "fmt"
)

// Root entity for all groups and acls.
type GroupRoot struct{}

func GroupRootKey(c appengine.Context) *datastore.Key {
  return datastore.NewKey(c, "GroupRoot", "dev", 0, nil)
}

type Group struct {
  Name string
}

type GroupMembership struct {
  GroupKey *datastore.Key
  UserKey  *datastore.Key
  Owner    bool
}

type ProposedGroupMembership struct {
  GroupKey *datastore.Key
  UserKey  *datastore.Key
  Notes    string
}

func GroupId(groupKey *datastore.Key) string {
  return EncodeKeyShort(groupKey)
}

func GroupUri(groupKey *datastore.Key) string {
  return fmt.Sprintf("/groups/%s", GroupId(groupKey))
}

func GroupKeyById(c appengine.Context, groupId string) (*datastore.Key, error) {
  return DecodeKeyShort(c, "Group", groupId, GroupRootKey(c))
}

func GroupByKey(
  c appengine.Context,
  groupKey *datastore.Key,
  userKey *datastore.Key) (*Group, *GroupMembership, error) {
  q := datastore.NewQuery("GroupMembership").Ancestor(GroupRootKey(c)).
    Filter("GroupKey =", groupKey).
    Filter("UserKey =", userKey).
    Limit(1)
  var memberships []*GroupMembership
  _, err := q.GetAll(c, &memberships)
  if err != nil {
    return nil, nil, err
  }
  if len(memberships) == 0 {
    return nil, nil, ErrNotAuthorized{PermissionView, groupKey}
  }

  group := new(Group)
  if err := datastore.Get(c, groupKey, group); err != nil {
    return nil, nil, err
  }

  return group, memberships[0], nil
}

func GroupById(
  c appengine.Context,
  userKey *datastore.Key,
  groupId string) (*Group, *datastore.Key, *GroupMembership, error) {
  groupKey, err := GroupKeyById(c, groupId)
  if err != nil {
    return nil, nil, nil, err
  }
  group, membership, err := GroupByKey(c, groupKey, userKey)
  return group, groupKey, membership, err
}

func GetGroupsForUser(
  c appengine.Context, userKey *datastore.Key) ([]*Group, []*GroupMembership, error) {
  var groups []*Group
  var memberships []*GroupMembership

  err := datastore.RunInTransaction(c, func(c appengine.Context) error {
    q := datastore.NewQuery("GroupMembership").Ancestor(GroupRootKey(c)).
      Filter("UserKey =", userKey)
    _, err := q.GetAll(c, &memberships)
    if err != nil {
      return err
    }

    groupKeys := make([]*datastore.Key, len(memberships))
    groups = make([]*Group, len(memberships))

    for i, m := range memberships {
      groupKeys[i] = m.GroupKey
      groups[i] = new(Group)
    }

    return datastore.GetMulti(c, groupKeys, groups)
  }, nil)
  if err != nil {
    return nil, nil, err
  }
  return groups, memberships, nil
}

func GetGroupMemberships(
  c appengine.Context, groupKey *datastore.Key) ([]*GroupMembership, error) {
  q := datastore.NewQuery("GroupMembership").Ancestor(GroupRootKey(c)).
    Filter("GroupKey =", groupKey)

  var memberships []*GroupMembership
  _, err := q.GetAll(c, &memberships)
  return memberships, err
}

func GetProposedGroupMemberships(
  c appengine.Context, groupKey *datastore.Key) ([]*ProposedGroupMembership, error) {
  q := datastore.NewQuery("ProposedGroupMembership").Ancestor(GroupRootKey(c)).
    Filter("GroupKey =", groupKey)

  var memberships []*ProposedGroupMembership
  _, err := q.GetAll(c, &memberships)
  return memberships, err
}

// Creates a new group with the current user as the owner.
func CreateGroup(c appengine.Context, name string) (*Group, *datastore.Key, error) {
  _, userKey, err := GetUser(c)
  if err != nil {
    return nil, nil, err
  }

  groot := GroupRootKey(c)

  group := new(Group)
  group.Name = name
  var groupKey *datastore.Key

  err = datastore.RunInTransaction(c, func(c appengine.Context) error {
    groupKey, err = datastore.Put(c, datastore.NewIncompleteKey(c, "Group", groot), group)
    if err != nil {
      return err
    }

    groupMembership := &GroupMembership{
      GroupKey: groupKey,
      UserKey:  userKey,
      Owner:    true,
    }
    _, err = datastore.Put(c, datastore.NewIncompleteKey(c, "GroupMembership", groot),
      groupMembership)
    return err
  }, nil)
  if err != nil {
    return nil, nil, err
  }
  return group, groupKey, nil
}

func GroupAddProposedMember(
  c appengine.Context, groupKey *datastore.Key, userKey *datastore.Key, notes string) error {
  groot := GroupRootKey(c)
  return datastore.RunInTransaction(c, func(c appengine.Context) error {
    // If the user is already a member, this operation is a no-op.
    q := datastore.NewQuery("GroupMembership").Ancestor(groot).
      Filter("GroupKey =", groupKey).
      Filter("UserKey =", userKey).
      Limit(1).
      KeysOnly()
    membershipKeys, err := q.GetAll(c, nil)
    if err != nil {
      return err
    }
    if len(membershipKeys) > 0 {
      return nil
    }
    
    // If the user is already a proposed member, update the entry.
    q = datastore.NewQuery("ProposedGroupMembership").Ancestor(groot).
      Filter("GroupKey =", groupKey).
      Filter("UserKey =", userKey).
      Limit(1).
      KeysOnly()
    proposedMembershipKeys, err := q.GetAll(c, nil)
    if err != nil {
      return err
    }
    proposedMembership := new(ProposedGroupMembership)
    var key *datastore.Key
    if len(proposedMembershipKeys) > 0 {
      key = proposedMembershipKeys[0]
      err = datastore.Get(c, key, proposedMembership)
      if err != nil {
        return err
      }
    } else {
      key = datastore.NewIncompleteKey(c, "ProposedGroupMembership", groot)  
    }
    
    proposedMembership.GroupKey = groupKey
    proposedMembership.UserKey = userKey
    proposedMembership.Notes = notes
    _, err = datastore.Put(c, key, proposedMembership)
    return err
  }, nil)
}

func GroupAddMember(
  c appengine.Context, groupKey *datastore.Key, userKey *datastore.Key, owner bool) error {
  groot := GroupRootKey(c)
  return datastore.RunInTransaction(c, func(c appengine.Context) error {
    // If there is a proposed membership for this user, delete it.
    q := datastore.NewQuery("ProposedGroupMembership").Ancestor(groot).
      Filter("GroupKey =", groupKey).
      Filter("UserKey =", userKey).
      Limit(1).
      KeysOnly()
    proposedMembershipKeys, err := q.GetAll(c, nil)
    if err != nil {
      return err
    }
    if len(proposedMembershipKeys) > 0 {
      err = datastore.DeleteMulti(c, proposedMembershipKeys)
      if err != nil {
        return err
      }
    }
    
    // Add the membership.
    q = datastore.NewQuery("GroupMembership").Ancestor(groot).
      Filter("GroupKey =", groupKey).
      Filter("UserKey =", userKey).
      Limit(1).
      KeysOnly()
    membershipKeys, err := q.GetAll(c, nil)
    if err != nil {
      return err
    }
    if len(membershipKeys) > 0 {
      return nil
    }

    membership := new(GroupMembership)
    membership.GroupKey = groupKey
    membership.UserKey = userKey
    membership.Owner = owner
    key := datastore.NewIncompleteKey(c, "GroupMembership", groot)
    _, err = datastore.Put(c, key, membership)
    return err
  }, nil)
}

func GroupDelMember(
  c appengine.Context, groupKey *datastore.Key, userKey *datastore.Key) error {
  groot := GroupRootKey(c)
  return datastore.RunInTransaction(c, func(c appengine.Context) error {
    // Remove proposed memberships
    q := datastore.NewQuery("ProposedGroupMembership").Ancestor(groot).
      Filter("GroupKey =", groupKey).
      Filter("UserKey =", userKey).
      KeysOnly()
    proposedMembershipKeys, err := q.GetAll(c, nil)
    if err != nil {
      return err
    }
    err = datastore.DeleteMulti(c, proposedMembershipKeys)
    if err != nil {
      return err
    }
    
    // Remove actual memberships
    q = datastore.NewQuery("GroupMembership").Ancestor(groot).
      Filter("GroupKey =", groupKey).
      Filter("UserKey =", userKey).
      KeysOnly()
    membershipKeys, err := q.GetAll(c, nil)
    if err != nil {
      return err
    }
    return datastore.DeleteMulti(c, membershipKeys)
  }, nil)
}
