package task

import (
  "appengine"
  "github.com/OwenDurni/loltools/view"
  "net/http"
)

func ReportPermanentError(c appengine.Context, w http.ResponseWriter, err error) {
  c.Errorf("[Permanent Task Error] %s", err.Error())
  
  // We write a 200 response so that the task is not retried.
  view.HttpReplyOkEmpty(w)
}

func ReportTemporaryError(w http.ResponseWriter, r *http.Request, httpStatusCode int, err error) {
  // This will cause the task to be retried.
  view.HttpReplyError(w, r, httpStatusCode, err)
}