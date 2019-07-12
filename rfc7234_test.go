package httpheader

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func ExampleWarning() {
	header := http.Header{"Warning": {`299 gw1 "something is wrong"`}}
	fmt.Printf("%+v", Warning(header))
	// Output: [{Code:299 Agent:gw1 Text:something is wrong Date:0001-01-01 00:00:00 +0000 UTC}]
}

func ExampleAddWarning() {
	header := http.Header{}
	AddWarning(header, WarningElem{
		Code: 299,
		Text: "something is fishy",
	})
}

func TestWarning(t *testing.T) {
	tests := []struct {
		header http.Header
		result []WarningElem
	}{
		// Valid headers.
		{
			http.Header{"Warning": {`299 - "good"`}},
			[]WarningElem{{299, "-", "good", time.Time{}}},
		},
		{
			http.Header{"Warning": {`299 example.net:80 "good"`}},
			[]WarningElem{{299, "example.net:80", "good", time.Time{}}},
		},
		{
			// See RFC 6874.
			http.Header{"Warning": {`299 [fe80::a%25en1] "good"`}},
			[]WarningElem{{299, "[fe80::a%25en1]", "good", time.Time{}}},
		},
		{
			http.Header{"Warning": {`199 - "good", 299 - "better"`}},
			[]WarningElem{
				{199, "-", "good", time.Time{}},
				{299, "-", "better", time.Time{}},
			},
		},
		{
			http.Header{"Warning": {`199 - "good" , 299 - "better"`}},
			[]WarningElem{
				{199, "-", "good", time.Time{}},
				{299, "-", "better", time.Time{}},
			},
		},
		{
			http.Header{"Warning": {
				`299 - "good" "Sat, 06 Jul 2019 05:45:48 GMT"`,
			}},
			[]WarningElem{{
				299, "-", "good",
				time.Date(2019, time.July, 6, 5, 45, 48, 0, time.UTC),
			}},
		},
		{
			http.Header{"Warning": {
				`199 - "good" "Sat, 06 Jul 2019 05:45:48 GMT",299 - "better"`,
			}},
			[]WarningElem{
				{
					199, "-", "good",
					time.Date(2019, time.July, 6, 5, 45, 48, 0, time.UTC),
				},
				{
					299, "-", "better",
					time.Time{},
				},
			},
		},
		{
			http.Header{"Warning": {
				`199 - "good" "Sat, 06 Jul 2019 05:45:48 GMT"\t,299 - "better"`,
			}},
			[]WarningElem{
				{
					199, "-", "good",
					time.Date(2019, time.July, 6, 5, 45, 48, 0, time.UTC),
				},
				{
					299, "-", "better",
					time.Time{},
				},
			},
		},
		{
			http.Header{"Warning": {`299 - "with \"escaped\" quotes"`}},
			[]WarningElem{{299, "-", `with "escaped" quotes`, time.Time{}}},
		},
		{
			http.Header{"Warning": {`299 - "\"escaped\" quotes"`}},
			[]WarningElem{{299, "-", `"escaped" quotes`, time.Time{}}},
		},
		{
			http.Header{"Warning": {`299 - "with \"escaped\""`}},
			[]WarningElem{{299, "-", `with "escaped"`, time.Time{}}},
		},
		{
			// This is a valid warn-agent, per uri-host -> IPvFuture.
			http.Header{"Warning": {
				`214 [v9.a51c00de,route=51]:8080 "converted from 5D media!"`,
			}},
			[]WarningElem{
				{
					214,
					"[v9.a51c00de,route=51]:8080",
					"converted from 5D media!",
					time.Time{},
				},
			},
		},
		{
			// This is a valid warn-agent, per uri-host -> reg-name -> sub-delims,
			// but we currently don't parse it. This is a documented bug.
			http.Header{"Warning": {`214 funky,reg-name "WAT"`}},
			[]WarningElem{
				{214, "funky", "", time.Time{}},
				{0, `"WAT"`, "", time.Time{}},
			},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Warning": {"299"}},
			[]WarningElem{{299, "", "", time.Time{}}},
		},
		{
			http.Header{"Warning": {"299 -"}},
			[]WarningElem{{299, "-", "", time.Time{}}},
		},
		{
			http.Header{"Warning": {"299 - unquoted"}},
			[]WarningElem{{299, "-", "", time.Time{}}},
		},
		{
			http.Header{"Warning": {`299  - "two spaces"`}},
			[]WarningElem{{299, "-", "two spaces", time.Time{}}},
		},
		{
			http.Header{"Warning": {`?????,299 - "good"`}},
			[]WarningElem{
				{0, "", "", time.Time{}},
				{299, "-", "good", time.Time{}},
			},
		},
		{
			http.Header{"Warning": {`299  bad, 299 - "good"`}},
			[]WarningElem{
				{299, "bad", "", time.Time{}},
				{299, "-", "good", time.Time{}},
			},
		},
		{
			http.Header{"Warning": {`299 - "good" "bad date"`}},
			[]WarningElem{{299, "-", "good", time.Time{}}},
		},
		{
			http.Header{"Warning": {`299 - "unterminated`}},
			[]WarningElem{{299, "-", "unterminated", time.Time{}}},
		},
		{
			http.Header{"Warning": {`299 - "unterminated\"`}},
			[]WarningElem{{299, "-", `unterminated"`, time.Time{}}},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, Warning(test.header))
		})
	}
}

func TestWarningFuzz(t *testing.T) {
	checkFuzz(t, "Warning", Warning, SetWarning)
}

func TestWarningRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetWarning, Warning,
		[]WarningElem{{
			Code:  999,
			Agent: "token",
			Text:  "quotable | empty",
			Date:  time.Time{},
		}},
	)
}

func BenchmarkWarning(b *testing.B) {
	header := http.Header{"Warning": {
		`299 api.example.com "This API is deprecated; see docs"`,
		`214 proxy1.example.net "Tracking pixels removed"`,
		`112 cache3.example.net "No route to host" "Fri, 12 Jul 2019 17:51:24 GMT", 110 cache3.example.net "Response is stale" "Fri, 12 Jul 2019 17:51:24 GMT"`,
	}}
	for i := 0; i < b.N; i++ {
		Warning(header)
	}
}

func ExampleSetCacheControl() {
	header := http.Header{}
	SetCacheControl(header, CacheDirectives{
		Public: true,
		MaxAge: Just(600),
		Ext:    map[string]string{"priority": "5"},
	})
	fmt.Println(header)
	// Output: map[Cache-Control:[public, max-age=600, priority=5]]
}

func TestCacheControl(t *testing.T) {
	tests := []struct {
		header http.Header
		result CacheDirectives
	}{
		// Valid headers.
		{
			http.Header{"Cache-Control": {", no-store, no-transform"}},
			CacheDirectives{NoStore: true, NoTransform: true},
		},
		{
			http.Header{"Cache-Control": {"Only-If-Cached"}},
			CacheDirectives{OnlyIfCached: true},
		},
		{
			http.Header{"Cache-Control": {
				"foo=bar,,public,",
				"must-revalidate,proxy-revalidate",
			}},
			CacheDirectives{
				Public:          true,
				MustRevalidate:  true,
				ProxyRevalidate: true,
				Ext:             map[string]string{"foo": "bar"},
			},
		},
		{
			http.Header{"Cache-Control": {"Immutable, Max-Age=3600"}},
			CacheDirectives{Immutable: true, MaxAge: Just(3600)},
		},
		{
			http.Header{"Cache-Control": {"private,no-cache"}},
			CacheDirectives{Private: true, NoCache: true},
		},
		{
			http.Header{"Cache-Control": {`private="",no-cache=""`}},
			CacheDirectives{Private: true, NoCache: true},
		},
		{
			http.Header{"Cache-Control": {
				`private=set-cookie`,
				`no-cache="authorization-info, warning, "`,
			}},
			CacheDirectives{
				PrivateHeaders: []string{"Set-Cookie"},
				NoCacheHeaders: []string{"Authorization-Info", "Warning"},
			},
		},
		{
			http.Header{"Cache-Control": {"max-age=0, s-maxage=0"}},
			CacheDirectives{MaxAge: Just(0), SMaxage: Just(0)},
		},
		{
			http.Header{"Cache-Control": {"only-if-cached,max-stale"}},
			CacheDirectives{OnlyIfCached: true, MaxStale: -1},
		},
		{
			http.Header{"Cache-Control": {`only-if-cached,max-stale="3600"`}},
			CacheDirectives{OnlyIfCached: true, MaxStale: 3600},
		},
		{
			http.Header{"Cache-Control": {
				`Min-Fresh=300, Urgent, Foo="bar, baz"`,
			}},
			CacheDirectives{
				MinFresh: 300,
				Ext: map[string]string{
					"urgent": "",
					"foo":    "bar, baz",
				},
			},
		},
		{
			http.Header{"Cache-Control": {
				", public, max-age=86400, stale-while-revalidate=300",
				`, stale-if-error="180"`,
			}},
			CacheDirectives{
				Public:               true,
				MaxAge:               Just(86400),
				StaleWhileRevalidate: 300,
				StaleIfError:         180,
			},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Cache-Control": {`no-store=foo, public="bar"`}},
			CacheDirectives{NoStore: true, Public: true},
		},
		{
			http.Header{"Cache-Control": {"max-age=foo,min-fresh=bar"}},
			CacheDirectives{},
		},
		{
			http.Header{"Cache-Control": {"max-age,min-fresh"}},
			CacheDirectives{},
		},
		{
			http.Header{"Cache-Control": {"max-age=60.098"}},
			CacheDirectives{},
		},
		{
			http.Header{"Cache-Control": {"max-age=60=300, private"}},
			CacheDirectives{MaxAge: Just(60), Private: true},
		},
		{
			http.Header{"Cache-Control": {"stale-if-error = 60"}},
			CacheDirectives{StaleIfError: 60},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, CacheControl(test.header))
		})
	}
}

func TestSetCacheControl(t *testing.T) {
	tests := []struct {
		input  CacheDirectives
		result http.Header
	}{
		{
			CacheDirectives{NoStore: true, NoTransform: true, NoCache: true},
			http.Header{"Cache-Control": {"no-store, no-transform, no-cache"}},
		},
		{
			CacheDirectives{OnlyIfCached: true, MaxStale: 900},
			http.Header{"Cache-Control": {"only-if-cached, max-stale=900"}},
		},
		{
			CacheDirectives{MustRevalidate: true, ProxyRevalidate: true},
			http.Header{"Cache-Control": {"must-revalidate, proxy-revalidate"}},
		},
		{
			CacheDirectives{Public: true, Immutable: true, MaxAge: Just(86400)},
			http.Header{"Cache-Control": {"public, immutable, max-age=86400"}},
		},
		{
			// "A sender SHOULD NOT generate the token form"
			CacheDirectives{NoCacheHeaders: []string{"Set-Cookie"}},
			http.Header{"Cache-Control": {`no-cache="Set-Cookie"`}},
		},
		{
			// "A sender SHOULD NOT generate the token form"
			CacheDirectives{PrivateHeaders: []string{"Set-Cookie"}},
			http.Header{"Cache-Control": {`private="Set-Cookie"`}},
		},
		{
			CacheDirectives{PrivateHeaders: []string{"Set-Cookie", "Request-ID"}},
			http.Header{"Cache-Control": {`private="Set-Cookie,Request-ID"`}},
		},
		{
			CacheDirectives{Private: true, StaleWhileRevalidate: 300},
			http.Header{"Cache-Control": {"private, stale-while-revalidate=300"}},
		},
		{
			CacheDirectives{MaxStale: -1},
			http.Header{"Cache-Control": {"max-stale"}},
		},
		{
			CacheDirectives{MinFresh: 900, Ext: map[string]string{"Qux": ""}},
			http.Header{"Cache-Control": {"min-fresh=900, Qux"}},
		},
		{
			CacheDirectives{
				StaleIfError: 1800,
				Ext:          map[string]string{"bar": "baz, qux"},
			},
			http.Header{"Cache-Control": {`stale-if-error=1800, bar="baz, qux"`}},
		},
		{
			CacheDirectives{
				SMaxage: Just(300),
				Ext:     map[string]string{"priority": "40"},
			},
			http.Header{"Cache-Control": {"s-maxage=300, priority=40"}},
		},
		{
			CacheDirectives{
				NoCache:        true,
				NoCacheHeaders: []string{"Set-Cookie"},
				Private:        true,
				PrivateHeaders: []string{"Authorization-Info"},
			},
			http.Header{"Cache-Control": {"private, no-cache"}},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			header := http.Header{}
			SetCacheControl(header, test.input)
			checkGenerate(t, test.input, test.result, header)
		})
	}
}

func TestCacheControlFuzz(t *testing.T) {
	checkFuzz(t, "Cache-Control", CacheControl, SetCacheControl)
}

func BenchmarkCacheControl(b *testing.B) {
	header := http.Header{"Cache-Control": {
		`private="Set-Cookie", max-age=900, s-maxage=600, stale-if-error=30`,
		`no-transform, must-revalidate`,
	}}
	for i := 0; i < b.N; i++ {
		CacheControl(header)
	}
}
