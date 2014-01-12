package errwrap

import (
  "errors"
  "fmt"
  "runtime"
)

func Wrap(e error) error {
  if e == nil {
    return nil
  }
  _, file, line, ok := runtime.Caller(1)
  if ok {
    return errors.New(fmt.Sprintf("%s |>%s:%d", e.Error(), file, line))
  }
  return errors.New(fmt.Sprintf("%s |>(unknown)", e.Error()))
}
