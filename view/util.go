package view

import (
  "appengine"
  "fmt"
  "github.com/OwenDurni/loltools/model"
  "net/http"
  "strings"
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

// date: The form value sent by an HTML input with type="date"
// time: The form value sent by an HTML input with type="time" step="1"
// tz: The IANA timezone string corresponding to the given date (ex: "America/Los_Angeles")
func parseDatetime(datestr string, timestr string, tz string) (time.Time, error) {
  // Pad a truncated time string
  for strings.Count(timestr, ":") < 2 {
    timestr += ":00"
  }
  
  location, err := time.LoadLocation(tz)
  if err != nil {
    return time.Time{}, err
  }
  return time.ParseInLocation(
    "2006-01-02T15:04:05", fmt.Sprintf("%sT%s", datestr, timestr), location)
}

func HttpReplyOkEmpty(w http.ResponseWriter) {
  w.WriteHeader(http.StatusNoContent)
}

func HttpReplyResourceCreated(w http.ResponseWriter, loc string) {
  w.Header().Add("Location", loc)
  w.WriteHeader(http.StatusCreated)
}

func ApiHandleError(c appengine.Context, w http.ResponseWriter, errs ...error) bool {
  useTemplate := false
  for _, err := range errs {
    if err == nil {
      continue
    }
    if _, ok := err.(model.ErrNotAuthorized); ok {
      HttpReplyError(c, w, http.StatusForbidden, useTemplate, err)
      return true
    }
    HttpReplyError(c, w, http.StatusInternalServerError, useTemplate, err)
    return true
  }
  return false
}

func HandleError(c appengine.Context, w http.ResponseWriter, errs ...error) bool {
  useTemplate := true
  for _, err := range errs {
    if err == nil {
      continue
    }
    if _, ok := err.(model.ErrNotAuthorized); ok {
      HttpReplyError(c, w, http.StatusForbidden, useTemplate, err)
      return true
    } else {
      HttpReplyError(c, w, http.StatusInternalServerError, useTemplate, err)
      return true
    }
  }
  return false
}

// See http://golang.org/pkg/net/http/#pkg-constants for status codes.
func HttpReplyError(
  c appengine.Context,
  w http.ResponseWriter,
  httpStatusCode int,
  useTemplate bool,
  err error) {

  errorString := ""
  if err != nil {
    errorString = fmt.Sprintf("%d: %s", httpStatusCode, err.Error())
  }

  // Log if this was a server-side error
  if 500 <= httpStatusCode && httpStatusCode < 600 {
    c.Errorf("%d: %s", httpStatusCode, errorString)
  }

  if !useTemplate {
    http.Error(w, err.Error(), httpStatusCode)
  } else {
    // Don't send an error code as some browsers won't render html for non-2XX responses.

    ctx := struct {
      ctxBase
      HttpStatusCode int
    }{}
    ctx.ctxBase.init(c)
    ctx.ctxBase.AddError(err)
    ctx.HttpStatusCode = httpStatusCode

    if tmplerr := RenderTemplate(w, "httperror.html", "base", ctx); tmplerr != nil {
      // Fallback to plain old response.
      http.Error(w, errorString, httpStatusCode)
    }
  }
}
