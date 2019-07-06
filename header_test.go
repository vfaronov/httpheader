package httpheader

import (
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"testing"
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
	b := make([]byte, 1+r.Intn(5))
	r.Read(b)
	return string(b)
}

func mkToken(r *rand.Rand) interface{} {
	const chars = "-!#$%&'*+.^_`|~0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
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

func mkSlice(r *rand.Rand, value func(*rand.Rand) interface{}) interface{} {
	nitems := 1 + r.Intn(3)
	valueT := reflect.TypeOf(value(r))
	sliceV := reflect.MakeSlice(reflect.SliceOf(valueT), nitems, nitems)
	for i := 0; i < nitems; i++ {
		valueV := reflect.ValueOf(value(r))
		sliceV.Index(i).Set(valueV)
	}
	return sliceV.Interface()
}

func mkMap(r *rand.Rand, key, value func(*rand.Rand) interface{}) interface{} {
	nkeys := 1 + r.Intn(3)
	keyT := reflect.TypeOf(key(r))
	valueT := reflect.TypeOf(value(r))
	mapV := reflect.MakeMap(reflect.MapOf(keyT, valueT))
	for i := 0; i < nkeys; i++ {
		keyV := reflect.ValueOf(key(r))
		valueV := reflect.ValueOf(value(r))
		mapV.SetMapIndex(keyV, valueV)
	}
	return mapV.Interface()
}
