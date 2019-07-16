package httpheader

import (
	"net/http"
	"strings"
)

// A Pref contains a preference's value and any associated parameters (RFC 7240).
// The Value is lowercased for preferences known to be case-insensitive,
// including 'return' and 'handling'. All keys in Params are also lowercased.
type Pref struct {
	Value  string
	Params map[string]string
}

// Prefer parses the Prefer header from h (RFC 7240 with errata),
// returning a map where keys are lowercase preference names.
func Prefer(h http.Header) map[string]Pref {
	var r map[string]Pref
	for v, vs := iterElems("", h["Prefer"]); v != ""; v, vs = iterElems(v, vs) {
		var name string
		var pref Pref
		name, pref.Value, v = consumeParam(v)
		pref.Params, v = consumeParams(v)
		// RFC 7240 page 5: ``If any preference is specified more than once,
		// only the first instance is to be considered.''
		if _, seen := r[name]; seen {
			continue
		}
		if name == "return" || name == "handling" {
			pref.Value = strings.ToLower(pref.Value)
		}
		if r == nil {
			r = make(map[string]Pref)
		}
		r[name] = pref
	}
	return r
}

// SetPrefer replaces the Prefer header in h (RFC 7240 with errata).
// See also AddPrefer.
func SetPrefer(h http.Header, prefs map[string]Pref) {
	if len(prefs) == 0 {
		h.Del("Prefer")
		return
	}
	h.Set("Prefer", buildPrefer(prefs))
}

// AddPrefer is like SetPrefer but appends instead of replacing.
func AddPrefer(h http.Header, prefs map[string]Pref) {
	if len(prefs) == 0 {
		return
	}
	h.Add("Prefer", buildPrefer(prefs))
}

func buildPrefer(prefs map[string]Pref) string {
	b := &strings.Builder{}
	var wrote bool
	for name, pref := range prefs {
		wrote = writeDirective(b, wrote, name, pref.Value)
		writeNullableParams(b, pref.Params)
	}
	return b.String()
}
