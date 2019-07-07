package httpheader

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"
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

func TestAllowFuzz(t *testing.T) {
	checkFuzz(t, "Allow", Allow, SetAllow)
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
			[]string{""},
		},
		{
			http.Header{"Allow": {";;;,GET"}},
			[]string{"", "GET"},
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

func TestVaryFuzz(t *testing.T) {
	checkFuzz(t, "Vary", Vary, SetVary)
}

func TestVaryRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetVary, Vary, func(r *rand.Rand) interface{} {
		return mkSlice(r, mkHeaderName)
	})
}

func ExampleUserAgent() {
	header := http.Header{"User-Agent": {"Mozilla/5.0 (compatible) Chrome/123"}}
	fmt.Printf("%+v", UserAgent(header))
	// Output: [{Name:Mozilla Version:5.0 Comment:compatible} {Name:Chrome Version:123 Comment:}]
}

func TestUserAgent(t *testing.T) {
	// Most of the tests are in TestServer. Here, just check a few real-world
	// examples from browsers, notorious for their exuberant User-Agent strings.
	tests := []struct {
		header http.Header
		result []Product
	}{
		{
			http.Header{"User-Agent": {"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:67.0) Gecko/20100101 Firefox/67.0"}},
			[]Product{
				{"Mozilla", "5.0", "X11; Ubuntu; Linux x86_64; rv:67.0"},
				{"Gecko", "20100101", ""},
				{"Firefox", "67.0", ""},
			},
		},
		{
			http.Header{"User-Agent": {"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/75.0.3770.90 Chrome/75.0.3770.90 Safari/537.36"}},
			[]Product{
				{"Mozilla", "5.0", "X11; Linux x86_64"},
				{"AppleWebKit", "537.36", "KHTML, like Gecko"},
				{"Ubuntu", "", ""},
				{"Chromium", "75.0.3770.90", ""},
				{"Chrome", "75.0.3770.90", ""},
				{"Safari", "537.36", ""},
			},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, UserAgent(test.header))
		})
	}
}

func TestUserAgentFuzz(t *testing.T) {
	checkFuzz(t, "User-Agent", UserAgent, SetUserAgent)
}

func ExampleSetUserAgent() {
	header := http.Header{}
	SetUserAgent(header, []Product{
		{Name: "My-App", Version: "1.2.3", Comment: "example.com"},
		{Name: "Go-http-client"},
	})
	fmt.Printf("%q", header)
	// Output: map["User-Agent":["My-App/1.2.3 (example.com) Go-http-client"]]
}

func TestServer(t *testing.T) {
	tests := []struct {
		header http.Header
		result []Product
	}{
		// Valid headers.
		{
			http.Header{"Server": {"nginx"}},
			[]Product{{"nginx", "", ""}},
		},
		{
			http.Header{"Server": {"nginx/1.17.1"}},
			[]Product{{"nginx", "1.17.1", ""}},
		},
		{
			http.Header{"Server": {"nginx (Ubuntu)"}},
			[]Product{{"nginx", "", "Ubuntu"}},
		},
		{
			http.Header{"Server": {"nginx/1.17.1 (Ubuntu)"}},
			[]Product{{"nginx", "1.17.1", "Ubuntu"}},
		},
		{
			http.Header{"Server": {"nginx (Ubuntu) (i386)"}},
			[]Product{{"nginx", "", "Ubuntu; i386"}},
		},
		{
			http.Header{"Server": {"nginx/1.17.1 (Ubuntu) (i386)"}},
			[]Product{{"nginx", "1.17.1", "Ubuntu; i386"}},
		},
		{
			http.Header{"Server": {"nginx (Ubuntu) Linux"}},
			[]Product{{"nginx", "", "Ubuntu"}, {"Linux", "", ""}},
		},
		{
			http.Header{"Server": {"nginx (Ubuntu) Linux (i386)"}},
			[]Product{{"nginx", "", "Ubuntu"}, {"Linux", "", "i386"}},
		},
		{
			http.Header{"Server": {"nginx/1.17.1 (Ubuntu) (Lua) Linux (i386)"}},
			[]Product{{"nginx", "1.17.1", "Ubuntu; Lua"}, {"Linux", "", "i386"}},
		},
		{
			http.Header{"Server": {"nginx/1.17.1  (Ubuntu) (Lua)\tLinux  (i386)"}},
			[]Product{{"nginx", "1.17.1", "Ubuntu; Lua"}, {"Linux", "", "i386"}},
		},
		{
			http.Header{"Server": {"uWSGI nginx Linux"}},
			[]Product{{"uWSGI", "", ""}, {"nginx", "", ""}, {"Linux", "", ""}},
		},
		{
			http.Header{"Server": {"CERN/3.0 libwww/2.17"}},
			[]Product{{"CERN", "3.0", ""}, {"libwww", "2.17", ""}},
		},
		{
			// Syntactially valid, although not what the sender intended.
			http.Header{"Server": {"foo 1.2.3"}},
			[]Product{{"foo", "", ""}, {"1.2.3", "", ""}},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			// Server is not comma-delimited, so cannot be split into two fields.
			http.Header{"Server": {"foo", "bar"}},
			[]Product{{"foo", "", ""}},
		},
		{
			http.Header{"Server": {"foo/bar/baz"}},
			[]Product{{"foo", "bar/baz", ""}},
		},
		{
			http.Header{"Server": {"foo (comment) (unterminated"}},
			[]Product{{"foo", "", "comment; unterminated"}},
		},
		{
			http.Header{"Server": {"Jetty(8.1.13.v20130916)"}},
			[]Product{{"Jetty(8.1.13.v20130916)", "", ""}},
		},
		{
			http.Header{"Server": {"foo, bar, baz"}},
			[]Product{{"foo", "", ""}, {"bar", "", ""}, {"baz", "", ""}},
		},
		{
			http.Header{"Server": {"foo; bar; baz"}},
			[]Product{{"foo", "", ""}, {"bar", "", ""}, {"baz", "", ""}},
		},
		{
			http.Header{"Server": {"foo=1.2.3"}},
			[]Product{{"foo", "", ""}, {"1.2.3", "", ""}},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, Server(test.header))
		})
	}
}

func TestServerFuzz(t *testing.T) {
	checkFuzz(t, "Server", Server, SetServer)
}

func TestServerRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetServer, Server, func(r *rand.Rand) interface{} {
		return mkSlice(r, func(r *rand.Rand) interface{} {
			return Product{
				Name:    mkToken(r).(string),
				Version: mkMaybeToken(r).(string),
				Comment: mkString(r).(string),
			}
		})
	})
}

func ExampleRetryAfter() {
	header := http.Header{
		"Date":        {"Sun, 07 Jul 2019 08:03:32 GMT"},
		"Retry-After": {"180"},
	}
	fmt.Print(RetryAfter(header))
	// Output: 2019-07-07 08:06:32 +0000 UTC
}

func TestRetryAfter(t *testing.T) {
	tests := []struct {
		header http.Header
		result time.Time
	}{
		// Valid headers.
		{
			http.Header{"Retry-After": {"Sun, 07 Jul 2019 08:06:01 GMT"}},
			time.Date(2019, time.July, 7, 8, 6, 1, 0, time.UTC),
		},
		{
			http.Header{
				"Date":        {"Sun, 07 Jul 2019 08:06:01 GMT"},
				"Retry-After": {"600"},
			},
			time.Date(2019, time.July, 7, 8, 16, 1, 0, time.UTC),
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Retry-After": {"whenever"}},
			time.Time{},
		},
		{
			http.Header{"Retry-After": {"Sun, 37 Jul 2019 08:06:01 GMT"}},
			time.Time{},
		},
		{
			http.Header{
				"Date":        {"Sun, 07 Jul 2019 08:06:01 GMT"},
				"Retry-After": {"60s"},
			},
			time.Time{},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, RetryAfter(test.header))
		})
	}
}

func TestRetryAfterCurrentTime(t *testing.T) {
	header := http.Header{"Retry-After": {"300"}}
	now := time.Now()
	target := now.Add(300 * time.Second)
	parsed := RetryAfter(header)
	if parsed.Before(target) || parsed.After(target.Add(1*time.Second)) {
		t.Fatalf("got %v, expected within 1s of %v", parsed, target)
	}
	header["Date"] = []string{"some invalid value"}
	parsed = RetryAfter(header)
	if parsed.Before(target) || parsed.After(target.Add(1*time.Second)) {
		t.Fatalf("got %v, expected within 1s of %v", parsed, target)
	}
}

func TestRetryAfterFuzz(t *testing.T) {
	checkFuzz(t, "Retry-After", RetryAfter, SetRetryAfter)
}

func TestRetryAfterRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetRetryAfter, RetryAfter, mkDate)
}

func ExampleContentType() {
	header := http.Header{"Content-Type": {"Text/HTML;Charset=UTF-8"}}
	fmt.Print(ContentType(header))
	// Output: {text/html map[charset:UTF-8]}
}

func TestContentType(t *testing.T) {
	tests := []struct {
		header http.Header
		result Par
	}{
		// Valid headers.
		{
			http.Header{"Content-Type": {"text/html"}},
			Par{"text/html", nil},
		},
		{
			http.Header{"Content-Type": {"Text/HTML"}},
			Par{"text/html", nil},
		},
		{
			http.Header{"Content-Type": {"application/vnd.api+json"}},
			Par{"application/vnd.api+json", nil},
		},
		{
			http.Header{"Content-Type": {"text/html;charset=utf-8"}},
			Par{"text/html", map[string]string{"charset": "utf-8"}},
		},
		{
			http.Header{"Content-Type": {"text/html; charset=utf-8"}},
			Par{"text/html", map[string]string{"charset": "utf-8"}},
		},
		{
			http.Header{"Content-Type": {`Text/HTML; Charset="utf-8"`}},
			Par{"text/html", map[string]string{"charset": "utf-8"}},
		},
		{
			http.Header{"Content-Type": {"text/html\t; \t charset=utf-8"}},
			Par{"text/html", map[string]string{"charset": "utf-8"}},
		},
		{
			http.Header{"Content-Type": {
				`application/foo; quux="xyz\\zy";bar=baz`,
			}},
			Par{"application/foo", map[string]string{
				"bar":  "baz",
				"quux": `xyz\zy`,
			}},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Content-Type": {""}},
			Par{},
		},
		{
			http.Header{"Content-Type": {"text"}},
			Par{"text", nil},
		},
		{
			http.Header{"Content-Type": {"text/plain/html"}},
			Par{"text/plain/html", nil},
		},
		{
			http.Header{"Content-Type": {"text+html"}},
			Par{"text+html", nil},
		},
		{
			http.Header{"Content-Type": {"text/html;;"}},
			Par{"text/html", nil},
		},
		{
			http.Header{"Content-Type": {"text/html;;charset=utf-8"}},
			Par{"text/html", map[string]string{"charset": "utf-8"}},
		},
		{
			http.Header{"Content-Type": {"text/html ; ; ; charset=utf-8"}},
			Par{"text/html", map[string]string{"charset": "utf-8"}},
		},
		{
			http.Header{"Content-Type": {"text/html; w3c; charset=utf-8"}},
			Par{"text/html", map[string]string{
				"w3c":     "",
				"charset": "utf-8",
			}},
		},
		{
			http.Header{"Content-Type": {"text/html,charset=utf-8"}},
			Par{"text/html", nil},
		},
		{
			http.Header{"Content-Type": {"text/html charset=utf-8"}},
			Par{"text/html", nil},
		},
		{
			http.Header{"Content-Type": {"charset=utf-8"}},
			Par{"charset", nil},
		},
		{
			http.Header{"Content-Type": {"text/html=utf-8"}},
			Par{"text/html", nil},
		},
		{
			http.Header{"Content-Type": {"text/html; charset = utf-8"}},
			Par{"text/html", map[string]string{"charset": "utf-8"}},
		},
		{
			http.Header{"Content-Type": {`text/html;charset  = "utf-8"`}},
			Par{"text/html", map[string]string{"charset": "utf-8"}},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, ContentType(test.header))
		})
	}
}

func TestContentTypeFuzz(t *testing.T) {
	checkFuzz(t, "Content-Type", ContentType, SetContentType)
}

func TestContentTypeRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetContentType, ContentType, func(r *rand.Rand) interface{} {
		return Par{
			Item:   mkLowerToken(r).(string) + "/" + mkLowerToken(r).(string),
			Params: mkMap(r, mkLowerToken, mkString).(map[string]string),
		}
	})
}

func ExampleSetContentType() {
	header := http.Header{}
	SetContentType(header, Par{
		Item:   "text/html",
		Params: map[string]string{"charset": "utf-8"},
	})
}
