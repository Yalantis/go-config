package config

import (
	"encoding/json"
	"time"
)

// Duration provides marshal/unmarshal of duration as a string
type Duration time.Duration

// MarshalJSON Duration as a string
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON Duration from float64/string
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	duration, err := time.ParseDuration(v)
	if err != nil {
		return err
	}
	*d = Duration(duration)
	return nil
}
