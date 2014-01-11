package view

import (
  "appengine"
  "appengine/user"
  "errors"
  "fmt"
  "html/template"
  "io/ioutil"
  "net/http"
  "time"
)

func tmplForm(id, uri, text string) interface{} {
  return struct {
    Id   string
    Uri  string
    Text string
  }{
    Id:   id,
    Uri:  uri,
    Text: text,
  }
}

var templateRegistry map[string]*template.Template
var tmplFuncRegistry = template.FuncMap{
  "form": tmplForm,
}

// Adds a template to the registry.
//
// templates is a list of template file name relative to the template/ directory
// that are needed to render the full template. The first file will be used as the name for
// the template group.
func AddTemplate(paths ...string) *template.Template {
  if len(paths) == 0 {
    return nil
  }
  id := paths[0]
  for i, v := range paths {
    paths[i] = fmt.Sprintf("template/%s", v)
  }
  tmpl := template.New("").Funcs(tmplFuncRegistry)
  for _, path := range paths {
    b, err := ioutil.ReadFile(path)
    if err != nil {
      panic(err)
    }
    s := string(b)
    tmpl = template.Must(tmpl.New("").Funcs(tmplFuncRegistry).Parse(s))
  }
  if templateRegistry == nil {
    templateRegistry = make(map[string]*template.Template)
  }
  templateRegistry[id] = tmpl
  return tmpl
}

func RenderTemplate(w http.ResponseWriter, id string, name string, ctx interface{}) error {
  if templateRegistry == nil {
    return errors.New("No templates were registered.")
  }
  tmpl, exists := templateRegistry[id]
  if !exists {
    return errors.New(fmt.Sprintf("Template '%s' does not exist", id))
  }
  return tmpl.ExecuteTemplate(w, name, ctx)
}

type ctxBase struct {
  Title   string
  TimeNow string
  User    string
}

func (ctx *ctxBase) init(c appengine.Context) *ctxBase {
  ctx.Title = ""
  ctx.TimeNow = fmtTime(time.Now(), "America/Los_Angeles")
  if u := user.Current(c); u != nil {
    ctx.User = u.Email
  }
  return ctx
}
