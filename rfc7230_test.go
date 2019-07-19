package httpheader

import (
	"fmt"
	"net/http"
	"os"
	"testing"
)

func ExampleVia() {
	header := http.Header{"Via": {
		"1.1 proxy2.example.com:8080 (corporate)",
		"2 edge3.example.net",
	}}
	fmt.Print(Via(header))
	// Output: [{HTTP/1.1 proxy2.example.com:8080 corporate} {HTTP/2.0 edge3.example.net }]
}

func ExampleAddVia() {
	header := http.Header{}
	AddVia(header, ViaElem{ReceivedProto: "HTTP/1.1", ReceivedBy: "api-gw1"})
	header.Write(os.Stdout)
	// Output: Via: 1.1 api-gw1
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
			http.Header{"Via": {"FSTR/3 bar (some new protocol)"}},
			[]ViaElem{{"FSTR/3", "bar", "some new protocol"}},
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
		{
			// This is a valid received-by, per uri-host -> IPvFuture.
			http.Header{"Via": {
				`1.1 [v9.a51c00de,route=51]:8080 (IPv9 Powered), 1.1 example.net`,
			}},
			[]ViaElem{
				{"HTTP/1.1", "[v9.a51c00de,route=51]:8080", "IPv9 Powered"},
				{"HTTP/1.1", "example.net", ""},
			},
		},
		{
			// This is a valid received-by, per uri-host -> reg-name -> sub-delims,
			// but we currently don't parse it. This is a documented bug.
			http.Header{"Via": {
				`1.1 funky,reg-name, 1.1 example.net`,
			}},
			[]ViaElem{
				{"HTTP/1.1", "funky", ""},
				{"HTTP/reg-name", "", ""},
				{"HTTP/1.1", "example.net", ""},
			},
		},
		{
			http.Header{"Via": {"2 example.com, HTTP/2 example.net"}},
			[]ViaElem{
				{"HTTP/2.0", "example.com", ""},
				{"HTTP/2.0", "example.net", ""},
			},
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

func TestSetVia(t *testing.T) {
	tests := []struct {
		input  []ViaElem
		result http.Header
	}{
		{
			[]ViaElem{{"HTTP/1.1", "foo", ""}},
			http.Header{"Via": {`1.1 foo`}},
		},
		{
			[]ViaElem{{"FSTR/3.0", "bar.example.net:8080", "(baz)"}},
			http.Header{"Via": {`FSTR/3.0 bar.example.net:8080 (\(baz\))`}},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			header := http.Header{}
			SetVia(header, test.input)
			checkGenerate(t, test.input, test.result, header)
		})
	}
}

func TestViaRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetVia, Via,
		[]ViaElem{{
			ReceivedProto: "token/token",
			ReceivedBy:    "token",
			Comment:       "quotable | empty",
		}},
	)
}

func BenchmarkVia(b *testing.B) {
	header := http.Header{"Via": {"1.1 proxy2.example.net (CWA (corporate Web accelerator))", "2 api-front.example.com:443 (trace: 97G9Hcio), 2 gw1-3.svc.example.com"}}
	for i := 0; i < b.N; i++ {
		Via(header)
	}
}
