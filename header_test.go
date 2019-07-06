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
			header := http.Header{}
			headerV := reflect.ValueOf(header)
			input := generator(r)
			inputV := reflect.ValueOf(input)
			serializeFuncV.Call([]reflect.Value{headerV, inputV})
			t.Logf("serialized: %#v", header)
			outputV := parseFuncV.Call([]reflect.Value{headerV})[0]
			output := outputV.Interface()
			if !reflect.DeepEqual(input, output) {
				t.Errorf("round-trip failure:\ninput:  %#v\noutput: %#v",
					input, output)
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
		valueV := reflect.ValueOf(value(r))
		sliceV.Index(i).Set(valueV)
	}
	return sliceV.Interface()
}

func mkMap(r *rand.Rand, key, value func(*rand.Rand) interface{}) interface{} {
	nkeys := r.Intn(4)
	keyT := reflect.TypeOf(key(r))
	valueT := reflect.TypeOf(value(r))
	mapT := reflect.MapOf(keyT, valueT)
	if nkeys == 0 {
		return reflect.Zero(mapT).Interface()
	}
	mapV := reflect.MakeMap(mapT)
	for i := 0; i < nkeys; i++ {
		keyV := reflect.ValueOf(key(r))
		valueV := reflect.ValueOf(value(r))
		mapV.SetMapIndex(keyV, valueV)
	}
	return mapV.Interface()
}
