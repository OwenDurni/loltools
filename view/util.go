package view

import (
  "appengine"
  "github.com/OwenDurni/loltools/model"
  "net/http"
  "time"
)

const TIME_FORMAT = "2006-01-02 03:04PM (MST)"

// loc is the IANA Time Zone location (ex: "America/New_York")
// If the string is malformed the time is returned in UTC.
func fmtTime(t time.Time, loc string) string {
  if location, err := time.LoadLocation(loc); err != nil {
    t = t.UTC()
  } else {
    t = t.In(location)
  }
  return t.Format(TIME_FORMAT)
}

func HttpReplyOkEmpty(w http.ResponseWriter) {
  w.WriteHeader(http.StatusNoContent)
}

func HttpReplyResourceCreated(w http.ResponseWriter, loc string) {
  w.Header().Add("Location", loc)
  w.WriteHeader(http.StatusCreated)
}

func HandleError(c appengine.Context, w http.ResponseWriter, err error) bool {
  if err == nil { return false }
  if _, ok := err.(model.ErrNotAuthorized); ok {
    HttpReplyError(c, w, http.StatusForbidden, err)
    return true
  }
  HttpReplyError(c, w, http.StatusInternalServerError, err)
  return true
}

// See http://golang.org/pkg/net/http/#pkg-constants for status codes.
func HttpReplyError(
  c appengine.Context,
  w http.ResponseWriter,
  httpStatusCode int,
  err error) {
  errorString := ""
  if err != nil {
    errorString = err.Error()
  }

  // Log if this was a server-side error
  if 500 <= httpStatusCode && httpStatusCode < 600 {
    c.Errorf("%d: %s", httpStatusCode, errorString)
  }

  http.Error(w, errorString, httpStatusCode)
}
