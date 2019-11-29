package config

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrDstNotSlice    = errors.New("dst must be a slice")
	ErrPrefixRequired = errors.New("prefix is required")
	ErrDstUnsettable  = errors.New("unsetable dst value")
)

var camelCaseRegex = regexp.MustCompile("(^[A-Za-z])|_([A-Za-z])")

// ToCamelCase converts string to camel case
func ToCamelCase(str string) string {
	return camelCaseRegex.ReplaceAllStringFunc(strings.ToLower(str), func(s string) string {
		return strings.ToUpper(strings.Replace(s, "_", "", -1))
	})
}

// applyEnvOverridesToSlice merges elements of slice with ENV
func applyEnvOverridesToSlice(prefix string, dst interface{}) error {
	if prefix == "" {
		return ErrPrefixRequired
	}

	rv, ok := dst.(reflect.Value)
	if !ok {
		// fallback
		rv = reflect.ValueOf(dst)
		if rv.Kind() != reflect.Ptr {
			return ErrDstNotPointer
		}
	}

	riv := indirectWalk(rv)
	rit := indirectType(riv.Type())

	if rit.Kind() != reflect.Slice {
		return ErrDstNotSlice
	}

	if !rv.CanSet() && !riv.CanSet() {
		return fmt.Errorf("unsettable type: %s %s at prefix: %s", rv.Type(), riv.Type(), prefix)
	}

	zero := rv.IsZero()
	parseKeyVal, err := regexp.Compile(prefix + "_(\\d+)_(\\w+)=(.+)")
	if err != nil {
		return err
	}

	sliceOf := rit.Elem()
	mapConfigs := make(map[int]reflect.Value)
	envs := environ()

	if !zero {
		l := riv.Len()
		for i := 0; i < l; i++ {
			value := reflect.Indirect(riv.Index(i))
			// set defaults for element, it was created after first defaults was applied
			if err := applyDefaultToEmpty(reflect.StructField{}, value); err != nil {
				return err
			}
			mapConfigs[i] = value
		}
	}

	for _, keyVal := range envs {
		if strings.HasPrefix(keyVal, prefix) {
			matches := parseKeyVal.FindStringSubmatch(keyVal)
			if len(matches) != 4 {
				continue
			}

			index, err := strconv.ParseInt(matches[1], 10, 64)
			if err != nil {
				return err
			}

			i := int(index)
			envKey := matches[2]
			name := ToCamelCase(matches[2])
			value := matches[3]

			if _, ok := mapConfigs[i]; !ok {
				value := reflect.Indirect(reflect.New(sliceOf))
				// set defaults for new created element
				if err := applyDefaultToEmpty(reflect.StructField{}, value); err != nil {
					return err
				}
				mapConfigs[i] = value
			}

			v := mapConfigs[i]

			f, ok := fieldByEnvconfig(v, envKey)
			if !ok {
				// fallback in case field with
				f = v.FieldByName(name)
				if (f == reflect.Value{}) {
					return fmt.Errorf("field %s not found", name)
				}
			}

			if err := setValue(f, value); err != nil {
				return err
			}
		}
	}

	if len(mapConfigs) == 0 {
		return nil
	}

	values := make([]reflect.Value, len(mapConfigs))
	for i, item := range mapConfigs {
		values[i] = item
	}

	ptr := reflect.New(rit)
	tmp := ptr.Elem()
	tmp.Set(reflect.Append(tmp, values...))

	if rv.CanSet() {
		setPtrValue(rv, ptr, tmp)
		return nil
	}

	if riv.CanSet() {
		setPtrValue(riv, ptr, tmp)
		return nil
	}

	return ErrDstUnsettable
}

// setPtrValue set as Ptr or Value
func setPtrValue(dst, ptr, val reflect.Value) {
	if dst.Kind() == reflect.Ptr {
		dst.Set(ptr)
	} else {
		dst.Set(val)
	}
}

// fieldByEnvconfig lookup field by `envconfig` tag
func fieldByEnvconfig(v reflect.Value, key string) (rv reflect.Value, ok bool) {
	for i, l := 0, v.NumField(); i < l; i++ {
		envconfig := v.Type().Field(i).Tag.Get(envConfigTag)
		if key == envconfig {
			return v.Field(i), true
		}
	}
	return
}

// applyDefaultToEmpty applies default to empty field only
func applyDefaultToEmpty(t reflect.StructField, v reflect.Value) error {
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			if err := applyDefaultToEmpty(v.Type().Field(i), v.Field(i)); err != nil {
				return err
			}
		}
		return nil
	}

	if !isEmpty(v) {
		return nil
	}

	value, ok := t.Tag.Lookup("default")
	if !ok {
		return nil
	}

	return setValue(v, value)
}
