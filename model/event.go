package model

import (
  "appengine/datastore"
  "time"
)

const DATASTORE_KIND_EVENT = "Event"

type Event struct {
  Name string
  StartTime time.Time
  EndTime time.Time
}

const DATASTORE_KIND_EVENT_RSVP = "EventRsvp"

type EventRsvp struct {
  EventKey *datastore.Key
  UserKey *datastore.Key
  Responded bool
  AttendencePercent int
  StartTime time.Time
  EndTime time.Time
}
