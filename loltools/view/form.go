package view

import (
  "template/html"
)

type FormCtx struct {
  // HTML id attribute for the form element.
  FormId string
  // Text for the button that submits the form.
  SubmitText string
  // Url to send the form contents to via POST.
  SubmitUrl string
  // Status text to display while the form is sending.
  StatusTextActive string
  // Status text to display if the form send completes successfully.
  StatusTextDone string
  // Status text prefix to display if the form send completes unsuccessfully.
  StatusTextFail string

  // Content of the form as rendered HTML.
  FormHTML template.HTML
}

// Initializes a FormCtx with reasonable default values.
func (ctx *FormCtx) Init() {
  ctx.FormId = ""
  ctx.SubmitText = "Save"
  ctx.SubmitUrl = "#"
  ctx.StatusTextActive = "Saving..."
  ctx.StatusTextDone = "Saved"
  ctx.StatusTextFail = "Error saving: "
  ctx.FormHTML = template.HTML("")
}

func RenderForm(ctx *FormCtx) (out []byte) {
  out, err := parseTemplate("template/form.html", ctx)
  if err != nil {
    print(err.Error())
  }
  return out
}
