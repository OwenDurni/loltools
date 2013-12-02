package view

import (
  "appengine"
  "appengine/user"
  "bytes"
  "html/template"
  "time"
)

type commonCtx struct {
  Title       string
  TimeNow     string
  ContentHTML template.HTML
  User        string
}

func (ctx *commonCtx) init(c appengine.Context) *commonCtx {
  ctx.Title = ""
  ctx.TimeNow = fmtTime(time.Now(), "America/Los_Angeles")
  ctx.ContentHTML = template.HTML("")
  if u := user.Current(c); u != nil {
    ctx.User = u.Email
  }
  return ctx
}

func parseTemplate(file string, data interface{}) (out []byte, error error) {
  var buf bytes.Buffer
  t, err := template.ParseFiles(file)
  if err != nil {
    return nil, err
  }
  err = t.Execute(&buf, data)
  if err != nil {
    return nil, err
  }
  return buf.Bytes(), nil
}
