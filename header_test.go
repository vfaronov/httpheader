package httpheader

import (
	"net/http"
	"reflect"
	"testing"
)

func checkParse(t *testing.T, header http.Header, expected, actual interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("header: %#v\nexpected: %#v\nactual:   %#v",
			header, expected, actual)
	}
}
