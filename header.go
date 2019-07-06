package httpheader

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Allow parses the Allow header from h (RFC 7231 Section 7.4.1).
//
// If there is no such header in h, Allow returns nil.
// If the header is present but empty (meaning all methods are disallowed),
// Allow returns a non-nil slice of length 0.
func Allow(h http.Header) []string {
	var methods []string
	for v, vs := iterElems("", h["Allow"]); vs != nil; v, vs = iterElems(v, vs) {
		var method string
		method, v = consumeItem(v)
		if method != "" {
			methods = append(methods, method)
		}
	}
	if methods == nil && h["Allow"] != nil {
		methods = make([]string, 0)
	}
	return methods
}

// SetAllow replaces the Allow header in h.
// Each of methods must be valid as per RFC 7230 Section 7.1.1.
func SetAllow(h http.Header, methods []string) {
	h.Set("Allow", strings.Join(methods, ", "))
}

// Vary parses the Vary header from h (RFC 7231 Section 7.1.4).
// Parsed names are canonicalized with http.CanonicalHeaderKey.
// A wildcard (Vary: *) is returned as a slice of 1 element.
func Vary(h http.Header) []string {
	var names []string
	for v, vs := iterElems("", h["Vary"]); vs != nil; v, vs = iterElems(v, vs) {
		var name string
		name, v = consumeItem(v)
		if name != "" {
			names = append(names, http.CanonicalHeaderKey(name))
		}
	}
	return names
}

// SetVary replaces the Vary header in h.
// Each of names must be a valid field-name as per RFC 7230 Section 3.2.
// See also AddVary.
func SetVary(h http.Header, names []string) {
	h.Set("Vary", strings.Join(names, ", "))
}

// AddVary is like SetVary but appends instead of replacing.
func AddVary(h http.Header, names []string) {
	h.Add("Vary", strings.Join(names, ", "))
}

// Via parses the Via header from h (RFC 7230 Section 5.7.1).
func Via(h http.Header) []ViaElem {
	var elems []ViaElem
	for v, vs := iterElems("", h["Via"]); vs != nil; v, vs = iterElems(v, vs) {
		var elem ViaElem
		elem.ReceivedProto, v = consumeItem(v)
		if strings.IndexByte(elem.ReceivedProto, '/') == -1 {
			elem.ReceivedProto = "HTTP/" + elem.ReceivedProto
		}
		v = skipWS(v)
		elem.ReceivedBy, v = consumeItem(v)
		v = skipWS(v)
		if peek(v) == '(' {
			elem.Comment, v = consumeComment(v)
		}
		elems = append(elems, elem)
	}
	return elems
}

// SetVia replaces the Via header in h.
// In each of elems, ReceivedProto and ReceivedBy must be valid
// as per RFC 7230 Section 5.7.1; Comment may contain any text,
// which will be escaped automatically.
// See also AddVia.
func SetVia(h http.Header, elems []ViaElem) {
	h.Set("Via", buildVia(elems))
}

// AddVia is like SetVia but appends instead of replacing.
func AddVia(h http.Header, elem ViaElem) {
	h.Add("Via", buildVia([]ViaElem{elem}))
}

func buildVia(elems []ViaElem) string {
	b := &strings.Builder{}
	for i, elem := range elems {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(strings.TrimPrefix(elem.ReceivedProto, "HTTP/"))
		b.WriteString(" ")
		b.WriteString(elem.ReceivedBy)
		if elem.Comment != "" {
			b.WriteString(" ")
			writeComment(b, elem.Comment)
		}
	}
	return b.String()
}

// A ViaElem represents one element of the Via header (RFC 7230 Section 5.7.1).
type ViaElem struct {
	ReceivedProto string // canonicalized to include name: "HTTP/1.1", not "1.1"
	ReceivedBy    string
	Comment       string
}

// Warning parses the Warning header from h (RFC 7234 Section 5.5).
func Warning(h http.Header) []WarningElem {
	var elems []WarningElem
	for v, vs := iterElems("", h["Warning"]); vs != nil; v, vs = iterElems(v, vs) {
		var elem WarningElem
		var codeStr string
		codeStr, v = consumeItem(v)
		elem.Code, _ = strconv.Atoi(codeStr)
		v = skipWS(v)
		elem.Agent, v = consumeItem(v)
		v = skipWS(v)
		elem.Text, v = consumeQuoted(v)
		v = skipWS(v)
		if peek(v) == '"' {
			nextQuote := strings.IndexByte(v[1:], '"')
			if nextQuote > 0 {
				elem.Date, _ = http.ParseTime(v[1 : nextQuote+1])
				v = v[nextQuote+1:]
			}
		}
		elems = append(elems, elem)
	}
	return elems
}

// A WarningElem represents one element of the Warning header
// (RFC 7234 Section 5.5).
type WarningElem struct {
	Code  int
	Agent string    // defaults to "-" on output
	Text  string
	Date  time.Time // zero if missing
}

// SetWarning replaces the Warning header in h.
// In each of elems, Code and Agent must be valid as per RFC 7234 Section 5.5;
// Text may contain any text, which will be escaped automatically.
// See also AddWarning.
func SetWarning(h http.Header, elems []WarningElem) {
	h.Set("Warning", buildWarning(elems))
}

// AddWarning is like SetWarning but appends instead of replacing.
func AddWarning(h http.Header, elem WarningElem) {
	h.Add("Warning", buildWarning([]WarningElem{elem}))
}

func buildWarning(elems []WarningElem) string {
	b := &strings.Builder{}
	for i, elem := range elems {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(strconv.Itoa(elem.Code))
		b.WriteString(" ")
		if elem.Agent == "" {
			elem.Agent = "-"
		}
		b.WriteString(elem.Agent)
		b.WriteString(" ")
		writeQuoted(b, elem.Text)
		if !elem.Date.IsZero() {
			b.WriteString(` "`)
			b.WriteString(elem.Date.Format(http.TimeFormat))
			b.WriteString(`"`)
		}
	}
	return b.String()
}

// Prefer parses the Prefer header from h (RFC 7240 with errata),
// returning a map where keys are preference names, canonicalized to lowercase.
func Prefer(h http.Header) map[string]Pref {
	var r map[string]Pref
	for v, vs := iterElems("", h["Prefer"]); vs != nil; v, vs = iterElems(v, vs) {
		var name string
		var pref Pref
		name, pref.Value, v = consumePrefParam(v)
		for {
			v = skipWS(v)
			if peek(v) != ';' {
				break
			}
			v = skipWS(v[1:])
			switch peek(v) {
			case ';', ',':
				// This is an empty parameter.
			default:
				var paramName, paramValue string
				paramName, paramValue, v = consumePrefParam(v)
				if pref.Params == nil {
					pref.Params = make(map[string]string)
				}
				pref.Params[paramName] = paramValue
			}
		}
		if r == nil {
			r = make(map[string]Pref)
		}
		r[name] = pref
	}
	return r
}

func consumePrefParam(v string) (name, value, newv string) {
	// RFC 7240 errata 4439 'preference-parameter':
	// `name` or `name=value` or `name="quoted value"` (no WS around `=`)
	name, v = consumeItem(v)
	name = strings.ToLower(name)
	if peek(v) == '=' {
		value, v = consumeItemOrQuoted(v[1:])
	}
	return name, value, v
}

// SetPrefer replaces the Prefer header in h (RFC 7240 with errata).
// All keys must be valid tokens (RFC 7230 Section 3.2.6);
// values may contain any text, which will be quoted and escaped as necessary.
// See also AddPrefer.
func SetPrefer(h http.Header, prefs map[string]Pref) {
	h.Set("Prefer", buildPrefer(prefs))
}

// AddPrefer is like SetPrefer but appends instead of replacing.
func AddPrefer(h http.Header, prefs map[string]Pref) {
	h.Add("Prefer", buildPrefer(prefs))
}

func buildPrefer(prefs map[string]Pref) string {
	b := &strings.Builder{}
	first := true
	for name, pref := range prefs {
		if !first {
			b.WriteString(", ")
		}
		first = false
		b.WriteString(name)
		if pref.Value != "" {
			b.WriteString("=")
			writeTokenOrQuoted(b, pref.Value)
		}
		for paramName, paramValue := range pref.Params {
			b.WriteString(paramName)
			if paramValue != "" {
				b.WriteString("=")
				writeTokenOrQuoted(b, paramValue)
			}
		}
	}
	return b.String()
}

// A Pref contains a preference's value and any associated parameters.
type Pref struct {
	Value  string            // must be a valid token
	Params map[string]string // keys are canonicalized to lowercase
}
