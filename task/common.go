package task

import (
  "appengine"
  "github.com/OwenDurni/loltools/model"
  "github.com/OwenDurni/loltools/view"
  "net/http"
)

func ReportError(c appengine.Context, w http.ResponseWriter, err error) bool {
  if err == nil { return false }
  
  shouldRetry := false
  
  if _, ok := err.(model.ErrRateLimitExceeded); ok {
    shouldRetry = true
  }
  
  if shouldRetry {
    // We write a non-2XX response so that the task is retried.
    c.Warningf("[Temporary Task Error] %v", err)
    view.HttpReplyError(c, w, http.StatusInternalServerError, err)
  } else {
    // We write a 200 response so that the task is not retried.
    c.Errorf("[Permanent Task Error] %v", err)
    view.HttpReplyOkEmpty(w)
  }
  return true
}