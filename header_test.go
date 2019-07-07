package httpheader

import (
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

func checkParse(t *testing.T, header http.Header, expected, actual interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("parsing: %#v\nexpected: %#v\nactual:   %#v",
			header, expected, actual)
	}
}

func checkSerialize(t *testing.T, input interface{}, expected, actual http.Header) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("serializing: %#v\nexpected: %#v\nactual:   %#v",
			input, expected, actual)
	}
}

func checkFuzz(t *testing.T, name string, parseFunc, serializeFunc interface{}) {
	// Simplistic fuzz testing: On any input, the parse function must not panic,
	// and the serialize function must not panic on the result of the parse.
	t.Helper()
	parseFuncV := reflect.ValueOf(parseFunc)
	serializeFuncV := reflect.ValueOf(serializeFunc)
	for i := 0; i < 100; i++ {
		t.Run("", func(t *testing.T) {
			r := rand.New(rand.NewSource(int64(i)))
			header := http.Header{}
			for i := 0; i < 1+r.Intn(3); i++ {
				b := make([]byte, r.Intn(64))
				for j := range b {
					// Biased towards punctuation, to trigger more parser states.
					const chars = "\x00 \t,;=-()'*/\"\\abcdefghijklmnopqrstuvwxyz"
					b[j] = chars[r.Intn(len(chars))]
				}
				header.Add(name, string(b))
			}
			t.Logf("header: %#v", header)
			headerV := reflect.ValueOf(header)
			resultV := parseFuncV.Call([]reflect.Value{headerV})[0]
			t.Logf("parsed: %#v", resultV)
			serializeFuncV.Call([]reflect.Value{headerV, resultV})
		})
	}
}

func checkRoundTrip(
	t *testing.T,
	serializeFunc, parseFunc interface{},
	generator func(*rand.Rand) interface{},
) {
	// Property-based test: Serializing and then parsing a valid value
	// should give the same value (modulo canonicalization).
	// Generator is a function (composed of the various mk* functions below),
	// to generate a random value suitable for serializeFunc.
	t.Helper()
	serializeFuncV := reflect.ValueOf(serializeFunc)
	parseFuncV := reflect.ValueOf(parseFunc)
	for i := 0; i < 100; i++ {
		t.Run("", func(t *testing.T) {
			r := rand.New(rand.NewSource(int64(i)))
			headerV := reflect.ValueOf(http.Header{})
			inputV := reflect.ValueOf(generator(r))
			serializeFuncV.Call([]reflect.Value{headerV, inputV})
			t.Logf("serialized: %#v", headerV)
			outputV := parseFuncV.Call([]reflect.Value{headerV})[0]
			if !reflect.DeepEqual(inputV.Interface(), outputV.Interface()) {
				t.Errorf("round-trip failure:\ninput:  %#v\noutput: %#v",
					inputV, outputV)
			}
		})
	}
}

func mkString(r *rand.Rand) interface{} {
	// Any characters allowed inside a quoted-string or a comment.
	b := make([]byte, r.Intn(5))
	r.Read(b)
	for i := range b {
		if b[i] <= 0x20 || b[i] == 0x7E {
			b[i] = '.'
		}
	}
	return string(b)
}

func mkToken(r *rand.Rand) interface{} {
	const (
		punctuation = "-!#$%&'*+.^_`|~"
		digits      = "0123456789"
		letters     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		chars       = punctuation + digits + letters + letters + letters
	)
	b := make([]byte, 1+r.Intn(5))
	for i := range b {
		b[i] = chars[r.Intn(len(chars))]
	}
	return string(b)
}

func mkMaybeToken(r *rand.Rand) interface{} {
	if r.Intn(2) == 0 {
		return ""
	}
	return mkToken(r)
}

func mkLowerToken(r *rand.Rand) interface{} {
	token := mkToken(r).(string)
	return strings.ToLower(token)
}

func mkHeaderName(r *rand.Rand) interface{} {
	token := mkToken(r).(string)
	return http.CanonicalHeaderKey(token)
}

func mkDate(r *rand.Rand) interface{} {
	return time.Date(2000+r.Intn(30), time.Month(1+r.Intn(12)), 1+r.Intn(28),
		r.Intn(24), r.Intn(60), r.Intn(60), 0, time.UTC)
}

func mkMaybeDate(r *rand.Rand) interface{} {
	if r.Intn(2) == 0 {
		return time.Time{}
	}
	return mkDate(r)
}

func mkSlice(r *rand.Rand, value func(*rand.Rand) interface{}) interface{} {
	nitems := r.Intn(4)
	sliceT := reflect.SliceOf(reflect.TypeOf(value(r)))
	if nitems == 0 {
		return reflect.Zero(sliceT).Interface()
	}
	sliceV := reflect.MakeSlice(sliceT, nitems, nitems)
	for i := 0; i < nitems; i++ {
		sliceV.Index(i).Set(reflect.ValueOf(value(r)))
	}
	return sliceV.Interface()
}

func mkMap(r *rand.Rand, key, value func(*rand.Rand) interface{}) interface{} {
	nkeys := r.Intn(4)
	mapT := reflect.MapOf(reflect.TypeOf(key(r)), reflect.TypeOf(value(r)))
	if nkeys == 0 {
		return reflect.Zero(mapT).Interface()
	}
	mapV := reflect.MakeMap(mapT)
	for i := 0; i < nkeys; i++ {
		mapV.SetMapIndex(reflect.ValueOf(key(r)), reflect.ValueOf(value(r)))
	}
	return mapV.Interface()
}
