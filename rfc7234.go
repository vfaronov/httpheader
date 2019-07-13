package httpheader

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// A WarningElem represents one element of the Warning header
// (RFC 7234 Section 5.5).
type WarningElem struct {
	Code  int
	Agent string // defaults to "-" on output
	Text  string
	Date  time.Time // zero if missing
}

// Warning parses the Warning header from h (RFC 7234 Section 5.5).
//
// BUG(vfaronov): Incorrectly parses some extravagant values of uri-host
// that do not occur in practice but are theoretically admitted by RFC 3986.
func Warning(h http.Header) []WarningElem {
	var elems []WarningElem
	for v, vs := iterElems("", h["Warning"]); v != ""; v, vs = iterElems(v, vs) {
		var elem WarningElem
		var codeStr string
		codeStr, v = consumeItem(v)
		elem.Code, _ = strconv.Atoi(codeStr)
		v = skipWS(v)
		elem.Agent, v = consumeAgent(v)
		v = skipWS(v)
		elem.Text, v, _ = consumeQuoted(v, false)
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

// SetWarning replaces the Warning header in h. See also AddWarning.
func SetWarning(h http.Header, elems []WarningElem) {
	if len(elems) == 0 {
		h.Del("Warning")
		return
	}
	h.Set("Warning", buildWarning(elems))
}

// AddWarning is like SetWarning but appends instead of replacing.
func AddWarning(h http.Header, elems ...WarningElem) {
	if len(elems) == 0 {
		return
	}
	h.Add("Warning", buildWarning(elems))
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

// CacheDirectives represents directives of the Cache-Control header
// (RFC 7234 Section 5.2). Standard directives are stored in the corresponding
// fields; any unknown extensions are stored in Ext.
type CacheDirectives struct {
	NoStore         bool
	NoTransform     bool
	OnlyIfCached    bool
	MustRevalidate  bool
	Public          bool
	ProxyRevalidate bool
	Immutable       bool // RFC 8246

	// NoCache is true if the no-cache directive is present without an argument.
	// If it has an argument -- a list of header names -- these are
	// stored in NoCacheHeaders, canonicalized with http.CanonicalHeaderKey;
	// while NoCache remains false. Similarly for the private directive.
	NoCache        bool
	NoCacheHeaders []string
	Private        bool
	PrivateHeaders []string

	// For the max-age and s-maxage directives:
	// nil means the directive is absent;
	// a pointer to 0 means the directive is present with the argument 0.
	// Use func Just to set these fields in a CacheDirectives literal.
	MaxAge  *int
	SMaxage *int

	// For the max-stale directive:
	// 0 means the directive is absent or (equivalently) has the argument 0;
	// -1 means the directive is present without a value ("any age").
	MaxStale int

	// For the min-fresh, stale-while-revalidate, and stale-if-error directives:
	// 0 means the directive is absent or (equivalently) has the argument 0.
	MinFresh             int
	StaleWhileRevalidate int // RFC 5861 Section 3
	StaleIfError         int // RFC 5861 Section 4

	// Any unknown extension directives, with keys lowercased.
	// A key mapping to an empty string is serialized to a directive
	// without an argument.
	Ext map[string]string
}

// CacheControl parses the Cache-Control header from h (RFC 7234 Section 5.2).
func CacheControl(h http.Header) CacheDirectives {
	var cc CacheDirectives
	for v, vs := iterElems("", h["Cache-Control"]); v != ""; v, vs = iterElems(v, vs) {
		var name, value string
		name, value, v = consumeParam(v)
		switch name {
		case "private":
			if value == "" {
				cc.Private = true
			} else {
				cc.PrivateHeaders = headerNames(value)
			}
		case "public":
			cc.Public = true
		case "max-age":
			if seconds, err := strconv.Atoi(value); err == nil {
				cc.MaxAge = &seconds
			}
		case "s-maxage":
			if seconds, err := strconv.Atoi(value); err == nil {
				cc.SMaxage = &seconds
			}
		case "no-cache":
			if value == "" {
				cc.NoCache = true
			} else {
				cc.NoCacheHeaders = headerNames(value)
			}
		case "must-revalidate":
			cc.MustRevalidate = true
		case "no-store":
			cc.NoStore = true
		case "stale-while-revalidate":
			cc.StaleWhileRevalidate, _ = strconv.Atoi(value)
		case "stale-if-error":
			cc.StaleIfError, _ = strconv.Atoi(value)
		case "no-transform":
			cc.NoTransform = true
		case "immutable":
			cc.Immutable = true
		case "only-if-cached":
			cc.OnlyIfCached = true
		case "proxy-revalidate":
			cc.ProxyRevalidate = true
		case "max-stale":
			if value == "" {
				cc.MaxStale = -1
			} else {
				cc.MaxStale, _ = strconv.Atoi(value)
			}
		case "min-fresh":
			cc.MinFresh, _ = strconv.Atoi(value)
		default:
			if cc.Ext == nil {
				cc.Ext = make(map[string]string)
			}
			cc.Ext[name] = value
		}
	}
	return cc
}

// SetCacheControl replaces the Cache-Control header in h.
func SetCacheControl(h http.Header, cc CacheDirectives) {
	b := &strings.Builder{}
	var wrote bool
	if cc.NoStore {
		wrote = writeDirective(b, wrote, "no-store", "")
	}
	if cc.NoTransform {
		wrote = writeDirective(b, wrote, "no-transform", "")
	}
	if cc.OnlyIfCached {
		wrote = writeDirective(b, wrote, "only-if-cached", "")
	}
	if cc.MustRevalidate {
		wrote = writeDirective(b, wrote, "must-revalidate", "")
	}
	if cc.Public {
		wrote = writeDirective(b, wrote, "public", "")
	}
	if cc.ProxyRevalidate {
		wrote = writeDirective(b, wrote, "proxy-revalidate", "")
	}
	if cc.Immutable {
		wrote = writeDirective(b, wrote, "immutable", "")
	}
	if cc.Private || len(cc.PrivateHeaders) > 0 {
		// "A sender SHOULD NOT generate the token form"
		wrote = writeDirective(b, wrote, "private", "")
		if !cc.Private {
			b.WriteString(`="`)
			b.WriteString(strings.Join(cc.PrivateHeaders, ","))
			b.WriteString(`"`)
		}
	}
	if cc.NoCache || len(cc.NoCacheHeaders) > 0 {
		// "A sender SHOULD NOT generate the token form"
		wrote = writeDirective(b, wrote, "no-cache", "")
		if !cc.NoCache {
			b.WriteString(`="`)
			b.WriteString(strings.Join(cc.NoCacheHeaders, ","))
			b.WriteString(`"`)
		}
	}
	if cc.MaxAge != nil {
		wrote = writeDirective(b, wrote, "max-age", strconv.Itoa(*cc.MaxAge))
	}
	if cc.SMaxage != nil {
		wrote = writeDirective(b, wrote, "s-maxage", strconv.Itoa(*cc.SMaxage))
	}
	if cc.MaxStale != 0 {
		var value string
		if cc.MaxStale != -1 {
			value = strconv.Itoa(cc.MaxStale)
		}
		wrote = writeDirective(b, wrote, "max-stale", value)
	}
	if cc.MinFresh != 0 {
		wrote = writeDirective(b, wrote, "min-fresh", strconv.Itoa(cc.MinFresh))
	}
	if cc.StaleWhileRevalidate != 0 {
		wrote = writeDirective(b, wrote, "stale-while-revalidate",
			strconv.Itoa(cc.StaleWhileRevalidate))
	}
	if cc.StaleIfError != 0 {
		wrote = writeDirective(b, wrote, "stale-if-error",
			strconv.Itoa(cc.StaleIfError))
	}
	for name, value := range cc.Ext {
		wrote = writeDirective(b, wrote, name, value)
	}
	if b.Len() == 0 {
		h.Del("Cache-Control")
		return
	}
	h.Set("Cache-Control", b.String())
}

func headerNames(v string) []string {
	names := strings.FieldsFunc(v, func(r rune) bool {
		return r == ' ' || r == '\t' || r == ','
	})
	for i := range names {
		names[i] = http.CanonicalHeaderKey(names[i])
	}
	return names
}

// Just returns a pointer to val.
func Just(val int) *int {
	return &val
}
