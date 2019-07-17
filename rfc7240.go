package httpheader

import (
	"net/http"
	"strings"
)

// A Pref contains a preference's value and any associated parameters (RFC 7240).
type Pref struct {
	Value  string
	Params map[string]string
}

// Prefer parses the Prefer header from h (RFC 7240 with errata),
// returning a map where keys are preference names.
func Prefer(h http.Header) map[string]Pref {
	values := h["Prefer"]
	if values == nil {
		return nil
	}
	r := make(map[string]Pref)
	for v, vs := iterElems("", values); v != ""; v, vs = iterElems(v, vs) {
		var name string
		var pref Pref
		name, pref.Value, v = consumeParam(v)
		pref.Params, v = consumeParams(v)
		// RFC 7240 page 5: ``If any preference is specified more than once,
		// only the first instance is to be considered.''
		if _, seen := r[name]; seen {
			continue
		}
		pref.Value = canonicalPref(name, pref.Value)
		r[name] = pref
	}
	return r
}

// SetPrefer replaces the Prefer header in h (RFC 7240 with errata).
func SetPrefer(h http.Header, prefs map[string]Pref) {
	if len(prefs) == 0 {
		h.Del("Prefer")
		return
	}
	b := &strings.Builder{}
	var wrote bool
	for name, pref := range prefs {
		wrote = writeDirective(b, wrote, name, pref.Value)
		writeNullableParams(b, pref.Params)
	}
	h.Set("Prefer", b.String())
}

// PreferenceApplied parses the Preference-Applied header from h (RFC 7240
// with errata), returning a map where keys are preference names.
func PreferenceApplied(h http.Header) map[string]string {
	values := h["Preference-Applied"]
	if values == nil {
		return nil
	}
	r := make(map[string]string)
	for v, vs := iterElems("", values); v != ""; v, vs = iterElems(v, vs) {
		var name, value string
		name, value, v = consumeParam(v)
		if _, seen := r[name]; seen {
			continue
		}
		value = canonicalPref(name, value)
		r[name] = value
	}
	return r
}

// SetPreferenceApplied replaces the Preference-Applied header in h.
func SetPreferenceApplied(h http.Header, prefs map[string]string) {
	if len(prefs) == 0 {
		h.Del("Preference-Applied")
		return
	}
	b := &strings.Builder{}
	var wrote bool
	for name, value := range prefs {
		wrote = writeDirective(b, wrote, name, value)
	}
	h.Set("Preference-Applied", b.String())
}

func canonicalPref(name, value string) string {
	switch name {
	case "handling", "return":
		// These preferences are case-insensitive because RFC 7240 defines them
		// with ABNF's case-insensitive literal strings (RFC 5234 Section 2.3).
		value = strings.ToLower(value)
	}
	return value
}
