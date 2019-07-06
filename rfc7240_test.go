package httpheader

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"
)

func ExamplePrefer() {
	header := http.Header{"Prefer": []string{
		"wait=10, respond-async",
		`Foo; Bar="quux, xyzzy"`,
	}}
	prefer := Prefer(header)
	fmt.Printf("%q\n", prefer["wait"])
	fmt.Printf("%q\n", prefer["respond-async"])
	fmt.Printf("%q\n", prefer["foo"])
	// Output: {"10" map[]}
	// {"" map[]}
	// {"" map["bar":"quux, xyzzy"]}
}

func ExampleAddPrefer() {
	header := http.Header{}
	SetPrefer(header, map[string]Pref{
		"wait":          {"10", nil},
		"respond-async": {},
		"foo": {
			Value:  "bar, baz",
			Params: map[string]string{"quux": "xyzzy"},
		},
	})
}

func TestPrefer(t *testing.T) {
	tests := []struct {
		header http.Header
		result map[string]Pref
	}{
		// Valid headers.
		{
			http.Header{},
			nil,
		},
		{
			http.Header{"Prefer": {"respond-async"}},
			map[string]Pref{"respond-async": {"", nil}},
		},
		{
			http.Header{"Prefer": {"Respond-Async"}},
			map[string]Pref{"respond-async": {"", nil}},
		},
		{
			http.Header{"Prefer": {"wait=10"}},
			map[string]Pref{"wait": {"10", nil}},
		},
		{
			http.Header{"Prefer": {`wait="10"`}},
			map[string]Pref{"wait": {"10", nil}},
		},
		{
			http.Header{"Prefer": {"respond-async, wait=10"}},
			map[string]Pref{"respond-async": {"", nil}, "wait": {"10", nil}},
		},
		{
			http.Header{"Prefer": {
				"respond-async, wait=10",
				"handling=lenient",
			}},
			map[string]Pref{
				"respond-async": {"", nil},
				"wait":          {"10", nil},
				"handling":      {"lenient", nil},
			},
		},
		{
			http.Header{"Prefer": {"foo;baz"}},
			map[string]Pref{"foo": {"", map[string]string{"baz": ""}}},
		},
		{
			http.Header{"Prefer": {"foo; baz"}},
			map[string]Pref{"foo": {"", map[string]string{"baz": ""}}},
		},
		{
			http.Header{"Prefer": {"foo ;\t\t;; ; BAZ ;,"}},
			map[string]Pref{"foo": {"", map[string]string{"baz": ""}}},
		},
		{
			http.Header{"Prefer": {"foo=Bar; baz"}},
			map[string]Pref{"foo": {"Bar", map[string]string{"baz": ""}}},
		},
		{
			http.Header{"Prefer": {`foo="quoted \"bar\""; baz`}},
			map[string]Pref{"foo": {`quoted "bar"`, map[string]string{"baz": ""}}},
		},
		{
			http.Header{"Prefer": {`foo=bar;baz=Qux`}},
			map[string]Pref{"foo": {"bar", map[string]string{"baz": "Qux"}}},
		},
		{
			http.Header{"Prefer": {`foo=bar;baz="quoted \"qux\"" ;xyzzy`}},
			map[string]Pref{
				"foo": {"bar", map[string]string{
					"baz":   `quoted "qux"`,
					"xyzzy": "",
				}},
			},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Prefer": {"???"}},
			map[string]Pref{"???": {"", nil}},
		},
		{
			// Whitespace around '=' is not allowed by RFC 7240 errata 4439.
			http.Header{"Prefer": {"foo = bar"}},
			map[string]Pref{"foo": {"", nil}},
		},
		{
			http.Header{"Prefer": {"foo = bar, baz = qux"}},
			map[string]Pref{"foo": {"", nil}, "baz": {"", nil}},
		},
		{
			http.Header{"Prefer": {"foo=???"}},
			map[string]Pref{"foo": {"???", nil}},
		},
		{
			http.Header{"Prefer": {"foo bar; baz, qux"}},
			map[string]Pref{"foo": {"", nil}, "qux": {"", nil}},
		},
		{
			http.Header{"Prefer": {";;;, foo=yes"}},
			map[string]Pref{"": {"", nil}, "foo": {"yes", nil}},
		},
		{
			http.Header{"Prefer": {"foo=bar=baz"}},
			map[string]Pref{"foo": {"bar", nil}},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, Prefer(test.header))
		})
	}
}

func TestSetPrefer(t *testing.T) {
	tests := []struct {
		input  map[string]Pref
		result http.Header
	}{
		{
			map[string]Pref{"respond-async": Pref{}},
			http.Header{"Prefer": {"respond-async"}},
		},
		{
			map[string]Pref{"wait": {"10", nil}},
			http.Header{"Prefer": {"wait=10"}},
		},
		{
			map[string]Pref{"foo": {"bar", nil}},
			http.Header{"Prefer": {"foo=bar"}},
		},
		{
			map[string]Pref{"foo": {"bar baz", nil}},
			http.Header{"Prefer": {`foo="bar baz"`}},
		},
		{
			map[string]Pref{"foo": {`bar "baz"`, nil}},
			http.Header{"Prefer": {`foo="bar \"baz\""`}},
		},
		{
			map[string]Pref{"foo": {"", map[string]string{"qux": ""}}},
			http.Header{"Prefer": {`foo;qux`}},
		},
		{
			map[string]Pref{"foo": {"", map[string]string{"qux": "xyzzy"}}},
			http.Header{"Prefer": {`foo;qux=xyzzy`}},
		},
		{
			map[string]Pref{
				"foo": {"", map[string]string{"qux": `quoted "xyzzy"`}},
			},
			http.Header{"Prefer": {`foo;qux="quoted \"xyzzy\""`}},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			header := http.Header{}
			SetPrefer(header, test.input)
			checkSerialize(t, test.input, test.result, header)
		})
	}
}

func TestPreferFuzz(t *testing.T) {
	checkFuzz(t, "Prefer", Prefer, SetPrefer)
}

func TestPreferRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetPrefer, Prefer, func(r *rand.Rand) interface{} {
		return mkMap(r, mkLowerToken, func(r *rand.Rand) interface{} {
			return Pref{
				Value:  mkString(r).(string),
				Params: mkMap(r, mkLowerToken, mkString).(map[string]string),
			}
		})
	})
}
