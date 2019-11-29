package config

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDuration_MarshalJSON(t *testing.T) {
	d := struct {
		Duration Duration
	}{
		Duration: Duration(time.Second),
	}
	b, err := json.Marshal(d)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"Duration":"1s"}`), b)
}

func TestDuration_UnmarshalJSON(t *testing.T) {
	var d struct{ Duration Duration }
	err := json.Unmarshal([]byte(`{"Duration": "1m"}`), &d)
	assert.NoError(t, err)
	assert.Equal(t, time.Minute, time.Duration(d.Duration))

	var d2 struct{ Duration Duration }
	err = json.Unmarshal([]byte(`{"Duration": 1}`), &d2)
	assert.Error(t, err)
	assert.Equal(t, time.Duration(0), time.Duration(d2.Duration))
}
