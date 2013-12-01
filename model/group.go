package model

import (
  "appengine/datastore"
)

const DATASTORE_GROUP_KIND = "Group"

type Group struct {
  Name string
}

const DATASTORE_GROUP_MEMBERSHIP_KIND = "GroupMembership"

type GroupMembership struct {
  GroupKey *datastore.Key
  UserKey *datastore.Key
}