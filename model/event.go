package model

import (
  "appengine/datastore"
  "time"
)

type Event struct {
  Name      string
  StartTime time.Time
  EndTime   time.Time
}

type EventGroupMembership struct {
  EventKey *datastore.Key
  GroupKey *datastore.Key
}

type EventUserRsvp struct {
  EventKey          *datastore.Key
  UserKey           *datastore.Key
  AttendencePercent int
  StartTime         time.Time
  EndTime           time.Time
}
