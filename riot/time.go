package riot

import (
	"encoding/json"
	"time"
)

type RiotTime time.Time

func (rt *RiotTime) UnmarshalJSON(data []byte) error {
	var epochMilliseconds int64
	if err := json.Unmarshal(data, &epochMilliseconds); err != nil {
		return err
	}
	seconds := epochMilliseconds / 1000
	nanoseconds := epochMilliseconds % 1000 * 1000000
	t := time.Unix(seconds, nanoseconds)
	*rt = (RiotTime)(t)
	return nil
}

func (rt RiotTime) MarshalJSON() ([]byte, error) {
	return json.Marshal((*time.Time)(&rt).Unix())
}
