package view

import (
  "loltools/model"
  "template/html"
)

func ProfileEditHandler(w http.ResponseWriter, r *http.Request) {
  user := model.User{} 
  formContents, err := ParseTemplate("/template/profile/edit.html", user)

  formCtx := new(FormCtx)
  formCtx.Init()
  formCtx.FormId = "edit-profile"
  formCtx.SubmitUrl = "/profile/update"
  formCtx.FormHTML = template.HTML(formContents)
  if out := RenderForm(formCtx); out != nil {
    w.Write(out)
  }
}

func ProfileViewHandler(w http.ResponseWriter, r *http.Request) {

}
