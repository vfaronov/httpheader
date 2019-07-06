package httpheader

import (
	"fmt"
	"net/http"
	"testing"
)

func ExampleAllow() {
	header := http.Header{"Allow": []string{"GET, HEAD, OPTIONS"}}
	fmt.Printf("%q", Allow(header))
	// Output: ["GET" "HEAD" "OPTIONS"]
}

func ExampleSetAllow() {
	header := http.Header{}
	SetAllow(header, []string{"GET", "HEAD", "OPTIONS"})
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
			http.Header{"Foo": []string{"bar"}},
			nil,
		},
		{
			http.Header{"Allow": []string{""}},
			[]string{},
		},
		{
			http.Header{"Allow": []string{
				"",
				"",
			}},
			[]string{},
		},
		{
			http.Header{"Allow": []string{"GET"}},
			[]string{"GET"},
		},
		{
			http.Header{"Allow": []string{"GET,"}},
			[]string{"GET"},
		},
		{
			http.Header{"Allow": []string{",GET"}},
			[]string{"GET"},
		},
		{
			http.Header{"Allow": []string{"  ,\t ,, GET, , "}},
			[]string{"GET"},
		},
		{
			http.Header{"Allow": []string{"GET,HEAD"}},
			[]string{"GET", "HEAD"},
		},
		{
			http.Header{"Allow": []string{"GET, HEAD"}},
			[]string{"GET", "HEAD"},
		},
		{
			http.Header{"Allow": []string{
				"GET",
				"HEAD",
				"OPTIONS",
			}},
			[]string{"GET", "HEAD", "OPTIONS"},
		},
		{
			http.Header{"Allow": []string{
				"GET\t,\t  HEAD\t",
				"\tOPTIONS",
			}},
			[]string{"GET", "HEAD", "OPTIONS"},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Allow": []string{";;;"}},
			[]string{},
		},
		{
			http.Header{"Allow": []string{";;;,GET"}},
			[]string{"GET"},
		},
		{
			http.Header{"Allow": []string{"GET;;;whatever, HEAD"}},
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
	header := http.Header{"Vary": []string{"cookie, accept-encoding"}}
	fmt.Printf("%q", Vary(header))
	// Output: ["Cookie" "Accept-Encoding"]
}
