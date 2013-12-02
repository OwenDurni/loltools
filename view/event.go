package view

import (
  "appengine"
  "appengine/datastore"
  "github.com/OwenDurni/loltools/model"
  "net/http"
)

type ContextEvent struct {
  Name      string
  StartTime string
  EndTime   string
}

func (c *ContextEvent) Init(e *model.Event) {
  c.Name = e.Name
  c.StartTime = e.StartTime.Format(TIME_FORMAT)
  c.EndTime = e.EndTime.Format(TIME_FORMAT)
}

type ContextEventList struct {
  Events []model.Event
}

func EventList(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)

  // Get current user.
  _, userKey, err := model.GetUser(c)
  if err != nil {
    httpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  // Get all keys for groups this user is a member of.
  var groupKeys []*datastore.Key
  if err := model.GetGroupKeysForUser(c, userKey, &groupKeys); err != nil {
    httpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }

  // Get all events this user is invited to that have not elapsed.
  eventList := new(model.EventList).Init()
  for _, groupKey := range groupKeys {
    if err := eventList.AppendFutureEventsForGroup(c, groupKey); err != nil {
      httpReplyError(w, r, http.StatusInternalServerError, err)
      return
    }
  }

  // TODO: Sort events by start time.
}
