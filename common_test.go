package httpheader

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func checkParse(t *testing.T, header http.Header, pairs ...interface{}) {
	t.Helper()
	for i := 0; i < len(pairs); i += 2 {
		expected, actual := pairs[i], pairs[i+1]
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("parsing: %#v\nexpected: %#v\nactual:   %#v",
				header, expected, actual)
		}
	}
}

func checkGenerate(t *testing.T, input interface{}, expected, actual http.Header) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("generating: %#v\nexpected: %#v\nactual:   %#v",
			input, expected, actual)
	}
}

// checkRoundTrip runs a battery of sub-tests for the following property:
// Given any valid, canonicalized input value(s), generateFunc must generate
// a header that, when parsed by parseFunc, gives back the same value(s).
// Input values for each of generateFunc's positional arguments (after http.Header)
// are produced from examples, using func likeExample.
func checkRoundTrip(
	t *testing.T,
	generateFunc, parseFunc interface{},
	examples ...interface{},
) {
	t.Helper()
	generateFuncV := reflect.ValueOf(generateFunc)
	parseFuncV := reflect.ValueOf(parseFunc)
	for i := 0; i < 100; i++ {
		t.Run("", func(t *testing.T) {
			r := rand.New(rand.NewSource(int64(i)))
			header := http.Header{}
			var input []interface{}
			for _, ex := range examples {
				input = append(input, likeExample(r, ex))
			}
			argsV := []reflect.Value{reflect.ValueOf(header)}
			for _, in := range input {
				argsV = append(argsV, reflect.ValueOf(in))
			}
			generateFuncV.Call(argsV)
			t.Logf("generated: %#v", header)
			outputV := parseFuncV.Call(argsV[:1])
			var output []interface{}
			for _, outV := range outputV {
				output = append(output, outV.Interface())
			}
			if !reflect.DeepEqual(input, output) {
				t.Errorf("round-trip failure:\ninput:  %#v\noutput: %#v",
					input, output)
			}
		})
	}
}

// likeExample returns a random value that is recursively structured like ex.
func likeExample(rand *rand.Rand, ex interface{}) interface{} {
	exV := reflect.ValueOf(ex)
	switch exV.Kind() {
	case reflect.Bool:
		return rand.Intn(2) == 0
	case reflect.Int:
		return likeInt(rand, ex.(int))
	case reflect.Float32:
		return randFloat(rand)
	case reflect.String:
		return likeString(rand, ex.(string))
	case reflect.Struct:
		switch exV.Type() {
		case reflect.TypeOf(time.Time{}):
			return randTime(rand, !ex.(time.Time).IsZero())
		case reflect.TypeOf(url.URL{}):
			return randURL(rand)
		default:
			return likeStruct(rand, ex)
		}
	case reflect.Ptr:
		exElem := exV.Elem().Interface()
		newV := reflect.New(exV.Elem().Type())
		newV.Elem().Set(reflect.ValueOf(likeExample(rand, exElem)))
		return newV.Interface()
	case reflect.Slice:
		switch exV.Type() {
		case reflect.TypeOf(net.IP{}):
			return likeIP(rand, ex.(net.IP))
		default:
			return likeSlice(rand, ex)
		}
	case reflect.Map:
		return likeMap(rand, ex)
	default:
		panic(fmt.Sprintf("cannot generate value like %#v", ex))
	}
}

func likeInt(rand *rand.Rand, ex int) int {
	switch ex {
	case 9999:
		return 1000 + rand.Intn(9000)
	case 999:
		return 100 + rand.Intn(900)
	case 99:
		return 10 + rand.Intn(90)
	case 9:
		return rand.Intn(10)
	case 0:
		return 0
	default:
		panic(fmt.Sprintf("cannot generate int like %v", ex))
	}
}

func randFloat(rand *rand.Rand) float32 {
	q := rand.Float64()
	// Truncate to 3 digits after decimal point.
	q, _ = strconv.ParseFloat(strconv.FormatFloat(q, 'f', 3, 64), 64)
	return float32(q)
}

func likeString(rand *rand.Rand, ex string) string {
	// like "X | Y" = like "X" or like "Y" at random
	if exs := strings.Split(ex, " | "); len(exs) > 1 {
		return likeString(rand, exs[rand.Intn(len(exs))])
	}
	if ex == "empty" || ex == "" {
		return ""
	}
	// like "X without bc" = like "X" with letters 'b' and 'c' replaced with 'z'
	var without string
	if exs := strings.Split(ex, " without "); len(exs) == 2 {
		ex, without = exs[0], exs[1]
	}
	// like "lower X" = like "X", lowercased
	var lower bool
	if ex1 := strings.TrimPrefix(ex, "lower "); ex1 != ex {
		ex = ex1
		lower = true
	}
	var s string
	switch ex {
	case "token":
		s = randString(rand, tchar)
	case "token68":
		s = randString(rand, alnum+"-._~+/") + strings.Repeat("=", rand.Intn(3))
	case "token*":
		s = randString(rand, tchar) + "*"
	case "Header-Name":
		s = http.CanonicalHeaderKey(randString(rand, tchar))
	case "token/token":
		s = randString(rand, tchar) + "/" + randString(rand, tchar)
	case "quotable":
		s = randString(rand, quotable)
	case "UTF-8":
		s = randUTF8(rand)
	case "URL":
		u := randURL(rand)
		s = u.String()
	case "_obfID":
		s = "_" + randString(rand, alnum+"._-")
	default:
		panic(fmt.Sprintf("cannot generate string like %q", ex))
	}
	if lower {
		s = strings.ToLower(s)
	}
	bs := []byte(s)
	for i, b := range bs {
		if strings.IndexByte(without, b) != -1 {
			bs[i] = 'z'
		}
	}
	s = string(bs)
	return s
}

const (
	digit   = "0123456789"
	loalpha = "abcdefghijklmnopqrstuvwxyz"
	hialpha = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alpha   = hialpha + loalpha
	alnum   = digit + alpha

	// RFC 7230 Section 3.2.6.
	tchar = "!#$%&'*+-.^_`|~" + alnum
	// Characters that can be represented inside a quoted-string or comment.
	quotable = "\t !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~" + alnum +
		"\x80\x81\x82\x83\x84\x85\x86\x87\x88\x89\x8A\x8B\x8C\x8D\x8E\x8F" +
		"\x90" // ...and so on to 0xFF, but this should be enough
)

func randString(rand *rand.Rand, alphabet string) string {
	b := make([]byte, 1+rand.Intn(10))
	for i := 0; i < len(b); i++ {
		b[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(b)
}

func randUTF8(rand *rand.Rand) string {
	runes := make([]rune, 1+rand.Intn(10))
	for i := range runes {
		runes[i] = rune(rand.Intn(0x10FFFF))
	}
	return string(runes)
}

func randTime(rand *rand.Rand, nonzero bool) time.Time {
	if !nonzero && rand.Intn(2) == 0 {
		return time.Time{}
	}
	return time.Date(
		2000+rand.Intn(30), time.Month(1+rand.Intn(12)), 1+rand.Intn(28),
		rand.Intn(24), rand.Intn(60), rand.Intn(60), 0, time.UTC,
	)
}

func randURL(rand *rand.Rand) url.URL {
	if rand.Intn(5) == 0 {
		return url.URL{
			Scheme: "urn",
			Opaque: randString(rand, alnum+":"),
		}
	}
	return url.URL{
		Scheme:   "http",
		Host:     randString(rand, loalpha+digit+".-"),
		Path:     "/" + randString(rand, alnum+"-_~+,;=:/"),
		RawQuery: randString(rand, alnum+"&="),
		Fragment: randString(rand, alnum),
	}
}

func likeIP(rand *rand.Rand, ex net.IP) net.IP {
	if ex == nil {
		return nil
	}
	var ip net.IP
	if rand.Intn(2) == 0 {
		ip = make(net.IP, 4)
	} else {
		ip = make(net.IP, 16)
	}
	rand.Read(ip)
	ip = net.ParseIP(ip.String()) // canonicalize
	return ip
}

// likeStruct returns a new struct of the same type as ex,
// with each field likeExample of ex's value for that field.
func likeStruct(rand *rand.Rand, ex interface{}) interface{} {
	exV := reflect.ValueOf(ex)
	newV := reflect.New(exV.Type()).Elem()
	for i := 0; i < newV.NumField(); i++ {
		fieldEx := exV.Field(i).Interface()
		fieldNew := likeExample(rand, fieldEx)
		newV.Field(i).Set(reflect.ValueOf(fieldNew))
	}
	return newV.Interface()
}

// likeSlice returns a short slice (nil if empty) of the same type as ex,
// with each element likeExample of a random one of ex's elements.
func likeSlice(rand *rand.Rand, ex interface{}) interface{} {
	exV := reflect.ValueOf(ex)
	n := rand.Intn(4)
	if exV.IsNil() || n == 0 {
		return reflect.Zero(exV.Type()).Interface()
	}
	newV := reflect.MakeSlice(exV.Type(), n, n)
	for i := 0; i < n; i++ {
		elemEx := exV.Index(rand.Intn(exV.Len())).Interface()
		elemNew := likeExample(rand, elemEx)
		newV.Index(i).Set(reflect.ValueOf(elemNew))
	}
	return newV.Interface()
}

// likeMap returns a small map (nil if empty) of the same type as ex,
// with each key/value pair likeExample of a random one of ex's key/value pairs.
func likeMap(rand *rand.Rand, ex interface{}) interface{} {
	exV := reflect.ValueOf(ex)
	n := rand.Intn(4)
	if exV.IsNil() || n == 0 {
		return reflect.Zero(exV.Type()).Interface()
	}
	newV := reflect.MakeMap(exV.Type())
	for i := 0; i < n; i++ {
		keyExV := exV.MapKeys()[rand.Intn(exV.Len())]
		keyEx := keyExV.Interface()
		keyNew := likeExample(rand, keyEx)
		valueEx := exV.MapIndex(keyExV).Interface()
		valueNew := likeExample(rand, valueEx)
		newV.SetMapIndex(reflect.ValueOf(keyNew), reflect.ValueOf(valueNew))
	}
	return newV.Interface()
}
