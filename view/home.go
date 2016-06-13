package view

import (
	"appengine"
	"github.com/OwenDurni/loltools/model"
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
	c := appengine.NewContext(r)
	// Ignore errors because the user may not be logged in.
	// Note that user may be nil.
	user, _, _ := model.GetUser(c)

	ctx := struct {
		ctxBase
	}{}
	ctx.ctxBase.init(c, user)

	err := RenderTemplate(w, "home.html", "base", ctx)
	if HandleError(c, w, err) {
		return
	}
}
