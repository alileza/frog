package json

import (
	"encoding/json"
	"errors"
	"reflect"
)

func ToMap(b []byte) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	if err := json.Unmarshal(b, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func ToType(b []byte) ([]byte, error) {
	m, err := ToMap(b)
	if err != nil {
		return nil, err
	}
	return json.Marshal(
		toType(m),
	)
}

func Compare(a, b []byte) error {
	result, diff := compare(a, b, &Options{})
	if result != FullMatch {
		return errors.New(diff)
	}
	return nil
}

func toType(m map[string]interface{}) map[string]interface{} {
	for key, val := range m {
		if val == nil {
			m[key] = "*"
			continue
		}
		t := reflect.TypeOf(val)
		if t.Kind() == reflect.Map {
			m[key] = toType(val.(map[string]interface{}))
			continue
		}
		m[key] = t.String()
	}
	return m
}
