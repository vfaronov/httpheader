package httpheader

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"
)

func ExampleVia() {
	header := http.Header{"Via": {
		"1.1 foo.example.com:8080 (corporate)",
		"2 bar.example.net",
	}}
	fmt.Printf("%q", Via(header))
	// Output: [{"HTTP/1.1" "foo.example.com:8080" "corporate"} {"HTTP/2" "bar.example.net" ""}]
}

func ExampleAddVia() {
	header := http.Header{"Via": {"1.0 foo"}}
	AddVia(header, ViaElem{
		ReceivedProto: "HTTP/1.1",
		ReceivedBy:    "my-service",
	})
}

func TestVia(t *testing.T) {
	tests := []struct {
		header http.Header
		result []ViaElem
	}{
		// Valid headers.
		{
			http.Header{"Via": {"1.0 foo"}},
			[]ViaElem{{"HTTP/1.0", "foo", ""}},
		},
		{
			http.Header{"Via": {"1.0 \tfoo"}},
			[]ViaElem{{"HTTP/1.0", "foo", ""}},
		},
		{
			http.Header{"Via": {"1.0 foo  "}},
			[]ViaElem{{"HTTP/1.0", "foo", ""}},
		},
		{
			http.Header{"Via": {"1.0 foo  ,"}},
			[]ViaElem{{"HTTP/1.0", "foo", ""}},
		},
		{
			http.Header{"Via": {"1.0 foo\t (comment)"}},
			[]ViaElem{{"HTTP/1.0", "foo", "comment"}},
		},
		{
			http.Header{"Via": {
				"1.0 foo,1.0   bar\t, \t 1.0 baz,,",
				"1.1 qux",
			}},
			[]ViaElem{
				{"HTTP/1.0", "foo", ""},
				{"HTTP/1.0", "bar", ""},
				{"HTTP/1.0", "baz", ""},
				{"HTTP/1.1", "qux", ""},
			},
		},
		{
			http.Header{"Via": {
				"HTTP/2 foo",
				"FSTR/3 bar (some new protocol)",
			}},
			[]ViaElem{
				{"HTTP/2", "foo", ""},
				{"FSTR/3", "bar", "some new protocol"},
			},
		},
		{
			http.Header{"Via": {"1.1 foo (comment (with) nesting)"}},
			[]ViaElem{{"HTTP/1.1", "foo", "comment (with) nesting"}},
		},
		{
			http.Header{"Via": {"1.1 foo (comment (with nesting))"}},
			[]ViaElem{{"HTTP/1.1", "foo", "comment (with nesting)"}},
		},
		{
			http.Header{"Via": {`1.1 foo (comment with \) quoting)`}},
			[]ViaElem{{"HTTP/1.1", "foo", "comment with ) quoting"}},
		},
		{
			http.Header{"Via": {
				`1.1 foo (comment (with \) quoting) and nesting)`,
			}},
			[]ViaElem{
				{"HTTP/1.1", "foo", "comment (with ) quoting) and nesting"},
			},
		},
		{
			http.Header{"Via": {`1.1 foo (\strange quoting)`}},
			[]ViaElem{{"HTTP/1.1", "foo", "strange quoting"}},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Via": {"1.0"}},
			[]ViaElem{{"HTTP/1.0", "", ""}},
		},
		{
			http.Header{"Via": {"1.0, 1.1 foo, 1.2, 1.3 bar"}},
			[]ViaElem{
				{"HTTP/1.0", "", ""},
				{"HTTP/1.1", "foo", ""},
				{"HTTP/1.2", "", ""},
				{"HTTP/1.3", "bar", ""},
			},
		},
		{
			http.Header{"Via": {
				"1.1 foo (unterminated",
				"1.1 bar",
			}},
			[]ViaElem{
				{"HTTP/1.1", "foo", "unterminated"},
				{"HTTP/1.1", "bar", ""},
			},
		},
		{
			http.Header{"Via": {"1.1 foo (unterminated (with nesting)"}},
			[]ViaElem{
				{"HTTP/1.1", "foo", "unterminated (with nesting)"},
			},
		},
		{
			http.Header{"Via": {
				`1.1 foo (unterminated with \quoting (and nesting`,
				"1.1 bar",
			}},
			[]ViaElem{
				{"HTTP/1.1", "foo", "unterminated with quoting (and nesting"},
				{"HTTP/1.1", "bar", ""},
			},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, Via(test.header))
		})
	}
}

func TestViaRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetVia, Via, func(r *rand.Rand) interface{} {
		return mkSlice(r, func(r *rand.Rand) interface{} {
			return ViaElem{
				ReceivedProto: mkToken(r).(string)+"/"+mkToken(r).(string),
				ReceivedBy:    mkToken(r).(string),
				Comment:       mkString(r).(string),
			}
		})
	})
}
