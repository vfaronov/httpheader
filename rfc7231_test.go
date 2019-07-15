package httpheader

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func ExampleAllow() {
	header := http.Header{"Allow": {"GET, HEAD, OPTIONS"}}
	fmt.Print(Allow(header))
	// Output: [GET HEAD OPTIONS]
}

func ExampleSetAllow() {
	header := http.Header{}
	SetAllow(header, []string{"GET", "HEAD", "OPTIONS"})
}

func TestAllowFuzz(t *testing.T) {
	checkFuzz(t, "Allow", Allow, SetAllow)
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

func BenchmarkAllow(b *testing.B) {
	header := http.Header{"Allow": {"GET, HEAD, POST, PUT, DELETE", "OPTIONS, TRACE"}}
	for i := 0; i < b.N; i++ {
		Allow(header)
	}
}

func ExampleVary() {
	header := http.Header{"Vary": {"cookie, accept-encoding"}}
	vary := Vary(header)
	if vary["Accept-Encoding"] || vary["*"] {
		// this response varies by the client's acceptable encoding
	}
}

func TestVary(t *testing.T) {
	tests := []struct {
		header http.Header
		result map[string]bool
	}{
		{
			http.Header{"Vary": {"user-agent"}},
			map[string]bool{"User-Agent": true},
		},
		{
			http.Header{"Vary": {"accept,prefer"}},
			map[string]bool{"Accept": true, "Prefer": true},
		},
		{
			http.Header{"Vary": {"*"}},
			map[string]bool{"*": true},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, Vary(test.header))
		})
	}
}

func ExampleAddVary() {
	header := http.Header{}
	AddVary(header, map[string]bool{"Accept": true, "Accept-Encoding": true})
	// Output:
}

func TestVaryFuzz(t *testing.T) {
	checkFuzz(t, "Vary", Vary, SetVary)
}

func BenchmarkVary(b *testing.B) {
	header := http.Header{"Vary": {"Accept, Accept-Language, Accept-Encoding, Prefer", "User-Agent, Cookie"}}
	for i := 0; i < b.N; i++ {
		Vary(header)
	}
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
	checkRoundTrip(t, SetServer, Server,
		[]Product{{
			Name:    "token",
			Version: "token | empty",
			Comment: "quotable | empty",
		}},
	)
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

func ExampleContentType() {
	header := http.Header{"Content-Type": {"Text/HTML;Charset=UTF-8"}}
	mtype, params := ContentType(header)
	fmt.Println(mtype, params)
	// Output: text/html map[charset:UTF-8]
}

func TestContentType(t *testing.T) {
	tests := []struct {
		header http.Header
		mtype  string
		params map[string]string
	}{
		// Valid headers.
		{
			http.Header{"Content-Type": {"text/html"}},
			"text/html",
			nil,
		},
		{
			http.Header{"Content-Type": {"Text/HTML"}},
			"text/html",
			nil,
		},
		{
			http.Header{"Content-Type": {"application/vnd.api+json"}},
			"application/vnd.api+json",
			nil,
		},
		{
			http.Header{"Content-Type": {"text/html;charset=utf-8"}},
			"text/html",
			map[string]string{"charset": "utf-8"},
		},
		{
			http.Header{"Content-Type": {"text/html; charset=utf-8"}},
			"text/html",
			map[string]string{"charset": "utf-8"},
		},
		{
			http.Header{"Content-Type": {`Text/HTML; Charset="utf-8"`}},
			"text/html",
			map[string]string{"charset": "utf-8"},
		},
		{
			http.Header{"Content-Type": {"text/html\t; \t charset=utf-8"}},
			"text/html",
			map[string]string{"charset": "utf-8"},
		},
		{
			http.Header{"Content-Type": {
				`application/foo; quux="xyz\\zy";bar=baz`,
			}},
			"application/foo",
			map[string]string{
				"bar":  "baz",
				"quux": `xyz\zy`,
			},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Content-Type": {""}},
			"",
			nil,
		},
		{
			http.Header{"Content-Type": {"text"}},
			"text",
			nil,
		},
		{
			http.Header{"Content-Type": {"text/plain/html"}},
			"text/plain/html",
			nil,
		},
		{
			http.Header{"Content-Type": {"text+html"}},
			"text+html",
			nil,
		},
		{
			http.Header{"Content-Type": {"text/html;;"}},
			"text/html",
			nil,
		},
		{
			http.Header{"Content-Type": {"text/html;;charset=utf-8"}},
			"text/html",
			map[string]string{"charset": "utf-8"},
		},
		{
			http.Header{"Content-Type": {"text/html ; ; ; charset=utf-8"}},
			"text/html",
			map[string]string{"charset": "utf-8"},
		},
		{
			http.Header{"Content-Type": {"text/html; w3c; charset=utf-8"}},
			"text/html",
			map[string]string{
				"w3c":     "",
				"charset": "utf-8",
			},
		},
		{
			http.Header{"Content-Type": {"text/html,charset=utf-8"}},
			"text/html",
			nil,
		},
		{
			http.Header{"Content-Type": {"text/html charset=utf-8"}},
			"text/html",
			map[string]string{"charset": "utf-8"},
		},
		{
			http.Header{"Content-Type": {"charset=utf-8"}},
			"charset",
			nil,
		},
		{
			http.Header{"Content-Type": {"text/html=utf-8"}},
			"text/html",
			nil,
		},
		{
			http.Header{"Content-Type": {"text/html; charset = utf-8"}},
			"text/html",
			map[string]string{"charset": "utf-8"},
		},
		{
			http.Header{"Content-Type": {`text/html;charset  = "utf-8"`}},
			"text/html",
			map[string]string{"charset": "utf-8"},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			mtype, params := ContentType(test.header)
			checkParse(t, test.header, test.mtype, mtype, test.params, params)
		})
	}
}

func TestContentTypeFuzz(t *testing.T) {
	checkFuzz(t, "Content-Type", ContentType, SetContentType)
}

func TestContentTypeRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetContentType, ContentType,
		"lower token/token",
		map[string]string{"lower token": "quotable"},
	)
}

func ExampleSetContentType() {
	header := http.Header{}
	SetContentType(header, "text/html", map[string]string{"charset": "utf-8"})
}

func ExampleAccept() {
	header := http.Header{"Accept": {"Text/HTML; charset=utf-8; q=1; validate=yes"}}
	fmt.Printf("%+v", Accept(header))
	// Output: [{Type:text/html Q:1 Params:map[charset:utf-8] Ext:map[validate:yes]}]
}

func TestAccept(t *testing.T) {
	tests := []struct {
		header http.Header
		result []AcceptElem
	}{
		// Valid headers.
		{
			http.Header{"Accept": {"*/*"}},
			[]AcceptElem{{Type: "*/*", Q: 1}},
		},
		{
			http.Header{"Accept": {"text/html"}},
			[]AcceptElem{{Type: "text/html", Q: 1}},
		},
		{
			http.Header{"Accept": {"text/html, text/plain"}},
			[]AcceptElem{{Type: "text/html", Q: 1}, {Type: "text/plain", Q: 1}},
		},
		{
			http.Header{"Accept": {"text/html, text/*, */*"}},
			[]AcceptElem{
				{Type: "text/html", Q: 1},
				{Type: "text/*", Q: 1},
				{Type: "*/*", Q: 1},
			},
		},
		{
			http.Header{"Accept": {"text/html, text/*;q=0.25"}},
			[]AcceptElem{{Type: "text/html", Q: 1}, {Type: "text/*", Q: 0.25}},
		},
		{
			http.Header{"Accept": {"text/html;q=0.25"}},
			[]AcceptElem{{Type: "text/html", Q: 0.25}},
		},
		{
			http.Header{"Accept": {"text/html; q=1, text/*; q=0.5"}},
			[]AcceptElem{{Type: "text/html", Q: 1}, {Type: "text/*", Q: 0.5}},
		},
		{
			http.Header{"Accept": {"text/html;charset=utf-8, text/*;q=0.5"}},
			[]AcceptElem{
				{
					Type:   "text/html",
					Q:      1,
					Params: map[string]string{"charset": "utf-8"},
				},
				{
					Type: "text/*",
					Q:    0.5,
				},
			},
		},
		{
			http.Header{"Accept": {"text/html;charset=utf-8;q=0.25"}},
			[]AcceptElem{
				{
					Type:   "text/html",
					Q:      0.25,
					Params: map[string]string{"charset": "utf-8"},
				},
			},
		},
		{
			http.Header{"Accept": {"text/html;\tcharset=utf-8;  q=0.25"}},
			[]AcceptElem{
				{
					Type:   "text/html",
					Q:      0.25,
					Params: map[string]string{"charset": "utf-8"},
				},
			},
		},
		{
			http.Header{"Accept": {`Text/HTML;  Charset="UTF-8"`}},
			[]AcceptElem{
				{
					Type:   "text/html",
					Q:      1,
					Params: map[string]string{"charset": "UTF-8"},
				},
			},
		},
		{
			http.Header{"Accept": {"text/html;q=1;foo=bar"}},
			[]AcceptElem{
				{
					Type: "text/html",
					Q:    1,
					Ext:  map[string]string{"foo": "bar"},
				},
			},
		},
		{
			http.Header{"Accept": {`Text/HTML; Foo="bar; baz; qux"`}},
			[]AcceptElem{
				{
					Type:   "text/html",
					Q:      1,
					Params: map[string]string{"foo": "bar; baz; qux"},
				},
			},
		},
		{
			http.Header{"Accept": {`Text/HTML; Q=1; Foo="Bar; Baz; Qux"`}},
			[]AcceptElem{
				{
					Type: "text/html",
					Q:    1,
					Ext:  map[string]string{"foo": "Bar; Baz; Qux"},
				},
			},
		},
		{
			http.Header{"Accept": {`Text/HTML; Foo="Bar; Baz"; Q=1; Qux="Xyzzy"`}},
			[]AcceptElem{
				{
					Type:   "text/html",
					Q:      1,
					Params: map[string]string{"foo": "Bar; Baz"},
					Ext:    map[string]string{"qux": "Xyzzy"},
				},
			},
		},
		{
			http.Header{"Accept": {`Text/HTML; Q=1; Foo; Bar=Baz`}},
			[]AcceptElem{
				{
					Type: "text/html",
					Q:    1,
					Ext:  map[string]string{"foo": "", "bar": "Baz"},
				},
			},
		},
		{
			http.Header{"Accept": {`Text/HTML; Q=1; Foo=Bar; Baz`}},
			[]AcceptElem{
				{
					Type: "text/html",
					Q:    1,
					Ext:  map[string]string{"foo": "Bar", "baz": ""},
				},
			},
		},
		{
			http.Header{"Accept": {`Text/HTML; Q=1; Foo=Bar; Baz=""`}},
			[]AcceptElem{
				{
					Type: "text/html",
					Q:    1,
					Ext:  map[string]string{"foo": "Bar", "baz": ""},
				},
			},
		},
		{
			http.Header{"Accept": {`Text/HTML; Q=1; Foo=Bar; Baz, */*; Q=0.25`}},
			[]AcceptElem{
				{
					Type: "text/html",
					Q:    1,
					Ext:  map[string]string{"foo": "Bar", "baz": ""},
				},
				{
					Type: "*/*",
					Q:    0.25,
				},
			},
		},
		{
			http.Header{"Accept": {
				"text/plain; q=0.5, text/html",
				"text/x-dvi; q=0.8, text/x-c",
			}},
			[]AcceptElem{
				{Type: "text/plain", Q: 0.5},
				{Type: "text/html", Q: 1},
				{Type: "text/x-dvi", Q: 0.8},
				{Type: "text/x-c", Q: 1},
			},
		},
		{
			http.Header{"Accept": {
				"application/json",
				"application/vnd.api+json",
			}},
			[]AcceptElem{
				{Type: "application/json", Q: 1},
				{Type: "application/vnd.api+json", Q: 1},
			},
		},
		{
			http.Header{"Accept": {
				"text/*, text/plain, text/plain;format=flowed, */*",
			}},
			[]AcceptElem{
				{
					Type: "text/*",
					Q:    1,
				},
				{
					Type: "text/plain",
					Q:    1,
				},
				{
					Type:   "text/plain",
					Q:      1,
					Params: map[string]string{"format": "flowed"},
				},
				{
					Type: "*/*",
					Q:    1,
				},
			},
		},
		{
			http.Header{"Accept": {"*/* ; q=1 ; whatever"}},
			[]AcceptElem{
				{Type: "*/*", Q: 1, Ext: map[string]string{"whatever": ""}},
			},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Accept": {""}},
			[]AcceptElem{},
		},
		{
			http.Header{"Accept": {"text"}},
			[]AcceptElem{{Type: "text", Q: 1}},
		},
		{
			http.Header{"Accept": {"text html"}},
			[]AcceptElem{
				{
					Type:   "text",
					Q:      1,
					Params: map[string]string{"html": ""},
				},
			},
		},
		{
			http.Header{"Accept": {"text/html/plain"}},
			[]AcceptElem{{Type: "text/html/plain", Q: 1}},
		},
		{
			http.Header{"Accept": {"text/html; text/plain"}},
			[]AcceptElem{
				{
					Type:   "text/html",
					Q:      1,
					Params: map[string]string{"text/plain": ""},
				},
			},
		},
		{
			http.Header{"Accept": {"text/html ; charset = utf-8 ; "}},
			[]AcceptElem{
				{
					Type:   "text/html",
					Q:      1,
					Params: map[string]string{"charset": "utf-8"},
				},
			},
		},
		{
			http.Header{"Accept": {`application/xml, */*;q="0.1"`}},
			[]AcceptElem{
				{Type: "application/xml", Q: 1},
				{Type: "*/*", Q: 0.1},
			},
		},
		{
			http.Header{"Accept": {"text/plain; prose, text/plain; q=0.5"}},
			[]AcceptElem{
				{Type: "text/plain", Q: 1, Params: map[string]string{"prose": ""}},
				{Type: "text/plain", Q: 0.5},
			},
		},
		{
			http.Header{"Accept": {"text/plain; charset=, text/html"}},
			[]AcceptElem{
				{
					Type:   "text/plain",
					Q:      1,
					Params: map[string]string{"charset": ""},
				},
				{
					Type: "text/html",
					Q:    1,
				},
			},
		},
		{
			http.Header{"Accept": {
				"text/plain; charset=; format=flowed, text/html",
			}},
			[]AcceptElem{
				{
					Type:   "text/plain",
					Q:      1,
					Params: map[string]string{"charset": "", "format": "flowed"},
				},
				{
					Type: "text/html",
					Q:    1,
				},
			},
		},
		{
			http.Header{"Accept": {"text/plain;;;charset=utf-8;,text/html"}},
			[]AcceptElem{
				{
					Type:   "text/plain",
					Q:      1,
					Params: map[string]string{"charset": "utf-8"},
				},
				{
					Type: "text/html",
					Q:    1,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, Accept(test.header))
		})
	}
}

func BenchmarkAcceptSimple(b *testing.B) {
	header := http.Header{"Accept": {"application/json, text/xml;q=0.5"}}
	for i := 0; i < b.N; i++ {
		Accept(header)
	}
}

func BenchmarkAcceptComplex(b *testing.B) {
	header := http.Header{"Accept": {`application/x.my-custom+json;q=1;full;linkage=no, application/vnd.api+json;profile="http://example.com/last-modified";q=0.9, application/vnd.api+json;q=0.8, application/json;q=0.6, text/*;q=0.1`}}
	for i := 0; i < b.N; i++ {
		Accept(header)
	}
}

func ExampleSetAccept() {
	header := http.Header{}
	SetAccept(header, []AcceptElem{
		{Type: "application/json", Q: 1},
		{Type: "*/*", Q: 0.1},
	})
}

func TestSetAccept(t *testing.T) {
	tests := []struct {
		input  []AcceptElem
		result http.Header
	}{
		{
			nil,
			http.Header{},
		},
		{
			// Accept permits an empty list.
			[]AcceptElem{},
			http.Header{"Accept": {""}},
		},
		{
			[]AcceptElem{{Type: "text/html", Q: 1}},
			http.Header{"Accept": {"text/html"}},
		},
		{
			[]AcceptElem{{Type: "image/webp", Q: 1}, {Type: "*/*", Q: 1}},
			http.Header{"Accept": {"image/webp, */*"}},
		},
		{
			[]AcceptElem{{Type: "text/html"}},
			http.Header{"Accept": {"text/html;q=0"}},
		},
		{
			[]AcceptElem{{Type: "text/html", Q: 0.1234567}},
			http.Header{"Accept": {"text/html;q=0.123"}},
		},
		{
			[]AcceptElem{{Type: "text/html", Q: 0.001}},
			http.Header{"Accept": {"text/html;q=0.001"}},
		},
		{
			[]AcceptElem{
				{
					Type:   "text/plain",
					Q:      1,
					Params: map[string]string{"foo": "bar"},
				}},
			http.Header{"Accept": {"text/plain;foo=bar"}},
		},
		{
			[]AcceptElem{
				{
					Type:   "text/plain",
					Q:      0.5,
					Params: map[string]string{"foo": "bar"},
				}},
			http.Header{"Accept": {"text/plain;foo=bar;q=0.5"}},
		},
		{
			[]AcceptElem{
				{
					Type:   "text/plain",
					Q:      1,
					Params: map[string]string{"foo": "bar; baz"},
				}},
			http.Header{"Accept": {`text/plain;foo="bar; baz"`}},
		},
		{
			[]AcceptElem{
				{
					Type:   "text/plain",
					Q:      1,
					Params: map[string]string{"foo": ""},
				}},
			http.Header{"Accept": {`text/plain;foo=""`}},
		},
		{
			[]AcceptElem{
				{
					Type: "text/plain",
					Q:    1,
					Ext:  map[string]string{"foo": "bar"},
				}},
			http.Header{"Accept": {"text/plain;q=1;foo=bar"}},
		},
		{
			[]AcceptElem{
				{
					Type: "text/plain",
					Q:    1,
					Ext:  map[string]string{"foo": ""},
				}},
			http.Header{"Accept": {"text/plain;q=1;foo"}},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			header := http.Header{}
			SetAccept(header, test.input)
			checkGenerate(t, test.input, test.result, header)
		})
	}
}

func TestAcceptFuzz(t *testing.T) {
	checkFuzz(t, "Accept", Accept, SetAccept)
}

func TestAcceptRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetAccept, Accept,
		[]AcceptElem{{
			Type:   "lower token/token",
			Q:      0.999,
			Params: map[string]string{"lower token without q": "quotable | empty"},
			Ext:    map[string]string{"lower token without q": "quotable | empty"},
		}},
	)
}

func ExampleMatchAccept() {
	header := http.Header{"Accept": {"text/html, text/*;q=0.1"}}
	accept := Accept(header)
	fmt.Println(MatchAccept(accept, "text/html").Q)
	fmt.Println(MatchAccept(accept, "text/plain").Q)
	fmt.Println(MatchAccept(accept, "image/gif").Q)
	// Output: 1
	// 0.1
	// 0
}

func TestMatchAccept(t *testing.T) {
	tests := []struct {
		accept    []AcceptElem
		mediaType string
		result    AcceptElem
	}{
		{
			nil,
			"text/html",
			AcceptElem{},
		},
		{
			[]AcceptElem{{Type: "text/plain", Q: 1}},
			"text/html",
			AcceptElem{},
		},
		{
			[]AcceptElem{{Type: "text/plain", Q: 1}, {Type: "text/html", Q: 1}},
			"text/html",
			AcceptElem{Type: "text/html", Q: 1},
		},
		{
			[]AcceptElem{{Type: "text/plain", Q: 1}, {Type: "text/html", Q: 1}},
			"Text/HTML",
			AcceptElem{Type: "text/html", Q: 1},
		},
		{
			[]AcceptElem{
				{Type: "*/*", Q: 0.1},
				{Type: "text/*", Q: 0.5},
				{Type: "text/plain", Q: 0.8},
				{
					Type:   "text/plain",
					Q:      1,
					Params: map[string]string{"format": "flowed"},
				},
			},
			"image/gif",
			AcceptElem{Type: "*/*", Q: 0.1},
		},
		{
			[]AcceptElem{
				{Type: "*/*", Q: 0.1},
				{Type: "text/*", Q: 0.5},
				{Type: "text/plain", Q: 0.8},
				{
					Type:   "text/plain",
					Q:      1,
					Params: map[string]string{"format": "flowed"},
				},
			},
			"text/html",
			AcceptElem{Type: "text/*", Q: 0.5},
		},
		{
			[]AcceptElem{
				{Type: "*/*", Q: 0.1},
				{Type: "text/*", Q: 0.5},
				{Type: "text/plain", Q: 0.8},
				{
					Type:   "text/plain",
					Q:      1,
					Params: map[string]string{"format": "flowed"},
				},
			},
			"text/plain",
			AcceptElem{Type: "text/plain", Q: 0.8},
		},
		{
			[]AcceptElem{
				{
					Type:   "text/plain",
					Q:      1,
					Params: map[string]string{"format": "flowed"},
				},
				{Type: "text/plain", Q: 0.8},
			},
			"text/plain",
			AcceptElem{Type: "text/plain", Q: 0.8},
		},
		{
			[]AcceptElem{
				{
					Type:   "text/plain",
					Q:      1,
					Params: map[string]string{"format": "flowed"},
				},
				{Type: "text/*", Q: 0.5},
			},
			"text/plain",
			AcceptElem{Type: "text/*", Q: 0.5},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			actual := MatchAccept(test.accept, test.mediaType)
			if !reflect.DeepEqual(test.result, actual) {
				t.Fatalf("looking up %v in %#v:\nexpected: %#v\nactual:   %#v",
					test.mediaType, test.accept, test.result, actual)
			}
		})
	}
}
