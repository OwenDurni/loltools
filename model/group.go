package model

import (
  "appengine"
  "appengine/datastore"
  "errors"
)

type Group struct {
  Name string
}

type GroupMembership struct {
  GroupKey *datastore.Key
  UserKey  *datastore.Key
  Owner    bool
}

func GetGroupKeysForUser(
  c appengine.Context,
  userKey *datastore.Key,
  out *[]*datastore.Key) (err error) {
  q := datastore.NewQuery("GroupMembership").
    Filter("UserKey =", userKey).
    KeysOnly()
  _, err = q.GetAll(c, out)
  return
}

// Creates a new group with the current user as the owner.
func CreateGroup(c appengine.Context) (*Group, *datastore.Key, error) {
  _, userKey, err := GetUser(c)
  if err != nil {
    return nil, nil, err
  }

  group := new(Group)
  groupKey, err := datastore.Put(
    c, datastore.NewIncompleteKey(c, "Group", nil), group)
  if err != nil {
    return nil, nil, err
  }

  groupMembership := &GroupMembership{
    GroupKey: groupKey,
    UserKey:  userKey,
    Owner:    true,
  }
  _, err = datastore.Put(
    c,
    datastore.NewIncompleteKey(c, "GroupMembership", nil),
    groupMembership)

  return group, groupKey, err
}

func DeleteGroup(c appengine.Context, groupKey *datastore.Key) error {
  //user, userKey, err := GetUser(c)

  // TODO: This should be a transaction.

  // Error if this user is not an owner of the group.
  //q := datastore.NewQuery("GroupMembership").
  //  Filter("GroupKey =", groupKey).
  //  Filter("UserKey =", userKey).
  //  Filter("Owner =", true).
  //  KeysOnly()

  return errors.New("Not yet implemented")
}
