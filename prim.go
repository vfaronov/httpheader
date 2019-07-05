package httpheader

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

func token(v string) (tok, rest string) {
	i := 0
	for ; i < len(v); i++ {
		if !isTchar[v[i]] {
			break
		}
	}
	return v[:i], v[i:]
}
