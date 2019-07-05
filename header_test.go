package httpheader

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func ExampleAllow() {
	header := http.Header{"Allow": []string{"GET, HEAD, OPTIONS"}}
	fmt.Printf("%q", Allow(header))
	// Output: ["GET" "HEAD" "OPTIONS"]
}

var tabAllow = []struct {
	header http.Header
	result []string
}{
	{
		http.Header{},
		nil,
	},
	{
		http.Header{"Foo": []string{"bar"}},
		nil,
	},
	{
		http.Header{"Allow": []string{}},
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
		http.Header{"Allow": []string{
			"",
			"???",
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
		http.Header{"Allow": []string{"  , ,, GET, , "}},
		[]string{"GET"},
	},
	{
		http.Header{"Allow": []string{
			"",
			"???",
			",,,",
			"GET,,,",
			"",
		}},
		[]string{"GET"},
	},
	{
		http.Header{"Allow": []string{"GET,HEAD"}},
		[]string{"GET", "HEAD"},
	},
	{
		http.Header{"Allow": []string{"GET, HEAD, POST"}},
		[]string{"GET", "HEAD", "POST"},
	},
	{
		http.Header{"Allow": []string{
			"GET???whatever, HEAD,",
			"",
			"",
			",,OPTIONS",
		}},
		[]string{"GET", "HEAD", "OPTIONS"},
	},
}

func TestAllow(t *testing.T) {
	for _, tt := range tabAllow {
		t.Run("", func(t *testing.T) {
			result := Allow(tt.header)
			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("on %#v: got %#v, wanted %#v",
					tt.header, result, tt.result)
			}
		})
	}
}

func ExampleSetAllow() {
	header := http.Header{}
	SetAllow(header, []string{"GET", "HEAD", "OPTIONS"})
	fmt.Printf("%q", header)
	// Output: map["Allow":["GET, HEAD, OPTIONS"]]
}

func ExampleVary() {
	header := http.Header{"Vary": []string{"cookie, accept-encoding"}}
	fmt.Printf("%q", Vary(header))
	// Output: ["Cookie" "Accept-Encoding"]
}

func ExampleVia() {
	header := http.Header{"Via": []string{
		"1.1 foo.example.com:8080 (corporate)",
		"2 bar.example.net",
	}}
	fmt.Printf("%q", Via(header))
	// Output: [{"HTTP/1.1" "foo.example.com:8080" "corporate"} {"HTTP/2" "bar.example.net" ""}]
}

var tabVia = []struct{
	header http.Header
	result []ViaEntry
}{
	{
		http.Header{"Via": []string{"1.0"}},
		nil,
	},
	{
		http.Header{"Via": []string{"1.0 foo"}},
		[]ViaEntry{{"HTTP/1.0", "foo", ""}},
	},
	{
		http.Header{"Via": []string{"1.0 \tfoo"}},
		[]ViaEntry{{"HTTP/1.0", "foo", ""}},
	},
	{
		http.Header{"Via": []string{"1.0 foo  "}},
		[]ViaEntry{{"HTTP/1.0", "foo", ""}},
	},
	{
		http.Header{"Via": []string{"1.0 foo  ,"}},
		[]ViaEntry{{"HTTP/1.0", "foo", ""}},
	},
	{
		http.Header{"Via": []string{"1.0 foo\t (comment)"}},
		[]ViaEntry{{"HTTP/1.0", "foo", "comment"}},
	},
	{
		http.Header{"Via": []string{
			"1.0 foo, 1.0 bar",
			"1.1 qux",
		}},
		[]ViaEntry{
			{"HTTP/1.0", "foo", ""}, {"HTTP/1.0", "bar", ""},
			{"HTTP/1.1", "qux", ""},
		},
	},
	{
		http.Header{"Via": []string{"1.0, 1.1 foo, 1.2, 1.3 bar"}},
		[]ViaEntry{{"HTTP/1.1", "foo", ""}, {"HTTP/1.3", "bar", ""}},
	},
	{
		http.Header{"Via": []string{"", ",FSTR/3 foo (comment (with) nesting)"}},
		[]ViaEntry{{"FSTR/3", "foo", "comment (with) nesting"}},
	},
	{
		http.Header{"Via": []string{`1.1 foo  (with \) quoting (and nesting))`}},
		[]ViaEntry{{"HTTP/1.1", "foo", "with ) quoting (and nesting)"}},
	},
	{
		http.Header{"Via": []string{
			"1.1 foo (unterminated",
			"1.1 bar",
		}},
		[]ViaEntry{
			{"HTTP/1.1", "foo", "unterminated"},
			{"HTTP/1.1", "bar", ""},
		},
	},
	{
		http.Header{"Via": []string{
			"1.1 foo (unterminated (with nesting",
			"1.1 bar",
		}},
		[]ViaEntry{
			{"HTTP/1.1", "foo", "unterminated (with nesting"},
			{"HTTP/1.1", "bar", ""},
		},
	},
	{
		http.Header{"Via": []string{
			`1.1 foo (unterminated with \quoting (and nesting`,
			"1.1 bar",
		}},
		[]ViaEntry{
			{"HTTP/1.1", "foo", "unterminated with quoting (and nesting"},
			{"HTTP/1.1", "bar", ""},
		},
	},
}

func TestVia(t *testing.T) {
	for _, tt := range tabVia {
		t.Run("", func(t *testing.T) {
			result := Via(tt.header)
			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("on %#v: got %#v, wanted %#v",
					tt.header, result, tt.result)
			}
		})
	}
}

func ExampleAddVia() {
	header := http.Header{"Via": []string{"1.0 foo"}}
	AddVia(header, []ViaEntry{{"HTTP/1.1", "bar", "internal"}})
	fmt.Printf("%q", header)
	// Output: map["Via":["1.0 foo" "1.1 bar (internal)"]]
}
