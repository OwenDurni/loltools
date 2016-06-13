package tags

import (
	"appengine/datastore"
	"fmt"
	"github.com/OwenDurni/loltools/model"
)

func AutomaticallyDetectedMatchResultFor(matchKey *datastore.Key) string {
	return fmt.Sprintf("auto-result:%s", model.MatchId(matchKey))
}

func ReasonNotApplicable() string {
	return "n/a"
}
