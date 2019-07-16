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
func Warning(h http.Header) []WarningElem {
	values := h["Warning"]
	if values == nil {
		return nil
	}
	elems := make([]WarningElem, 0, estimateElems(values))
	for v, vs := iterElems("", values); v != ""; v, vs = iterElems(v, vs) {
		var elem WarningElem
		var codeStr string
		codeStr, v = consumeTo(v, ' ', false)
		elem.Code, _ = strconv.Atoi(codeStr)
		elem.Agent, v = consumeTo(v, ' ', false)
		elem.Text, v = consumeQuoted(v)
		v = skipWS(v)
		if peek(v) == '"' {
			var dateStr string
			dateStr, v = consumeTo(v[1:], '"', false)
			elem.Date, _ = http.ParseTime(dateStr)
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
			write(b, ", ")
		}
		write(b, strconv.Itoa(elem.Code), " ")
		if elem.Agent == "" {
			elem.Agent = "-"
		}
		write(b, elem.Agent, " ")
		writeQuoted(b, elem.Text)
		if !elem.Date.IsZero() {
			write(b, ` "`, elem.Date.Format(http.TimeFormat), `"`)
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
	Private        bool
	NoCacheHeaders []string
	PrivateHeaders []string

	MaxAge               Delta
	SMaxage              Delta
	MinFresh             Delta
	StaleWhileRevalidate Delta // RFC 5861 Section 3
	StaleIfError         Delta // RFC 5861 Section 4

	// A max-stale directive without an argument (meaning "any age")
	// is represented as the special very large value Eternity.
	MaxStale Delta

	// Any unknown extension directives.
	// A key mapping to an empty string is serialized to a directive
	// without an argument.
	Ext map[string]string
}

// A Delta represents a numeric cache directive which may be either absent
// or a number of seconds. The zero value of Delta is the absent value,
// not 0 seconds.
type Delta struct {
	seconds int
	ok      bool
}

// Value returns the duration of d if it is present; otherwise 0, false.
func (d Delta) Value() (dur time.Duration, ok bool) {
	return time.Duration(d.seconds) * time.Second, d.ok
}

// DeltaSeconds returns a Delta of the given number of seconds.
func DeltaSeconds(s int) Delta {
	return Delta{s, true}
}

// Eternity represents unlimited age for the max-stale cache directive.
var Eternity = Delta{1<<31 - 1, true}

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
				cc.MaxAge = DeltaSeconds(seconds)
			}
		case "s-maxage":
			if seconds, err := strconv.Atoi(value); err == nil {
				cc.SMaxage = DeltaSeconds(seconds)
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
			if seconds, err := strconv.Atoi(value); err == nil {
				cc.StaleWhileRevalidate = DeltaSeconds(seconds)
			}
		case "stale-if-error":
			if seconds, err := strconv.Atoi(value); err == nil {
				cc.StaleIfError = DeltaSeconds(seconds)
			}
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
				cc.MaxStale = Eternity
			} else if seconds, err := strconv.Atoi(value); err == nil {
				cc.MaxStale = DeltaSeconds(seconds)
			}
		case "min-fresh":
			if seconds, err := strconv.Atoi(value); err == nil {
				cc.MinFresh = DeltaSeconds(seconds)
			}
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
			write(b, `="`, strings.Join(cc.PrivateHeaders, ","), `"`)
		}
	}
	if cc.NoCache || len(cc.NoCacheHeaders) > 0 {
		// "A sender SHOULD NOT generate the token form"
		wrote = writeDirective(b, wrote, "no-cache", "")
		if !cc.NoCache {
			write(b, `="`, strings.Join(cc.NoCacheHeaders, ","), `"`)
		}
	}
	if cc.MaxAge.ok {
		wrote = writeDirective(b, wrote, "max-age",
			strconv.Itoa(cc.MaxAge.seconds))
	}
	if cc.SMaxage.ok {
		wrote = writeDirective(b, wrote, "s-maxage",
			strconv.Itoa(cc.SMaxage.seconds))
	}
	if cc.MaxStale.ok {
		var value string
		if cc.MaxStale != Eternity {
			value = strconv.Itoa(cc.MaxStale.seconds)
		}
		wrote = writeDirective(b, wrote, "max-stale", value)
	}
	if cc.MinFresh.ok {
		wrote = writeDirective(b, wrote, "min-fresh",
			strconv.Itoa(cc.MinFresh.seconds))
	}
	if cc.StaleWhileRevalidate.ok {
		wrote = writeDirective(b, wrote, "stale-while-revalidate",
			strconv.Itoa(cc.StaleWhileRevalidate.seconds))
	}
	if cc.StaleIfError.ok {
		wrote = writeDirective(b, wrote, "stale-if-error",
			strconv.Itoa(cc.StaleIfError.seconds))
	}
	for name, value := range cc.Ext {
		wrote = writeDirective(b, wrote, name, value)
	}
	if !wrote {
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
