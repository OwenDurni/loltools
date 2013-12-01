package view

import (
  "bytes"
  "html/template"
)

type commonCtx struct {
  Title       string
  ContentHTML template.HTML
}

func (ctx *commonCtx) init() {
  ctx.Title = ""
  ctx.ContentHTML = template.HTML("")
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
