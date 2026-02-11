package env

import (
	"reflect"
)

func Parse(getenv func(string) string, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Struct {
		return nil
	}

	st := rv.Elem()
	t := st.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("env")
		if tag == "" {
			continue
		}

		if value := getenv(tag); value != "" {
			st.Field(i).SetString(value)
		}
	}

	return nil
}
