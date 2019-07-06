package httpheader

import (
	"strings"
)

var (
	isTchar [256]bool
)

func init() {
	tchars := "!#$%&'*+-.^_`|~" +
		"0123456789" +
		"abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for _, c := range tchars {
		isTchar[c] = true
	}
}

func peek(v string) byte {
	if v == "" {
		return 0
	}
	return v[0]
}

func consume(v string, ch byte) (string, bool) {
	if peek(v) != ch {
		return v, false
	}
	return v[1:], true
}

func toNextElem(v string, vs []string) (string, []string) {
	// Skip over any possible unparsed junk at the front.
	for v != "" && v[0] != ',' {
		v = v[1:]
	}
	for {
		// Skip to the next field value if the current one is over.
		if v == "" {
			if len(vs) == 0 {
				return "", nil
			}
			v, vs = vs[0], vs[1:]
			continue
		}
		// Skip over any empty elements.
		// RFC 7230 Section 7: "Empty elements do not contribute
		// to the count of elements present."
		if v[0] == ' ' || v[0] == '\t' || v[0] == ',' {
			v = v[1:]
			continue
		}
		return v, vs
	}
}

func chomp(v string) (item, rest string) {
	for i := 0; i < len(v); i++ {
		if v[i] == ' ' || v[i] == '\t' || v[i] == ',' {
			return v[:i], skipWS(v[i:])
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

func token(v string) (tok, rest string) {
	for i := 0; i < len(v); i++ {
		if !isTchar[v[i]] {
			return v[:i], v[i:]
		}
	}
	return v, ""
}

func number(v string) (n int, rest string) {
	i := 0
	for ; i < len(v); i++ {
		if v[i] >= '0' && v[i] <= '9' {
			n = (n * 10) + int(v[i] - '0')
		} else {
			break
		}
	}
	rest = v[i:]
	return
}

func quoted(v string) (text, rest string) {
	return delimited(v, '"', '"')
}

func comment(v string) (text, rest string) {
	return delimited(v, '(', ')')
}

func delimited(v string, opener, closer byte) (text, rest string) {
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

func writeDelimited(b *strings.Builder, s string, opener, closer byte) {
	b.WriteByte(opener)
	for i := 0; i < len(s); i++ {
		if s[i] == opener || s[i] == closer {
			b.WriteByte('\\')
		}
		b.WriteByte(s[i])
	}
	b.WriteByte(closer)
}
