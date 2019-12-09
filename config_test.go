package config

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	t.Parallel()

	var cfg struct {
		User     string `json:"user"      envconfig:"INIT_USER"      default:""`
		Password string `json:"pass"      envconfig:"INIT_PASS"      default:""`
		PoolSize uint64 `json:"pool_size" envconfig:"INIT_POOL_SIZE" default:"0"`
		Enabled  bool   `json:"enabled"   envconfig:"INIT_ENABLED"   default:"true"`
	}

	tests := []struct {
		name     string
		filePath string
		cfg      interface{}
		envs     []string
		error    string
	}{
		{
			name:     "not a pointer",
			filePath: "",
			cfg:      cfg,
			error:    ErrNotPointer.Error(),
		},
		{
			name:     "not a struct",
			filePath: "",
			cfg:      new(int),
			error:    ErrNotStruct.Error(),
		},

		{
			name:     "no path to config",
			filePath: "",
			cfg:      &cfg,
		},
		{
			name:     "no path to config",
			filePath: "./testdata/config.json",
			cfg:      &cfg,
			envs:     []string{"INIT_USER", "user", "INIT_PASS", "pwd", "INIT_POOL_SIZE", "1", "INIT_ENABLED", "false"},
		},
		{
			name:     "wrong path to config",
			filePath: "config.json",
			cfg:      &cfg,
			error:    "open config.json: no such file or directory",
		},
		{
			name:     "invalid config file content",
			filePath: "config_test.go",
			cfg:      &cfg,
			error:    "invalid character 'p' looking for beginning of value",
		},
		{
			name:     "wrong env type",
			filePath: "./testdata/config.json",
			cfg:      &cfg,
			envs:     []string{"INIT_POOL_SIZE", "1.0"},
			error:    `failed to parse value "1.0" as Uint64 type`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer envs{}.set(tt.envs...).unset()

			err := Init(tt.cfg, tt.filePath)

			if tt.error != "" {
				assert.EqualError(t, err, tt.error, tt.name)
				return
			}
			assert.NoError(t, err, tt.name)
		})
	}
}

func TestApplyEnvironment(t *testing.T) {
	t.Parallel()

	cfg := &struct {
		Name        string     `json:"name"        envconfig:"ENV_NAME"         default:"def_name"`
		Pass        string     `json:"pass"        envconfig:"ENV_PASS"         default:"def_pass"`
		Age         int        `json:"age"         envconfig:"ENV_AGE"          default:"18"`
		Time        time.Time  `json:"time"        envconfig:"ENV_TIME"         default:"2019-07-07T20:00:00Z"`
		TimePtr     *time.Time `json:"timePtr"     envconfig:"ENV_TIME_PTR"     default:"2019-07-07T20:00:00Z"`
		Duration    Duration   `json:"duration"    envconfig:"ENV_DURATION"     default:"10s"`
		DurationPtr *Duration  `json:"durationPtr" envconfig:"ENV_DURATION_PTR" default:"10s"`
		Embedded    struct {
			ID     int  `json:"id"     envconfig:"ENV_ID"     default:"1"`
			Exists bool `json:"exists" envconfig:"ENV_EXISTS" default:"false"`
			Absent bool `json:"absent" envconfig:"ENV_ABSENT" default:"true"`
		}
	}{}

	e := reflect.TypeOf(cfg).Elem()
	v := reflect.ValueOf(cfg).Elem()

	defer envs{}.set(
		"ENV_NAME", "def_name",
		"ENV_PASS", "def_pass",
		"ENV_AGE", "18",
		"ENV_TIME", "2019-07-07T20:00:00Z",
		"ENV_TIME_PTR", "2019-07-07T20:00:00Z",
		"ENV_DURATION", "10s",
		"ENV_DURATION_PTR", "10s",
		"ENV_ID", "1",
		"ENV_EXISTS", "false",
		"ENV_ABSENT", "true",
	).unset()

	err := applyEnv(v)
	assert.NoError(t, err)

	for i := 0; i < v.NumField(); i++ {
		helperAssertFieldEnvironment(t, e.Field(i), v.Field(i))
	}
}

func TestApplyDefaultValues(t *testing.T) {
	t.Parallel()

	cfg := &struct {
		Name        string     `json:"name"        default:"def_name"`
		Pass        string     `json:"pass"        default:"def_pass"`
		Age         int        `json:"age"         default:"18"`
		Time        time.Time  `json:"time"        default:"2019-07-07T20:00:00Z"`
		TimePtr     *time.Time `json:"timePtr"     default:"2019-07-07T20:00:00Z"`
		Duration    Duration   `json:"duration"    default:"10s"`
		DurationPtr *Duration  `json:"durationPtr" default:"10s"`
		Embedded    struct {
			ID     int  `json:"id"     default:"1"`
			Exists bool `json:"exists" default:"false"`
			Absent bool `json:"absent" default:"true"`
		}
	}{}

	e := reflect.TypeOf(cfg).Elem()
	v := reflect.ValueOf(cfg).Elem()

	err := applyDefault(reflect.StructField{}, v)
	assert.NoError(t, err)

	for i := 0; i < v.NumField(); i++ {
		helperAssertFieldDefaultValue(t, e.Field(i), v.Field(i))
	}

	// test cases
	tests := []struct {
		name    string
		error   string
		payload func() (reflect.StructField, reflect.Value)
	}{
		{
			name: "no default",
			payload: func() (reflect.StructField, reflect.Value) {
				type test struct {
					Age int
				}
				e := reflect.TypeOf(&test{}).Elem()
				v := reflect.ValueOf(&test{}).Elem()
				return e.Field(0), v.Field(0)
			},
		},
		{
			name: "default slice",
			payload: func() (reflect.StructField, reflect.Value) {
				type test struct {
					User struct {
						Ages []string `default:"[1,2]"`
					}
				}
				e := reflect.TypeOf(&test{}).Elem()
				v := reflect.ValueOf(&test{}).Elem()
				return e.Field(0), v.Field(0)
			},
		},
		{
			name:  "fail to parse int",
			error: `failed to parse value "age" as Int64 type`,
			payload: func() (reflect.StructField, reflect.Value) {
				type test struct {
					User struct {
						Age int `default:"age"`
					}
				}
				e := reflect.TypeOf(&test{}).Elem()
				v := reflect.ValueOf(&test{}).Elem()
				return e.Field(0), v.Field(0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, val := tt.payload()
			err := applyDefault(typ, val)
			if tt.error != "" {
				assert.EqualError(t, err, tt.error)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestApplyEnvValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		error   string
		payload func() (reflect.StructField, reflect.Value)
		envs    []string
	}{
		{
			name: "no default",
			payload: func() (reflect.StructField, reflect.Value) {
				type test struct {
					Age int
				}
				e := reflect.TypeOf(&test{}).Elem()
				v := reflect.ValueOf(&test{}).Elem()
				return e.Field(0), v.Field(0)
			},
		},
		{
			name:  "fail to parse int",
			error: `failed to parse value "age" as Int64 type`,
			payload: func() (reflect.StructField, reflect.Value) {
				type test struct {
					User struct {
						Age int `envconfig:"age"`
					}
				}
				e := reflect.TypeOf(&test{}).Elem()
				v := reflect.ValueOf(&test{}).Elem()
				return e.Field(0), v.Field(0)
			},
			envs: []string{"age", "age"},
		},
		{
			name: "default slice",
			payload: func() (reflect.StructField, reflect.Value) {
				type test struct {
					User struct {
						Ages []string `envconfig:"ages"`
					}
				}
				e := reflect.TypeOf(&test{}).Elem()
				v := reflect.ValueOf(&test{}).Elem()
				return e.Field(0), v.Field(0)
			},
			envs: []string{"ages", "[1,2]"},
		},
		{
			name: "env not found",
			payload: func() (reflect.StructField, reflect.Value) {
				type test struct {
					User struct {
						NotFound int `envconfig:"not_found"`
					}
				}
				e := reflect.TypeOf(&test{}).Elem()
				v := reflect.ValueOf(&test{}).Elem()
				return e.Field(0), v.Field(0)
			},
		},
		{
			name: "EnvOverrider",
			payload: func() (reflect.StructField, reflect.Value) {
				type test struct {
					TestApplyEnvOverrides *fakeApplyEnvOverrides
				}
				e := reflect.TypeOf(&test{}).Elem()
				v := reflect.ValueOf(&test{}).Elem()
				return e.Field(0), v.Field(0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, val := tt.payload()
			defer envs{}.set(tt.envs...).unset()

			err := applyEnvValue(typ, val)
			if tt.error != "" {
				assert.EqualError(t, err, tt.error)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestValidateField(t *testing.T) {
	t.Parallel()

	// error case
	type (
		User struct {
			EmptyString string            `required:"true"`
			EmptyInt    int               `required:"true"`
			EmptySlice  []int             `required:"true"`
			Int         int               `required:"true"`
			PtrNil      *int              `required:"true"`
			Ptr         *int              `required:"true"`
			Map         map[string]string `required:"true"`
			TimeNil     *time.Time        `required:"true"`
			Time        time.Time         `required:"true"`
			DurationNil *Duration         `required:"true"`
			Duration    Duration          `required:"true"`
		}
		test struct {
			User User
		}
	)

	toPtr := func(i int) *int { return &i }

	data := &test{User: User{Int: 1, Ptr: toPtr(1)}}
	e := reflect.TypeOf(data).Elem()
	v := reflect.ValueOf(data).Elem()

	invalidFields := validateField(e.Field(0), v.Field(0))
	assert.Len(t, invalidFields, 9)

	err := validate(v)
	assert.Error(t, err)
}

func helperAssertFieldDefaultValue(t *testing.T, f reflect.StructField, v reflect.Value) {
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			helperAssertFieldDefaultValue(t, v.Type().Field(i), v.Field(i))
		}
		return
	}

	value, ok := f.Tag.Lookup("default")
	if !ok {
		return
	}

	helperAssert(t, f, v, value)
}

func helperAssertFieldEnvironment(t *testing.T, f reflect.StructField, v reflect.Value) {
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			helperAssertFieldEnvironment(t, v.Type().Field(i), v.Field(i))
		}
		return
	}

	value, ok := f.Tag.Lookup(envConfigTag)
	if !ok {
		return
	}

	value, _ = lookupEnv(value)

	helperAssert(t, f, v, value)
}

func helperAssert(t *testing.T, f reflect.StructField, v reflect.Value, value string) {
	switch indirectType(v.Type()) {
	case timeType:
		date, _ := time.Parse(time.RFC3339, value)
		assert.Equal(t, date, reflect.Indirect(v).Interface(), f.Name)
		return

	case durationType:
		d, _ := time.ParseDuration(value)
		assert.Equal(t, d, reflect.Indirect(v).Interface(), f.Name)
		assert.Equal(t, value, reflect.Indirect(v).Interface().(time.Duration).String(), f.Name)
		return

	case durationCustomType:
		d, _ := time.ParseDuration(value)
		assert.Equal(t, Duration(d), reflect.Indirect(v).Interface(), f.Name)
		assert.Equal(t, value, time.Duration(reflect.Indirect(v).Interface().(Duration)).String(), f.Name)
		return
	}

	switch v.Kind() {
	case reflect.String:
		assert.Equal(t, value, v.String(), f.Name)

	case reflect.Slice:
		sl := make([]string, 0, v.Len())
		for i := 0; i < v.Len(); i++ {
			sl = append(sl, v.Index(i).String())
		}
		assert.Equal(t, value, strings.Join(sl, ","), f.Name)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		assert.Equal(t, value, strconv.FormatInt(v.Int(), 10), f.Name)

	case reflect.Bool:
		assert.Equal(t, value, strconv.FormatBool(v.Bool()), f.Name)
	}
}

type fakeApplyEnvOverrides struct{}

func (*fakeApplyEnvOverrides) ApplyEnvOverrides() error { return nil }
