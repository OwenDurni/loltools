package view

import (
  "bytes"
  "html/template"
)

func ParseTemplate(file string, data interface{}) (out []byte, error error) {
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
