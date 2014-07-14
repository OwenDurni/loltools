package tags

import (
  "fmt"
)

const (
  ReasonNotApplicable = "n/a"
)

func AutomaticallyDetectedMatchResult(matchTag string) string {
  return fmt.Sprintf("auto-result:%s", matchTag)
}