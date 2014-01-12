package riot

import (
  "encoding/json"
  "time"
)

type RiotTime time.Time

func (rt *RiotTime) UnmarshalJSON(data []byte) error {
  var epochSeconds int64
  if err := json.Unmarshal(data, &epochSeconds); err != nil {
    return err
  }
  t := time.Unix(epochSeconds, 0)
  *rt = (RiotTime)(t)
  return nil
}

func (rt RiotTime) MarshalJSON() ([]byte, error) {
  return json.Marshal((*time.Time)(&rt).Unix())
}
