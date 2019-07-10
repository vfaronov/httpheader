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
// BUG(vfaronov): Incorrectly parses some extravagant values of warn-agent
// that do not occur in practice but are theoretically admitted by RFC 3986.
func Warning(h http.Header) []WarningElem {
	var elems []WarningElem
	for v, vs := iterElems("", h["Warning"]); vs != nil; v, vs = iterElems(v, vs) {
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
