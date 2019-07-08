package httpheader

import (
	"net/http"
	"strings"
)

// A Pref contains a preference's value and any associated parameters (RFC 7240).
type Pref struct {
	Value  string            // must be a valid token
	Params map[string]string // keys are canonicalized to lowercase
}

// Prefer parses the Prefer header from h (RFC 7240 with errata),
// returning a map where keys are preference names, canonicalized to lowercase.
func Prefer(h http.Header) map[string]Pref {
	var r map[string]Pref
	for v, vs := iterElems("", h["Prefer"]); vs != nil; v, vs = iterElems(v, vs) {
		var name string
		var pref Pref
		name, pref.Value, v = consumeParam(v, true)
		pref.Params, v = consumeParams(v, true)
		if r == nil {
			r = make(map[string]Pref)
		}
		r[name] = pref
	}
	return r
}

// SetPrefer replaces the Prefer header in h (RFC 7240 with errata).
// All keys must be valid tokens (RFC 7230 Section 3.2.6);
// values may contain any bytes except control characters.
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
		writeNullableParams(b, pref.Params)
	}
	return b.String()
}
