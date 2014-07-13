package view

import (
  "appengine"
  "appengine/user"
  "errors"
  "fmt"
  "github.com/OwenDurni/loltools/model"
  "github.com/OwenDurni/loltools/riot"
  "html/template"
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

func tmpl_unzip(values ...interface{}) (map[string]interface{}, error) {
  if len(values)%2 != 0 {
    return nil, errors.New("error: unzip must have even number of arguments")
  }
  dict := make(map[string]interface{}, len(values)/2)
  for i := 0; i < len(values); i+=2 {
    key, ok := values[i].(string)
    if !ok {
      return nil, errors.New("error: unzip keys must be strings")
    }
    dict[key] = values[i+1]
  }
  return dict, nil
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

func tmpl_gold(gold int) string {
  return fmt.Sprintf("%0.1fk", float64(gold) / 1000.)
}

func tmpl_riot_history_link(region string, gameId int64) string {
  return fmt.Sprintf(
    "http://matchhistory.%s.leagueoflegends.com/%s/#match-details/%s/%d",
    region, "en", "NA1", gameId);
}

func tmpl_time_deltanow(t time.Time) string {
  now := time.Now()
  
  var num    int
  var unit   string
  var suffix string
  var delta  time.Duration
  
  if now.Unix() >= t.Unix() {
    suffix = "ago"
    delta = now.Sub(t)
  } else {
    suffix = "from now"
    delta = t.Sub(now)
  }
  
  seconds := delta.Seconds()
  minutes := delta.Minutes()
  hours := delta.Hours()
  days := int(hours / 24)
  months := int(hours / (24*30))
  years := int(hours / (24*365))
  
  if years >= 1 {
    unit = "year(s)"
    num = years
  } else if months >= 1 {
    unit = "month(s)"
    num = months
  } else if days >= 1 {
    unit = "day(s)"
    num = days
  } else if hours >= 1 {
    unit = "hour(s)"
    num = int(hours)
  } else if minutes >= 1 {
    unit = "minute(s)"
    num = int(minutes)
  } else {
    unit = "second(s)"
    num = int(seconds)
  }
  
  return fmt.Sprintf("%d %s %s", num, unit, suffix)
}

var templateRegistry map[string]*template.Template
var tmplFuncRegistry = template.FuncMap{
  // combining pipelines
  "unzip": tmpl_unzip,
  
  // dd champion functions
  "ddc_name": tmpl_ddc_name,
  "ddc_s":    tmpl_ddc_s,
  "ddc_stw":  tmpl_ddc_stw,
  "ddc_sth":  tmpl_ddc_sth,
  "ddc_sw":   tmpl_ddc_sw,
  "ddc_sh":   tmpl_ddc_sh,
  "ddc_sx":   tmpl_ddc_sx,
  "ddc_sy":   tmpl_ddc_sy,

  // dd item functions
  "ddi_name": tmpl_ddi_name,
  "ddi_s":    tmpl_ddi_s,
  "ddi_stw":  tmpl_ddi_stw,
  "ddi_sth":  tmpl_ddi_sth,
  "ddi_sw":   tmpl_ddi_sw,
  "ddi_sh":   tmpl_ddi_sh,
  "ddi_sx":   tmpl_ddi_sx,
  "ddi_sy":   tmpl_ddi_sy,

  // dd summoner functions
  "dds_name": tmpl_dds_name,
  "dds_s":    tmpl_dds_s,
  "dds_stw":  tmpl_dds_stw,
  "dds_sth":  tmpl_dds_sth,
  "dds_sw":   tmpl_dds_sw,
  "dds_sh":   tmpl_dds_sh,
  "dds_sx":   tmpl_dds_sx,
  "dds_sy":   tmpl_dds_sy,

  "even": tmpl_even,
  "form": tmpl_form,
  "gold": tmpl_gold,
  "odd":  tmpl_odd,
  "riot_history_link": tmpl_riot_history_link,
  "time_deltanow": tmpl_time_deltanow,
}

var root string;
// Sets the root for further calls to AddTemplate
func SetTemplateRoot(newroot string) {
  root = newroot;
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
    paths[i] = fmt.Sprintf("%s%s", root, v)
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
  Errors  []error
  Regions []string
  TimeNow string
  Title   string
  User    string
}

func (ctx *ctxBase) init(c appengine.Context, mUser *model.User) *ctxBase {
  ctx.Errors = make([]error, 0)
  ctx.Regions = model.Regions
  ctx.TimeNow = fmtTime(time.Now(), "America/Los_Angeles")
  ctx.Title = ""
  if mUser != nil {
    if mUser.DisplayName != "" {
      ctx.User = fmt.Sprintf("%s", mUser.DisplayName)
    } else {
      ctx.User = fmt.Sprintf("[%s]", mUser.Email)
    }
  } else if u := user.Current(c); u != nil {
    ctx.User = fmt.Sprintf("[%s]", u.Email)
  }
  return ctx
}

func (ctx *ctxBase) AddError(err ...error) {
  if err == nil {
    return
  }
  ctx.Errors = append(ctx.Errors, err...)
}
