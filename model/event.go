package model

import (
  "appengine"
  "appengine/datastore"
  "time"
)

// datastore kind
type Event struct {
  Name      string
  StartTime time.Time
  EndTime   time.Time
}

// datastore kind
type EventGroupMembership struct {
  EventKey *datastore.Key
  GroupKey *datastore.Key
}

// datastore kind
type EventUserRsvp struct {
  EventKey          *datastore.Key
  UserKey           *datastore.Key
  AttendencePercent int
  StartTime         time.Time
  EndTime           time.Time
}

// local structs
type EventList struct {
  Events map[*datastore.Key]*Event
}

func (e *EventList) Init() *EventList {
  e.Events = make(map[*datastore.Key]*Event)
  return e
}
func (e *EventList) AppendFutureEventsForGroup(
  c appengine.Context,
  groupKey *datastore.Key) (err error) {
  q := datastore.NewQuery("EventGroupMembership").
    Filter("GroupKey =", groupKey).
    Filter("EndTime >", time.Now().UTC())
  var eventMemberships []EventGroupMembership
  if _, err = q.GetAll(c, &eventMemberships); err != nil {
    return
  }
  for _, eventMembership := range eventMemberships {
    eventKey := eventMembership.EventKey
    if event, exists := e.Events[eventKey]; !exists {
      // Lookup this event and add it to the EventList
      event = new(Event)
      if err = datastore.Get(c, eventKey, &event); err != nil {
        return
      }
      e.Events[eventKey] = event
    } else {
      // We've already looked up this event; nothing to do
    }
  }
  return
}
