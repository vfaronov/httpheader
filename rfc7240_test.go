package httpheader

import (
	"fmt"
	"net/http"
	"testing"
)

func ExamplePrefer() {
	header := http.Header{"Prefer": []string{
		"wait=10, respond-async",
		`check-spelling; lang="en-US, en-GB"`,
	}}
	prefer := Prefer(header)
	fmt.Printf("%+v\n%+v\n%+v\n",
		prefer["wait"],
		prefer["respond-async"],
		prefer["check-spelling"],
	)
	// Output: {Value:10 Params:map[]}
	// {Value: Params:map[]}
	// {Value: Params:map[lang:en-US, en-GB]}
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
		{
			// RFC 7240 page 5: ``If any preference is specified more than once,
			// only the first instance is to be considered.''
			http.Header{"Prefer": {`foo=bar, foo=baz;qux="a,b,c", xyzzy=123`}},
			map[string]Pref{"foo": {"bar", nil}, "xyzzy": {"123", nil}},
		},
		{
			// The same treatment is applied to preference parameters,
			// though not required there.
			http.Header{"Prefer": {"foo;bar=baz;bar=qux"}},
			map[string]Pref{"foo": {"", map[string]string{"bar": "baz"}}},
		},
		{
			http.Header{"Prefer": {"Handling=Lenient, Return=Minimal"}},
			map[string]Pref{
				"handling": {"lenient", nil},
				"return":   {"minimal", nil},
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
			// But we still parse it in consumeParam because it is
			// allowed elsewhere (e.g. in RFC 7230 transfer-extension).
			http.Header{"Prefer": {"foo = bar"}},
			map[string]Pref{"foo": {"bar", nil}},
		},
		{
			http.Header{"Prefer": {"foo = bar, baz = qux"}},
			map[string]Pref{"foo": {"bar", nil}, "baz": {"qux", nil}},
		},
		{
			http.Header{"Prefer": {"foo=???"}},
			map[string]Pref{"foo": {"???", nil}},
		},
		{
			http.Header{"Prefer": {"foo bar; baz, qux"}},
			map[string]Pref{
				"foo": {"", map[string]string{"bar": "", "baz": ""}},
				"qux": {"", nil},
			},
		},
		{
			http.Header{"Prefer": {";;;, foo=yes"}},
			map[string]Pref{"": {"", nil}, "foo": {"yes", nil}},
		},
		{
			http.Header{"Prefer": {"foo=bar=baz"}},
			map[string]Pref{"foo": {"bar", nil}},
		},
		{
			http.Header{"Prefer": {";foo=bar;baz=qux"}},
			map[string]Pref{"foo": {"bar", map[string]string{"baz": "qux"}}},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, Prefer(test.header))
		})
	}
}

func ExampleSetPrefer() {
	header := http.Header{}
	SetPrefer(header, map[string]Pref{
		"wait":           {Value: "10"},
		"respond-async":  {},
		"check-spelling": {Params: map[string]string{"lang": "en"}},
	})
	// Output:
}

func TestSetPrefer(t *testing.T) {
	tests := []struct {
		input  map[string]Pref
		result http.Header
	}{
		{
			map[string]Pref{},
			http.Header{},
		},
		{
			map[string]Pref{"respond-async": {}},
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
		{
			map[string]Pref{
				"foo": {"", map[string]string{"qux": `"quoted" xyzzy`}},
			},
			http.Header{"Prefer": {`foo;qux="\"quoted\" xyzzy"`}},
		},
		{
			map[string]Pref{
				"foo": {"", map[string]string{"qux": `"`}},
			},
			http.Header{"Prefer": {`foo;qux="\""`}},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			header := http.Header{}
			SetPrefer(header, test.input)
			checkGenerate(t, test.input, test.result, header)
		})
	}
}

func TestPreferRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetPrefer, Prefer,
		map[string]Pref{
			"lower token": {
				Value:  "quotable | empty",
				Params: map[string]string{"lower token": "quotable | empty"},
			},
		},
	)
}

func BenchmarkPreferSimple(b *testing.B) {
	header := http.Header{"Prefer": {"wait=10, respond-async"}}
	for i := 0; i < b.N; i++ {
		Prefer(header)
	}
}

func BenchmarkPreferComplex(b *testing.B) {
	header := http.Header{"Prefer": {
		`handling=lenient; ignore-errors="spelling, grammar, offensive-lang"`,
		`include-parameter="http://vocab.example/foo", wait=10, respond-async`,
	}}
	for i := 0; i < b.N; i++ {
		Prefer(header)
	}
}

func TestPreferenceApplied(t *testing.T) {
	tests := []struct {
		header http.Header
		result map[string]string
	}{
		// Valid headers.
		{
			http.Header{},
			nil,
		},
		{
			http.Header{"Preference-Applied": {"Handling=Lenient, Foo=Bar"}},
			map[string]string{"handling": "lenient", "foo": "Bar"},
		},
		{
			http.Header{"Preference-Applied": {`foo="bar, baz;qux=xyzzy"`}},
			map[string]string{"foo": "bar, baz;qux=xyzzy"},
		},
		{
			http.Header{"Preference-Applied": {"respond-async,depth-noroot"}},
			map[string]string{"respond-async": "", "depth-noroot": ""},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Preference-Applied": {"foo;bar=baz,qux=xyzzy"}},
			map[string]string{"foo": "", "qux": "xyzzy"},
		},
		{
			http.Header{"Preference-Applied": {`foo;bar="baz,qux",xyzzy`}},
			map[string]string{"foo": "", `qux"`: "", "xyzzy": ""},
		},
		{
			http.Header{"Preference-Applied": {"foo=bar=baz"}},
			map[string]string{"foo": "bar"},
		},
		{
			http.Header{"Preference-Applied": {";foo=bar;baz=qux"}},
			map[string]string{"foo": "bar"},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, PreferenceApplied(test.header))
		})
	}
}

func TestPreferenceAppliedRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetPreferenceApplied, PreferenceApplied,
		map[string]string{"lower token": "quotable | empty"},
	)
}
