package view

import (
  "appengine"
  "appengine/user"
  "errors"
  "fmt"
  "html/template"
  "github.com/OwenDurni/loltools/riot"
  "io/ioutil"
  "net/http"
  "time"
)

func tmpl_form(id, uri, text string) interface{} {
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

func tmpl_ddc_name(id int) string {
  return riot.Lookup.Champions[id].Name
}
func tmpl_ddc_s(id int) string {
  return riot.Lookup.Champions[id].Sprite.Url
}
func tmpl_ddc_stw(id int) int {
  return riot.Lookup.SpriteSheets[tmpl_ddc_s(id)].W
}
func tmpl_ddc_sth(id int) int {
  return riot.Lookup.SpriteSheets[tmpl_ddc_s(id)].H
}
func tmpl_ddc_sw(id int) int {
  return riot.Lookup.Champions[id].Sprite.W
}
func tmpl_ddc_sh(id int) int {
  return riot.Lookup.Champions[id].Sprite.H
}
func tmpl_ddc_sx(id int) int {
  return riot.Lookup.Champions[id].Sprite.X
}
func tmpl_ddc_sy(id int) int {
  return riot.Lookup.Champions[id].Sprite.Y
}

func tmpl_ddi_name(id int) string {
  return riot.Lookup.Items[id].Name
}
func tmpl_ddi_s(id int) string {
  return riot.Lookup.Items[id].Sprite.Url
}
func tmpl_ddi_stw(id int) int {
  return riot.Lookup.SpriteSheets[tmpl_ddi_s(id)].W
}
func tmpl_ddi_sth(id int) int {
  return riot.Lookup.SpriteSheets[tmpl_ddi_s(id)].H
}
func tmpl_ddi_sw(id int) int {
  return riot.Lookup.Items[id].Sprite.W
}
func tmpl_ddi_sh(id int) int {
  return riot.Lookup.Items[id].Sprite.H
}
func tmpl_ddi_sx(id int) int {
  return riot.Lookup.Items[id].Sprite.X
}
func tmpl_ddi_sy(id int) int {
  return riot.Lookup.Items[id].Sprite.Y
}

func tmpl_dds_name(id int) string {
  return riot.Lookup.Summoners[id].Name
}
func tmpl_dds_s(id int) string {
  return riot.Lookup.Summoners[id].Sprite.Url
}
func tmpl_dds_stw(id int) int {
  return riot.Lookup.SpriteSheets[tmpl_dds_s(id)].W
}
func tmpl_dds_sth(id int) int {
  return riot.Lookup.SpriteSheets[tmpl_dds_s(id)].H
}
func tmpl_dds_sw(id int) int {
  return riot.Lookup.Summoners[id].Sprite.W
}
func tmpl_dds_sh(id int) int {
  return riot.Lookup.Summoners[id].Sprite.H
}
func tmpl_dds_sx(id int) int {
  return riot.Lookup.Summoners[id].Sprite.X
}
func tmpl_dds_sy(id int) int {
  return riot.Lookup.Summoners[id].Sprite.Y
}

func tmpl_even(i int) bool {
  return i%2 == 0
}

func tmpl_odd(i int) bool {
  return i%2 == 1
}

var templateRegistry map[string]*template.Template
var tmplFuncRegistry = template.FuncMap{
  // dd champion functions
  "ddc_name": tmpl_ddc_name,
  "ddc_s": tmpl_ddc_s,
  "ddc_stw": tmpl_ddc_stw,
  "ddc_sth": tmpl_ddc_sth,
  "ddc_sw": tmpl_ddc_sw,
  "ddc_sh": tmpl_ddc_sh,
  "ddc_sx": tmpl_ddc_sx,
  "ddc_sy": tmpl_ddc_sy,
  
  // dd item functions
  "ddi_name": tmpl_ddi_name,
  "ddi_s": tmpl_ddi_s,
  "ddi_stw": tmpl_ddi_stw,
  "ddi_sth": tmpl_ddi_sth,
  "ddi_sw": tmpl_ddi_sw,
  "ddi_sh": tmpl_ddi_sh,
  "ddi_sx": tmpl_ddi_sx,
  "ddi_sy": tmpl_ddi_sy,
  
  // dd summoner functions
  "dds_name": tmpl_dds_name,
  "dds_s": tmpl_dds_s,
  "dds_stw": tmpl_dds_stw,
  "dds_sth": tmpl_dds_sth,
  "dds_sw": tmpl_dds_sw,
  "dds_sh": tmpl_dds_sh,
  "dds_sx": tmpl_dds_sx,
  "dds_sy": tmpl_dds_sy,
  
  "even": tmpl_even,
  "form": tmpl_form,
  "odd":  tmpl_odd,
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
  Title    string
  TimeNow  string
  User     string
  Errors   []error
}

func (ctx *ctxBase) init(c appengine.Context) *ctxBase {
  ctx.Title = ""
  ctx.TimeNow = fmtTime(time.Now(), "America/Los_Angeles")
  if u := user.Current(c); u != nil {
    ctx.User = u.Email
  }
  ctx.Errors = make([]error, 0)
  return ctx
}

func (ctx *ctxBase) AddError(err error) {
  if err == nil { return }
  ctx.Errors = append(ctx.Errors, err)
}
