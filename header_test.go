package httpheader

import (
	"net/http"
	"reflect"
	"testing"
)

func checkParse(t *testing.T, header http.Header, expected, actual interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("parsing: %#v\nexpected: %#v\nactual:   %#v",
			header, expected, actual)
	}
}

func checkSerialize(t *testing.T, input interface{}, expected, actual http.Header) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("serializing: %#v\nexpected: %#v\nactual:   %#v",
			input, expected, actual)
	}
}
