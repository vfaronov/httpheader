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
