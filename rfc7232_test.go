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
			nil,
		},
		{
			http.Header{"If-Match": {"*"}},
			[]EntityTag{AnyTag},
		},
		{
			http.Header{"If-Match": {`W/"",""`}},
			[]EntityTag{`W/""`, `""`},
		},
		{
			http.Header{"If-Match": {`"foo"`, `W/"bar"`, `"baz qux",W/",xyzzy,"`}},
			[]EntityTag{`"foo"`, `W/"bar"`, `"baz qux"`, `W/",xyzzy,"`},
		},
		{
			http.Header{"If-Match": {"W/\"\t\x81\x82\x83\\\""}},
			[]EntityTag{"W/\"\t\x81\x82\x83\\\""},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"If-Match": {`foo,bar,baz`}},
			[]EntityTag{`foo`, `bar`, `baz`},
		},
		{
			http.Header{"If-Match": {"\t, \tW/foo,bar"}},
			[]EntityTag{`W/foo`, `bar`},
		},
		{
			http.Header{"If-Match": {`"foo"`, `*`}},
			[]EntityTag{`"foo"`, `*`},
		},
		{
			http.Header{"If-Match": {`"foo`, `bar", baz`, `"`, `W/"qux`, `W/`}},
			[]EntityTag{`"foo`, `bar"`, `baz`, `"`, `W/"qux`},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, IfMatch(test.header))
		})
	}
}

func ExampleMatchWeak() {
	serverTag := MakeTag("v.62", true)
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

func TestMatch(t *testing.T) {
	tests := []struct {
		clientTags []EntityTag
		serverTag  EntityTag
		result     bool
	}{
		{[]EntityTag{}, `"foo"`, false},
		{[]EntityTag{`"foo"`}, `"foo"`, true},
		{[]EntityTag{`"foo"`}, `"Foo"`, false},
		{[]EntityTag{`W/"foo"`}, `"foo"`, false},
		{[]EntityTag{`"foo"`}, `W/"foo"`, false},
		{[]EntityTag{`"foo"`, `"bar"`}, `"bar"`, true},
		{[]EntityTag{`"foo"`, `W/"bar"`}, `W/"bar"`, false},
		{[]EntityTag{`foo`, `bar`}, `"bar"`, true},
		{[]EntityTag{`W/foo`, `W/bar`}, `bar`, false},
		{[]EntityTag{AnyTag}, `"foo"`, true},
		{[]EntityTag{`W/"bar"`, AnyTag}, `"foo"`, true},
		{[]EntityTag{`W/"foo"`, `"foo"`, `W/"bar"`}, `"foo"`, true},
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
		{[]EntityTag{}, `"foo"`, false},
		{[]EntityTag{`"foo"`}, `"foo"`, true},
		{[]EntityTag{`"foo"`}, `"Foo"`, false},
		{[]EntityTag{`W/"foo"`}, `"foo"`, true},
		{[]EntityTag{`"foo"`}, `W/"foo"`, true},
		{[]EntityTag{`"foo"`, `"bar"`}, `"bar"`, true},
		{[]EntityTag{`"foo"`, `W/"bar"`}, `W/"bar"`, true},
		{[]EntityTag{`foo`, `bar`}, `"bar"`, true},
		{[]EntityTag{`W/foo`, `W/bar`}, `bar`, true},
		{[]EntityTag{AnyTag}, `"foo"`, true},
		{[]EntityTag{`W/"bar"`, AnyTag}, `"foo"`, true},
		{[]EntityTag{`W/"foo"`, `"foo"`, `W/"bar"`}, `"foo"`, true},
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
	checkFuzz(t, "If-Match", IfMatch, SetIfMatch)
}

func TestEntityTagRoundTrip(t *testing.T) {
	testTags := []EntityTag{
		MakeTag("deadf00d", true),
		MakeTag("deadf00d", false),
		`deadf00d`, // malformed entity-tag without double quotes
	}
	for _, serverTag := range testTags {
		// Server sends ETag.
		h1 := http.Header{}
		SetETag(h1, serverTag)

		// Client receives and remembers this tag
		// in addition to a previously saved one.
		clientTags := []EntityTag{`W/"deadbeef"`}
		clientTags = append(clientTags, ETag(h1))

		// Client submits all the tags it has with the next request.
		h2 := http.Header{}
		SetIfNoneMatch(h2, clientTags)

		// Server extracts the client's tags and checks against its own.
		clientTags = IfNoneMatch(h2)
		match := MatchWeak(clientTags, serverTag)

		if !match {
			t.Errorf("serverTag = %#v: no match", serverTag)
		}
	}
}
