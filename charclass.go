package httpheader

import "strings"

// classify scans s to determine how it can be represented on the wire.
// tokenOK means s is a simple RFC 7230 token.
// quotedOK means s consists entirely of characters that can be represented in
// an RFC 7230 quoted-string, with the exception of bytes 0x80..0xFF (obs-text),
// which are poorly supported (at best, they are interpreted as ISO-8859-1, which
// is unlikely to be useful).
// quotedSafe means that, in addition to quotedOK, s doesn't contain delimiters
// that are known to confuse naive parsers (which in practice is most parsers).
func classify(s string) (tokenOK, quotedSafe, quotedOK bool) {
	if s == "" {
		return false, true, true
	}
	tokenOK, quotedSafe, quotedOK = true, true, true
	for i := 0; i < len(s); i++ {
		switch byteClass[s[i]] {
		case cUnsafe:
			return false, false, false
		case cQuotedOK:
			quotedSafe, tokenOK = false, false
		case cQuotedSafe:
			tokenOK = false
		}
	}
	return
}

func isToken(s string) bool {
	for i := 0; i < len(s); i++ {
		if byteClass[s[i]] > cTokenOK {
			return false
		}
	}
	return s != ""
}

type charClass int

const (
	cTokenOK charClass = iota
	cQuotedSafe
	cQuotedOK
	cUnsafe
)

var byteClass [256]charClass

func init() {
	for i := 0; i < 0xFF; i++ {
		b := byte(i)
		switch {
		case b < 0x20 || b > 0x7E:
			byteClass[b] = cUnsafe
		case (b >= '0' && b <= '9') ||
			(b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') ||
			strings.ContainsRune("!#$%&'*+-.^_`|~", rune(b)):
			byteClass[b] = cTokenOK
		case b == ',' || b == ';' || b == '"':
			byteClass[b] = cQuotedOK
		default:
			byteClass[b] = cQuotedSafe
		}
	}
}
