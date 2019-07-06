package httpheader

import (
	"fmt"
	"math/rand"
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
			http.Header{"Warning": {`299 [fe80::a%25en1]:80 "good"`}},
			[]WarningElem{{299, "[fe80::a%25en1]:80", "good", time.Time{}}},
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

func TestWarningRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetWarning, Warning, func(r *rand.Rand) interface{} {
		return mkSlice(r, func(r *rand.Rand) interface{} {
			return WarningElem{
				Code:  100 + r.Intn(900),
				Agent: mkToken(r).(string),
				Text:  mkString(r).(string),
				Date:  mkMaybeDate(r).(time.Time),
			}
		})
	})
}
