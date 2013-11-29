package view

import (
  "github.com/OwenDurni/loltools/model"
  "html/template"
  "net/http"
)

func ProfileEditHandler(w http.ResponseWriter, r *http.Request) {
  user := model.User{} 
  formContents, err := ParseTemplate("/template/profile/edit.html", user)
  if err != nil {
    print(err)
    return
  }

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
