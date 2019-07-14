package httpheader

import "strings"

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
// Returns the new values for v and vs, with v == "" meaning end of iteration.
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

// consumeTo returns any text from v up to (and possibly including) the first
// occurrence of delim, and the rest of v. If delim does not occur in v,
// it consumes the entire v.
func consumeTo(v string, delim byte, including bool) (text, newv string) {
	pos := strings.IndexByte(v, delim)
	if pos == -1 {
		return v, ""
	}
	if including {
		return v[:pos+1], v[pos+1:]
	}
	return v[:pos], v[pos+1:]
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
	// Unterminated string.
	return v, ""

buffered:
	// But once we have encountered a quoted pair,
	// we have to unquote into a buffer.
	buf := make([]byte, i)
	copy(buf, v)
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
	// Unterminated string.
	return string(buf), ""
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
		text, newv = consumeQuoted(v)
		return
	}
	return consumeItem(v)
}

func write(b *strings.Builder, ss ...string) {
	for _, s := range ss {
		b.WriteString(s)
	}
}

func writeTokenOrQuoted(b *strings.Builder, s string) {
	if isToken(s) {
		write(b, s)
	} else {
		writeQuoted(b, s)
	}
}

func consumeParams(v string) (params map[string]string, newv string) {
	for {
		var name, value string
		name, value, v = consumeParam(v)
		if name == "" {
			break
		}
		if params == nil {
			params = make(map[string]string)
		}
		params[name] = value
	}
	return params, v
}

func consumeParam(v string) (name, value, newv string) {
	v = skipWS(v)
	for peek(v) == ';' {
		v = skipWS(v[1:])
	}
	name, v = consumeItem(v)
	if name == "" {
		return "", "", v
	}
	name = strings.ToLower(name)
	v = skipWS(v)
	if peek(v) == '=' {
		v = skipWS(v[1:])
		value, v = consumeItemOrQuoted(v)
	}
	return name, value, v
}

func writeDirective(b *strings.Builder, wrote bool, name, value string) bool {
	// The wrote flag controls when to output the first comma.
	// See SetCacheControl for example of its usage.
	if wrote {
		write(b, ", ")
	}
	write(b, name)
	if value != "" {
		write(b, "=")
		writeTokenOrQuoted(b, value)
	}
	return true
}

func writeParam(b *strings.Builder, wrote bool, name, value string) bool {
	// The wrote flag controls when to output the first semicolon.
	// See buildForwarded for example of its usage.
	if wrote {
		write(b, ";")
	}
	write(b, name, "=")
	writeTokenOrQuoted(b, value)
	return true
}

func writeParams(b *strings.Builder, params map[string]string) {
	for name, value := range params {
		writeParam(b, true, name, value)
	}
}

func writeNullableParams(b *strings.Builder, params map[string]string) {
	for name, value := range params {
		write(b, ";", name)
		if value != "" {
			write(b, "=")
			writeTokenOrQuoted(b, value)
		}
	}
}

// insertVariform adds the given 'name=value' pair to params, automatically
// initializing params if nil, and decoding 'name*=ext-value' from RFC 8187,
// and returns the new params.
func insertVariform(params map[string]string, name, value string) map[string]string {
	if params == nil {
		params = make(map[string]string)
	}
	if strings.HasSuffix(name, "*") {
		plainName := name[:len(name)-1]
		if decoded, err := decodeExtValue(value); err == nil {
			params[plainName] = decoded
		}
	} else if params[name] == "" { // not filled in from 'name*' yet
		params[name] = value
	}
	return params
}

// writeVariform encodes the parameter with the given name and value into
// one or two of the forms name=token, name="quoted-string" and/or name*=ext-value,
// depending on value, and writes them to b.
func writeVariform(b *strings.Builder, name, value string) {
	tokenOK, quotedSafe, quotedOK := classify(value)
	write(b, "; ", name)
	switch {
	// Token is simplest and safest. Use it if we can. But RFC 8288 Section 3 says:
	// "Previous definitions of the Link header did not equate the token and
	// quoted-string forms explicitly; the title parameter was always quoted, and
	// the hreflang parameter was always a token. Senders wishing to maximize
	// interoperability will send them in those forms."
	case tokenOK && name == "title":
		write(b, `="`, value, `"`)

	case tokenOK:
		write(b, "=", value)

	// Many applications do not process quoted-strings correctly: they are
	// confused by any commas, semicolons, and/or (escaped) double quotes inside.
	// Here are just two random examples of such naive parsers for the Link header:
	//
	//   https://github.com/tomnomnom/linkheader/tree/02ca5825
	//   https://github.com/kennethreitz/requests/blob/4983a9bd/requests/utils.py
	//
	// When the value is without such problematic characters, we can send it
	// as a quoted-string and avoid the ext-value.
	case quotedSafe:
		write(b, "=")
		writeQuoted(b, value)

	// When the value fits into a quoted-string but does contain semicolons
	// and such, we send both the quoted-string form and the ext-value form,
	// so that even if a parser barfs on the semicolons, it might still get
	// the ext-value right (which should look like a normal token to it)
	// and report it to the application for possible use. By this logic,
	// the ext-value should come first. However, RFC 6266 Appendix D recommends
	// the reverse order for Content-Disposition filename in particular.
	case quotedOK && name == "filename":
		write(b, "=")
		writeQuoted(b, value)
		write(b, "; filename*=")
		writeExtValue(b, value)

	case quotedOK:
		write(b, "*=")
		writeExtValue(b, value)
		write(b, "; ", name, "=")
		writeQuoted(b, value)

	// Finally, if the value contains bytes that are not OK for quoted-string
	// at all (including obs-text, which in general is likely to be interpreted
	// as ISO-8859-1), we send only the ext-value.
	default:
		write(b, "*=")
		writeExtValue(b, value)
	}
}
