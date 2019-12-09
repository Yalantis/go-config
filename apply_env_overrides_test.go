package config

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToCamelCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		value  string
		expect string
	}{
		{
			name:   "snake case",
			value:  "snake_case",
			expect: "SnakeCase",
		},
		{
			name:   "space case",
			value:  "space case",
			expect: "Space case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ToCamelCase(tt.value)
			if tt.expect != actual {
				t.Errorf("mismatch %s %s", tt.expect, actual)
			}
		})
	}
}

func TestApplyEnvOverridesToSlice(t *testing.T) {
	t.Parallel()

	type Payload struct {
		Addr    string        `envconfig:"ADDR"`
		Timeout time.Duration `envconfig:"TIMEOUT"`
	}

	tests := []struct {
		name   string
		prefix string
		value  interface{}
		expect interface{}
		envs   []string
		err    error
	}{
		{
			name: "no prefix",
			err:  ErrPrefixRequired,
		},
		{
			name:   "not a pointer",
			prefix: "PREFIX_S",
			value:  []Payload{},
			err:    ErrNotPointer,
		},
		{
			name:   "not a slice",
			prefix: "PREFIX_S",
			value:  new(int),
			err:    ErrNotSlice,
		},
		{
			name:   "fail on time.ParseDuration",
			prefix: "PREFIX_S",
			value:  new([]Payload),
			envs:   []string{"PREFIX_S_0_ADDR", "localhost", "PREFIX_S_0_TIMEOUT", "localhost"},
			err:    errors.New(`failed to parse value "localhost" as time.Duration type`),
		},
		{
			name:   "ok",
			prefix: "PREFIX_S",
			value:  new([]Payload),
			expect: new([]Payload),
		},
		{
			name:   "fill with env values",
			prefix: "PREFIX_S",
			value:  new([]Payload),
			expect: &[]Payload{{Addr: "localhost", Timeout: time.Minute}, {Timeout: 2 * time.Minute}},
			envs:   []string{"PREFIX_S_0_ADDR", "localhost", "PREFIX_S_0_TIMEOUT", "1m", "PREFIX_S_1_TIMEOUT", "2m"},
		},
		{
			name:   "merge with env values",
			prefix: "PREFIX_S",
			value:  &[]Payload{{Timeout: 10 * time.Minute}, {Timeout: 20 * time.Minute}, {Timeout: 30 * time.Minute}},
			expect: &[]Payload{{Addr: "localhost", Timeout: time.Minute}, {Timeout: 2 * time.Minute}, {Timeout: 30 * time.Minute}},
			envs:   []string{"PREFIX_S_0_ADDR", "localhost", "PREFIX_S_0_TIMEOUT", "1m", "PREFIX_S_1_TIMEOUT", "2m"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer envs{}.set(tt.envs...).unset()

			err := applyEnvOverridesToSlice(tt.prefix, tt.value)

			assert.Equal(t, tt.err, err, tt.name)
			if tt.err != nil && err != nil {
				return
			}

			assert.Equal(t, tt.expect, tt.value, tt.name)
		})
	}
}
