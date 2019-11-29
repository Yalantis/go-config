package config

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitOnConfiguration(t *testing.T) {
	t.Parallel()

	type Payload struct {
		Addr        string        `envconfig:"ADDR" default:"0.0.0.0"`
		AddrRenamed string        `envconfig:"HOST"`
		Timeout     time.Duration `envconfig:"TIMEOUT"`
	}

	type Configuration struct {
		Payload    []Payload  `envprefix:"APP"`
		PayloadPtr *[]Payload `envprefix:"APP"`
	}

	tests := []struct {
		name   string
		value  interface{}
		expect interface{}
		envs   []string
		err    error
	}{
		{
			name:   "not pointer",
			value:  Configuration{},
			expect: Configuration{},
			err:    ErrDstNotPointer,
		},
		{
			name:   "fail on time.ParseDuration",
			value:  &Configuration{},
			expect: &Configuration{},
			envs:   []string{"APP_0_ADDR", "localhost", "APP_0_TIMEOUT", "localhost"},
			err:    errors.New(`failed to parse value "localhost" as time.Duration type`),
		},
		{
			name:   "ok",
			value:  &Configuration{},
			expect: &Configuration{},
		},
		{
			name:  "fill by envconfig or field name",
			value: &Configuration{},
			expect: &Configuration{
				Payload:    []Payload{{Addr: "0.0.0.0", AddrRenamed: "localhost"}, {Addr: "0.0.0.0", AddrRenamed: "localhost2"}},
				PayloadPtr: &[]Payload{{Addr: "0.0.0.0", AddrRenamed: "localhost"}, {Addr: "0.0.0.0", AddrRenamed: "localhost2"}},
			},
			envs: []string{"APP_0_HOST", "localhost", "APP_1_ADDR_RENAMED", "localhost2"},
		},
		{
			name:  "fill with env values",
			value: &Configuration{},
			expect: &Configuration{
				Payload:    []Payload{{Addr: "0.0.0.0", AddrRenamed: "localhost", Timeout: time.Minute}, {Addr: "0.0.0.0", Timeout: 2 * time.Minute}},
				PayloadPtr: &[]Payload{{Addr: "0.0.0.0", AddrRenamed: "localhost", Timeout: time.Minute}, {Addr: "0.0.0.0", Timeout: 2 * time.Minute}},
			},
			envs: []string{"APP_0_HOST", "localhost", "APP_0_TIMEOUT", "1m", "APP_1_TIMEOUT", "2m"},
		},
		{
			name: "merge with env values",
			value: &Configuration{
				Payload:    []Payload{{Timeout: 10 * time.Minute}, {Timeout: 20 * time.Minute}, {Timeout: 30 * time.Minute}},
				PayloadPtr: &[]Payload{{Timeout: 10 * time.Minute}, {Timeout: 20 * time.Minute}, {Timeout: 30 * time.Minute}},
			},
			expect: &Configuration{
				Payload:    []Payload{{Addr: "localhost", Timeout: time.Minute}, {Addr: "0.0.0.0", Timeout: 2 * time.Minute}, {Addr: "0.0.0.0", Timeout: 30 * time.Minute}},
				PayloadPtr: &[]Payload{{Addr: "localhost", Timeout: time.Minute}, {Addr: "0.0.0.0", Timeout: 2 * time.Minute}, {Addr: "0.0.0.0", Timeout: 30 * time.Minute}},
			},
			envs: []string{"APP_0_ADDR", "localhost", "APP_0_TIMEOUT", "1m", "APP_1_TIMEOUT", "2m"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer envs{}.set(tt.envs...).unset()

			err := Init(tt.value, "")

			assert.Equal(t, tt.err, err, tt.name)
			if tt.err != nil && err != nil {
				return
			}

			assert.Equal(t, tt.expect, tt.value, tt.name)
		})
	}
}
