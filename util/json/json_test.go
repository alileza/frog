package json

import (
	"testing"
)

func TestToType(t *testing.T) {
	for _, test := range []struct {
		A map[string]interface{}
		B map[string]interface{}
	}{
		{
			map[string]interface{}{"name": "string", "age": map[string]interface{}{"abc": "float64"}},
			map[string]interface{}{"name": "string", "age": map[string]interface{}{"abc": "float64"}},
		},
	} {
		out, err := Cmp(test.A, test.B, "")
		t.Error(err)
		t.Errorf("%+v", out)
	}
}
