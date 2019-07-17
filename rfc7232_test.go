package httpheader

import (
	"fmt"
	"net/http"
	"testing"
)

func TestIfMatch(t *testing.T) {
	tests := []struct {
		header http.Header
		result []EntityTag
	}{
		// Valid headers.
		{
			http.Header{"If-Match": {""}},
			[]EntityTag{},
		},
		{
			http.Header{"If-Match": {"*"}},
			[]EntityTag{AnyTag},
		},
		{
			http.Header{"If-Match": {`W/"",""`}},
			[]EntityTag{{Weak: true}, {}},
		},
		{
			http.Header{"If-Match": {`"foo"`, `W/"bar"`, `"baz, qux",W/"xýzzý"`}},
			[]EntityTag{
				{Opaque: "foo"},
				{Weak: true, Opaque: "bar"},
				{Opaque: "baz, qux"},
				{Weak: true, Opaque: "xýzzý"},
			},
		},
		{
			http.Header{"If-Match": {"W/\"\t\x81\x82\x83\\\""}},
			[]EntityTag{{Weak: true, Opaque: "\t\x81\x82\x83\\"}},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"If-Match": {`foo,bar,baz`}},
			[]EntityTag{{}},
		},
		{
			http.Header{"If-Match": {"\t, \tW/foo,bar"}},
			[]EntityTag{{}},
		},
		{
			http.Header{"If-Match": {`"foo"`, `*`}},
			[]EntityTag{{Opaque: "foo"}, AnyTag},
		},
		{
			http.Header{"If-Match": {`"foo`, `bar", baz`, `"`, `W/"qux`, `W/`}},
			[]EntityTag{
				{Opaque: "foo"},
				{Opaque: ", baz"},
				{},
				{Weak: true, Opaque: "qux"},
				{Weak: true},
			},
		},
		{
			http.Header{"If-Match": {"W"}},
			[]EntityTag{{}},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, IfMatch(test.header))
		})
	}
}

func ExampleIfNoneMatch() {
	serverTag := EntityTag{Weak: true, Opaque: "v.62"}
	request := http.Header{"If-None-Match": {`W/"v.57", W/"v.62", "f09a3ccd"`}}
	response := http.Header{}
	if MatchWeak(IfNoneMatch(request), serverTag) {
		fmt.Println("Status: 304 Not Modified")
		SetETag(response, serverTag)
		return
	}
	fmt.Println("Status: 200 OK")
	SetETag(response, serverTag)
	// Output: Status: 304 Not Modified
}

func TestSetETag(t *testing.T) {
	tests := []struct {
		input  EntityTag
		result http.Header
	}{
		{
			EntityTag{},
			http.Header{"Etag": {`""`}},
		},
		{
			EntityTag{Opaque: "foo"},
			http.Header{"Etag": {`"foo"`}},
		},
		{
			EntityTag{Weak: true, Opaque: "foo, bar"},
			http.Header{"Etag": {`W/"foo, bar"`}},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			header := http.Header{}
			SetETag(header, test.input)
			checkGenerate(t, test.input, test.result, header)
		})
	}
}

func TestMatch(t *testing.T) {
	tests := []struct {
		clientTags []EntityTag
		serverTag  EntityTag
		result     bool
	}{
		{
			[]EntityTag{},
			EntityTag{Opaque: "foo"},
			false,
		},
		{
			[]EntityTag{{Opaque: "foo"}},
			EntityTag{Opaque: "foo"},
			true,
		},
		{
			[]EntityTag{{Opaque: "foo"}},
			EntityTag{Opaque: "Foo"},
			false,
		},
		{
			[]EntityTag{{Weak: true, Opaque: "foo"}},
			EntityTag{Opaque: "foo"},
			false,
		},
		{
			[]EntityTag{{Opaque: "foo"}},
			EntityTag{Weak: true, Opaque: "foo"},
			false,
		},
		{
			[]EntityTag{{Opaque: "foo"}, {Opaque: "bar"}},
			EntityTag{Opaque: "bar"},
			true,
		},
		{
			[]EntityTag{{Opaque: "foo"}, {Weak: true, Opaque: "bar"}},
			EntityTag{Weak: true, Opaque: "bar"},
			false,
		},
		{
			[]EntityTag{AnyTag},
			EntityTag{Opaque: "foo"},
			true,
		},
		{
			[]EntityTag{{Weak: true, Opaque: "bar"}, AnyTag},
			EntityTag{Opaque: "foo"},
			true,
		},
		{
			[]EntityTag{
				{Weak: true, Opaque: "foo"},
				{Opaque: "foo"},
				{Weak: true, Opaque: "bar"},
			},
			EntityTag{Opaque: "foo"},
			true,
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			actual := Match(test.clientTags, test.serverTag)
			if actual != test.result {
				t.Errorf("Match(%#v, %#v) = %v, expected %v",
					test.clientTags, test.serverTag, actual, test.result)
			}
		})
	}
}

func TestMatchWeak(t *testing.T) {
	tests := []struct {
		clientTags []EntityTag
		serverTag  EntityTag
		result     bool
	}{
		{
			[]EntityTag{},
			EntityTag{Opaque: "foo"},
			false,
		},
		{
			[]EntityTag{{Opaque: "foo"}},
			EntityTag{Opaque: "foo"},
			true,
		},
		{
			[]EntityTag{{Opaque: "foo"}},
			EntityTag{Opaque: "Foo"},
			false,
		},
		{
			[]EntityTag{{Weak: true, Opaque: "foo"}},
			EntityTag{Opaque: "foo"},
			true,
		},
		{
			[]EntityTag{{Opaque: "foo"}},
			EntityTag{Weak: true, Opaque: "foo"},
			true,
		},
		{
			[]EntityTag{{Opaque: "foo"}, {Opaque: "bar"}},
			EntityTag{Opaque: "bar"},
			true,
		},
		{
			[]EntityTag{{Opaque: "foo"}, {Weak: true, Opaque: "bar"}},
			EntityTag{Weak: true, Opaque: "bar"},
			true,
		},
		{
			[]EntityTag{AnyTag},
			EntityTag{Opaque: "foo"},
			true,
		},
		{
			[]EntityTag{{Weak: true, Opaque: "bar"}, AnyTag},
			EntityTag{Opaque: "foo"},
			true,
		},
		{
			[]EntityTag{
				{Weak: true, Opaque: "foo"},
				{Opaque: "foo"},
				{Weak: true, Opaque: "bar"},
			},
			EntityTag{Opaque: "foo"},
			true,
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			actual := MatchWeak(test.clientTags, test.serverTag)
			if actual != test.result {
				t.Errorf("MatchWeak(%#v, %#v) = %v, expected %v",
					test.clientTags, test.serverTag, actual, test.result)
			}
		})
	}
}

func TestIfMatchFuzz(t *testing.T) {
	checkFuzz(t, "If-Match", IfMatch, nil)
}
