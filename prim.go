package httpheader

import (
	"regexp"
	"strings"
)

var tokenExp = regexp.MustCompile("^[-!#$%&'*+.^_`|~0-9a-zA-Z]+$")

func peek(v string) byte {
	if v == "" {
		return 0
	}
	return v[0]
}

// iterElems iterates over elements in comma-separated header fields
// (RFC 7230 Section 7) spanning multiple field-values (Section 3.2.2).
// iterElems moves to the beginning of the next non-empty element in v.
// If there are no more such elements in v, takes the next v from vs.
// Returns the new values for v and vs, with vs = nil meaning end of iteration.
func iterElems(v string, vs []string) (newv string, newvs []string) {
	orig := true // true means we are still at the same element we started at
	for {
		for v == "" {
			if len(vs) == 0 {
				return "", nil
			}
			v, vs = vs[0], vs[1:]
			orig = false
		}
		switch v[0] {
		case ',':
			orig = false
		case ' ', '\t':
			// Whitespace between elements.
		default:
			if !orig {
				return v, vs
			}
		}
		v = v[1:]
	}
}

// consumeItem returns the item from the beginning of v, and the rest of v.
// An item is a run of text up to whitespace, comma, semicolon, or equal sign.
// Callers should check that the item is non-empty if they need to make progress.
func consumeItem(v string) (item, newv string) {
	for i := 0; i < len(v); i++ {
		switch v[i] {
		case ' ', '\t', ',', ';', '=':
			return v[:i], v[i:]
		}
	}
	return v, ""
}

func skipWS(v string) string {
	for v != "" && (v[0] == ' ' || v[0] == '\t') {
		v = v[1:]
	}
	return v
}

func consumeQuoted(v string) (text, newv string) {
	return consumeDelimited(v, '"', '"')
}

func consumeComment(v string) (text, newv string) {
	return consumeDelimited(v, '(', ')')
}

func consumeDelimited(v string, opener, closer byte) (text, newv string) {
	if peek(v) != opener {
		return "", v
	}
	v = v[1:]

	// In the common case, when there are no quoted pairs,
	// we can simply slice the string between the outermost delimiters.
	nesting := 1
	i := 0
	for ; i < len(v); i++ {
		switch v[i] {
		case closer:
			nesting--
			if nesting == 0 {
				return v[:i], v[i+1:]
			}
		case opener:
			nesting++
		case '\\': // start of a quoted pair
			goto buffered
		}
	}
	return v, "" // unterminated string

buffered:
	// But once we have encountered a quoted pair,
	// we have to unquote into a buffer.
	buf := make([]byte, i, len(v))
	copy(buf, v[:i])
	quoted := false
	for ; i < len(v); i++ {
		switch {
		case quoted:
			buf = append(buf, v[i])
			quoted = false
		case v[i] == closer:
			nesting--
			if nesting == 0 {
				return string(buf), v[i+1:]
			}
			buf = append(buf, v[i])
		case v[i] == opener:
			nesting++
			buf = append(buf, v[i])
		case v[i] == '\\':
			quoted = true
		default:
			buf = append(buf, v[i])
		}
	}
	return string(buf), "" // unterminated string
}

func writeQuoted(b *strings.Builder, s string) {
	writeDelimited(b, s, '"', '"')
}

func writeComment(b *strings.Builder, s string) {
	writeDelimited(b, s, '(', ')')
}

func writeDelimited(b *strings.Builder, s string, opener, closer byte) {
	b.WriteByte(opener)
	for i := 0; i < len(s); i++ {
		if s[i] == opener || s[i] == closer || s[i] == '\\' {
			b.WriteByte('\\')
		}
		b.WriteByte(s[i])
	}
	b.WriteByte(closer)
}

func consumeItemOrQuoted(v string) (text, newv string) {
	if peek(v) == '"' {
		return consumeQuoted(v)
	}
	return consumeItem(v)
}

func writeTokenOrQuoted(b *strings.Builder, s string) {
	if tokenExp.MatchString(s) {
		b.WriteString(s)
	} else {
		writeQuoted(b, s)
	}
}

func consumeParameterized(v string) (
	item string,
	params map[string]string,
	newv string,
) {
	item, v = consumeItem(v)
	item = strings.ToLower(item)
	params, v = consumeParams(v)
	return item, params, v
}

func consumeParams(v string) (params map[string]string, newv string) {
	for {
		v = skipWS(v)
		if peek(v) != ';' {
			break
		}
		v = skipWS(v[1:])
		if c := peek(v); c == ';' || c == ',' || c == 0 {
			// This is an empty parameter.
			continue
		}
		var name, value string
		name, value, v = consumeParam(v)
		if params == nil {
			params = make(map[string]string)
		}
		params[name] = value
	}
	return params, v
}

func consumeParam(v string) (name, value, newv string) {
	name, v = consumeItem(v)
	name = strings.ToLower(name)
	v = skipWS(v)
	if peek(v) == '=' {
		v = skipWS(v[1:])
		value, v = consumeItemOrQuoted(v)
	}
	return name, value, v
}

func writeParameterized(b *strings.Builder, item string, params map[string]string) {
	b.WriteString(item)
	for name, value := range params {
		b.WriteString(";")
		b.WriteString(name)
		b.WriteString("=")
		writeTokenOrQuoted(b, value)
	}
}

func writeNullableParams(b *strings.Builder, params map[string]string) {
	for name, value := range params {
		b.WriteString(";")
		b.WriteString(name)
		if value != "" {
			b.WriteString("=")
			writeTokenOrQuoted(b, value)
		}
	}
}

func consumeAgent(v string) (agent, newv string) {
	// RFC 7230 received-by and RFC 7234 warn-agent. This is tricky because
	// it can contain commas, semicolons, equal signs (see test cases)
	// or even be empty, if you read the grammar literally (reg-name may be empty).
	// The reg-name cases are too much for me right now, but it's easy to handle
	// the IP-Literal case: it's delimited by brackets and never contains brackets.
	if peek(v) == '[' {
		if end := strings.IndexByte(v, ']'); end >= 0 {
			var maybePort string
			maybePort, newv = consumeItem(v[end+1:])
			agent = v[:end+1+len(maybePort)]
			return
		}
	}
	return consumeItem(v)
}
