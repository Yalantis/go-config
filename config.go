package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	ErrDstNotPointer = errors.New("dst should be a pointer")
	ErrNotStruct     = errors.New("should be a structure")
)

var (
	unixEpochTime      = time.Unix(0, 0)
	timeType           = reflect.TypeOf((*time.Time)(nil)).Elem()
	durationType       = reflect.TypeOf((*time.Duration)(nil)).Elem()
	durationCustomType = reflect.TypeOf((*Duration)(nil)).Elem()
)

const (
	envConfigTag = "envconfig"
	envPrefixTag = "envprefix"
)

// Init reads and init configuration to `config` variable, which must be a reference of struct
func Init(config interface{}, filename string) error {
	v := reflect.ValueOf(config)

	if v.Kind() != reflect.Ptr {
		return ErrDstNotPointer
	}

	v = reflect.Indirect(v)

	if v.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	if err := applyDefault(reflect.StructField{}, v); err != nil {
		return fmt.Errorf("init config with default values: %s", err)
	}

	if err := applyJSONConfig(config, filename); err != nil {
		return err
	}

	if err := applyEnv(v); err != nil {
		return err
	}

	return validate(v)
}

func validate(v reflect.Value) error {
	t := v.Type()

	var invalidFields []string

	for i := 0; i < v.NumField(); i++ {
		invalidFields = append(invalidFields, validateField(t.Field(i), v.Field(i))...)
	}

	if len(invalidFields) > 0 {
		return fmt.Errorf("required fields: %v are not filled up. Please check configuration", invalidFields)
	}

	return nil
}

func validateField(t reflect.StructField, v reflect.Value) (invalidFields []string) {
	if v.Kind() == reflect.Struct && !isTime(v.Type()) {
		for i := 0; i < v.NumField(); i++ {
			invalidFields = append(invalidFields, validateField(v.Type().Field(i), v.Field(i))...)
		}
		return invalidFields
	}

	value, ok := t.Tag.Lookup("required")
	if !ok || !isTrue(value) {
		return invalidFields
	}

	if isZero(v) {
		invalidFields = append(invalidFields, t.Name)
	}

	return invalidFields
}

// applyDefault recursively sets values to default
func applyDefault(t reflect.StructField, v reflect.Value) error {
	if v.Kind() == reflect.Struct && !isTime(v.Type()) {
		for i := 0; i < v.NumField(); i++ {
			if err := applyDefault(v.Type().Field(i), v.Field(i)); err != nil {
				return err
			}
		}
		return nil
	}

	value, ok := t.Tag.Lookup("default")
	if !ok {
		return nil
	}

	return setValue(v, value)
}

// setValue sets value depend on type
func setValue(v reflect.Value, value string) error {
	switch indirectType(v.Type()) {
	case timeType:
		return setTime(&v, value)
	case durationType:
		return setTimeDuration(&v, value)
	case durationCustomType:
		return setDuration(&v, value)
	}

	switch v.Kind() {
	case reflect.String:
		v.SetString(value)
	case reflect.Slice:
		setSlice(&v, value)
	case reflect.Int, reflect.Int32, reflect.Int64:
		return setInt(&v, value)
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		return setUint(&v, value)
	case reflect.Bool:
		setBool(&v, value)
	}

	return nil
}

func applyJSONConfig(config interface{}, filename string) error {
	if len(filename) == 0 {
		return nil
	}

	file, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(config)
}

func applyEnv(v reflect.Value) error {
	for i := 0; i < v.NumField(); i++ {
		err := applyEnvValue(v.Type().Field(i), v.Field(i))
		if err != nil {
			return err
		}
	}

	return nil
}

func applyEnvValue(t reflect.StructField, v reflect.Value) error {
	if value, ok := t.Tag.Lookup(envPrefixTag); ok && indirectType(v.Type()).Kind() == reflect.Slice {
		return applyEnvOverridesToSlice(value, v)
	}

	if v.Kind() == reflect.Struct && !isTime(v.Type()) {
		for i := 0; i < v.NumField(); i++ {
			err := applyEnvValue(v.Type().Field(i), v.Field(i))
			if err != nil {
				return err
			}
		}
		return nil
	}

	value, ok := t.Tag.Lookup(envConfigTag)
	if !ok {
		return nil
	}

	value, found := lookupEnv(value)
	if !found {
		return nil
	}

	return setValue(v, value)
}

func setTime(v *reflect.Value, value string) error {
	date, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return fmt.Errorf("failed to parse value %q as time.Time type", value)
	}
	if v.Kind() == reflect.Ptr {
		v.Set(reflect.ValueOf(&date))
	} else {
		v.Set(reflect.ValueOf(date))
	}
	return nil
}

func setTimeDuration(v *reflect.Value, value string) error {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return fmt.Errorf("failed to parse value %q as time.Duration type", value)
	}
	if v.Kind() == reflect.Ptr {
		v.Set(reflect.ValueOf(&duration))
	} else {
		v.Set(reflect.ValueOf(duration))
	}
	return nil
}

func setDuration(v *reflect.Value, value string) error {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return fmt.Errorf("failed to parse value %q as Duration type", value)
	}
	if v.Kind() == reflect.Ptr {
		d := Duration(duration)
		v.Set(reflect.ValueOf(&d))
	} else {
		v.Set(reflect.ValueOf(Duration(duration)))
	}
	return nil
}

func setInt(v *reflect.Value, value string) error {
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse value %q as Int64 type", value)
	}
	v.SetInt(intValue)
	return nil
}

func setUint(v *reflect.Value, value string) error {
	uintValue, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse value %q as Uint64 type", value)
	}
	v.SetUint(uintValue)
	return nil
}

func setBool(v *reflect.Value, value string) {
	v.SetBool(isTrue(value))
}

func setSlice(v *reflect.Value, value string) {
	if _, ok := v.Interface().([]string); !ok {
		return
	}

	values := strings.Split(value, ",")
	slice := reflect.MakeSlice(reflect.TypeOf([]string{}), len(values), len(values))

	for i, value := range values {
		slice.Index(i).SetString(value)
	}

	v.Set(slice)
}

func isTrue(value string) bool {
	return value == "1" || strings.ToLower(value) == "true"
}

func isTime(t reflect.Type) bool {
	return indirectType(t) == timeType
}

func isZeroTime(date time.Time) bool {
	return date.IsZero() || date.Equal(unixEpochTime)
}

func isZero(v reflect.Value) bool {
	switch v.Type() {
	case timeType:
		return isZeroTime(v.Interface().(time.Time))
	case durationType:
		return v.Interface().(time.Duration).Nanoseconds() == 0
	case durationCustomType:
		return time.Duration(v.Interface().(Duration)).Nanoseconds() == 0
	}

	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return math.Float64bits(v.Float()) == 0
	case reflect.Interface, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	case reflect.String:
		return v.Len() == 0
	}
	zero := reflect.Zero(v.Type())
	return reflect.DeepEqual(v.Interface(), zero.Interface())
}

// indirectWalk walks through Value to the last Value in chain
func indirectWalk(v reflect.Value) (rv reflect.Value) {
	for ; v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface; v = v.Elem() {
		if v.IsNil() {
			return v
		}
	}
	return v
}

// indirectType as reflect.Indirect but for Type
func indirectType(v reflect.Type) reflect.Type {
	for v.Kind() != reflect.Ptr {
		return v
	}
	return v.Elem()
}
