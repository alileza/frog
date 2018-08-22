package json

import (
	"reflect"
	"testing"
)

func TestToType(t *testing.T) {
	for _, test := range []struct {
		input  []byte
		output []byte
	}{
		{
			[]byte(`{"name":"ali"}`),
			[]byte(`{"name":"string"}`),
		},
		{
			[]byte(`{"name":123}`),
			[]byte(`{"name":"float64"}`),
		},
		{
			[]byte(`{"name":{}}`),
			[]byte(`{"name":{}}`),
		},
		{
			[]byte(`{"name":{"age":123}}`),
			[]byte(`{"name":{"age":"float64"}}`),
		},
	} {
		out, err := ToType(test.input)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(out, test.output) {
			t.Errorf("%s != %s", string(out), string(test.output))
		}
	}
}
