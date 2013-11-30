package view

import (
  "appengine"
  "net/http"
)

func httpReplyOkEmpty(w http.ResponseWriter) {
  w.WriteHeader(http.StatusNoContent)
}

// See http://golang.org/pkg/net/http/#Constants for status codes.
func httpReplyError(
    w http.ResponseWriter,
    r *http.Request,
    httpStatusCode int,
    err error) {
  c := appengine.NewContext(r)

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
