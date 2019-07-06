package httpheader

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"
)

func ExampleAllow() {
	header := http.Header{"Allow": {"GET, HEAD, OPTIONS"}}
	fmt.Printf("%q", Allow(header))
	// Output: ["GET" "HEAD" "OPTIONS"]
}

func ExampleSetAllow() {
	header := http.Header{}
	SetAllow(header, []string{"GET", "HEAD", "OPTIONS"})
}

func TestAllowRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetAllow, Allow, func(r *rand.Rand) interface{} {
		methods := mkSlice(r, mkToken).([]string)
		if methods == nil {
			methods = make([]string, 0)
		}
		return methods
	})
}

func TestAllow(t *testing.T) {
	tests := []struct {
		header http.Header
		result []string
	}{
		// Valid headers.
		{
			http.Header{},
			nil,
		},
		{
			http.Header{"Foo": {"bar"}},
			nil,
		},
		{
			http.Header{"Allow": {""}},
			[]string{},
		},
		{
			http.Header{"Allow": {
				"",
				"",
			}},
			[]string{},
		},
		{
			http.Header{"Allow": {"GET"}},
			[]string{"GET"},
		},
		{
			http.Header{"Allow": {"GET,"}},
			[]string{"GET"},
		},
		{
			http.Header{"Allow": {",GET"}},
			[]string{"GET"},
		},
		{
			http.Header{"Allow": {"  ,\t ,, GET, , "}},
			[]string{"GET"},
		},
		{
			http.Header{"Allow": {"GET,HEAD"}},
			[]string{"GET", "HEAD"},
		},
		{
			http.Header{"Allow": {"GET, HEAD"}},
			[]string{"GET", "HEAD"},
		},
		{
			http.Header{"Allow": {
				"GET",
				"HEAD",
				"OPTIONS",
			}},
			[]string{"GET", "HEAD", "OPTIONS"},
		},
		{
			http.Header{"Allow": {
				"GET\t,\t  HEAD\t",
				"\tOPTIONS",
			}},
			[]string{"GET", "HEAD", "OPTIONS"},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Allow": {";;;"}},
			[]string{},
		},
		{
			http.Header{"Allow": {";;;,GET"}},
			[]string{"GET"},
		},
		{
			http.Header{"Allow": {"GET;;;whatever, HEAD"}},
			[]string{"GET", "HEAD"},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, Allow(test.header))
		})
	}
}

func ExampleVary() {
	header := http.Header{"Vary": {"cookie, accept-encoding"}}
	fmt.Printf("%q", Vary(header))
	// Output: ["Cookie" "Accept-Encoding"]
}

func TestVaryRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetVary, Vary, func(r *rand.Rand) interface{} {
		return mkSlice(r, mkHeaderName)
	})
}
