package model

import (
  "appengine"
  "appengine/datastore"
)

type Group struct {
  Name string
}

type GroupMembership struct {
  GroupKey *datastore.Key
  UserKey  *datastore.Key
}

func GetGroupKeysForUser(c appengine.Context, userKey *datastore.Key) ([]*datastore.Key, error) {
  q := datastore.NewQuery("GroupMembership").
    Filter("UserKey =", userKey).
    KeysOnly()
  var out []*datastore.Key
  _, err := q.GetAll(c, &out)
  return out, err
}
