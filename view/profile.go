package view

import (
  "appengine"
  "github.com/OwenDurni/loltools/model"
  "html/template"
  "net/http"
)

func ProfileEditHandler(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)

  user, err := model.GetUser(c)
  if err != nil {
    c.Errorf(err.Error())
    return
  }

  formContents, err := ParseTemplate("template/profile/edit.html", user)
  if err != nil {
    c.Errorf(err.Error())
    return
  }

  formCtx := new(FormCtx)
  formCtx.Init()
  formCtx.FormId = "edit-profile"
  formCtx.SubmitUrl = "/profile/update"
  formCtx.FormHTML = template.HTML(formContents)
  formHtml, err := RenderForm(formCtx);
  if err != nil {
    c.Errorf(err.Error())
    return
  }
  
  pageCtx := new(CommonCtx)
  pageCtx.Init()
  pageCtx.Title = "Edit Profile"
  pageCtx.ContentHTML = template.HTML(formHtml)
  pageHtml, err := ParseTemplate("template/common.html", pageCtx)
  if err != nil {
    c.Errorf(err.Error())
    return
  }

  w.Write(pageHtml)
}

func ProfileViewHandler(w http.ResponseWriter, r *http.Request) {

}
