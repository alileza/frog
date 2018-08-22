package json

import (
	"encoding/json"
	"errors"
	"reflect"
)

type Diff struct {
	Field  string
	ValueA string
	ValueB string
}

func Cmp(A, B map[string]interface{}, parentKey string) ([]Diff, error) {
	result := make([]Diff, 0)

	for key := range A {
		a := A[key]
		at := reflect.TypeOf(a)

		b, ok := B[key]
		if !ok {
			if at.Kind() == reflect.Map {
				a = at.String()
			}
			if a.(string) == "*" {
				continue
			}

			result = append(result, Diff{parentKey + "." + key, a.(string), "missing"})
			continue
		}

		if at.Kind() == reflect.Map {
			diffs, err := Cmp(a.(map[string]interface{}), b.(map[string]interface{}), key)
			if err != nil {
				return nil, err
			}
			result = append(result, diffs...)
			continue
		}

		if b.(string) == "*" {
			continue
		}

		if a != b {
			result = append(result, Diff{parentKey + "." + key, a.(string), b.(string)})
		}
	}

	return result, nil
}

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
