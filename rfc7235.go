package httpheader

import (
	"net/http"
	"strings"
)

// Auth represents an authentication challenge or credentials (RFC 7235
// Section 2.1). When using the token68 form, the Token field is non-zero.
// When using the auth-param form, the Realm and/or Params fields are non-zero.
// Realm is the value of the 'realm' parameter, if any. Sending an empty realm=""
// is not supported, and any 'realm' key in Params is ignored.
//
// "Star" parameters like RFC 7616's 'username*' are not treated specially.
// Call DecodeExtValue and EncodeExtValue manually if needed.
//
// Scheme names are case-insensitive according to RFC 7235, but many
// implementations erroneously expect them to be in their canonical spelling
// as given in https://www.iana.org/assignments/http-authschemes/.
// Because of this, all functions returning Auth lowercase the Scheme,
// but all functions serializing Auth transform a lowercase Scheme into
// its canonical spelling, or to strings.Title for unknown schemes.
// If you supply a non-lowercase Scheme, its spelling will be preserved.
type Auth struct {
	Scheme string
	Token  string
	Realm  string
	Params map[string]string
}

// WWWAuthenticate parses the WWW-Authenticate header from h
// (RFC 7235 Section 4.1).
func WWWAuthenticate(h http.Header) []Auth {
	return parseChallenges(h["Www-Authenticate"])
}

// SetWWWAuthenticate replaces the WWW-Authenticate header in h.
func SetWWWAuthenticate(h http.Header, challenges []Auth) {
	setChallenges(h, "Www-Authenticate", challenges)
}

// ProxyAuthenticate parses the Proxy-Authenticate header from h
// (RFC 7235 Section 4.3).
func ProxyAuthenticate(h http.Header) []Auth {
	return parseChallenges(h["Proxy-Authenticate"])
}

// SetProxyAuthenticate replaces the Proxy-Authenticate header in h.
func SetProxyAuthenticate(h http.Header, challenges []Auth) {
	setChallenges(h, "Proxy-Authenticate", challenges)
}

// Authorization parses the Authorization header from h (RFC 7235 Section 4.2).
// If h doesn't contain Authorization, a zero Auth is returned.
func Authorization(h http.Header) Auth {
	return parseCredentials(h.Get("Authorization"))
}

// SetAuthorization replaces the Authorization header in h.
func SetAuthorization(h http.Header, credentials Auth) {
	h.Set("Authorization", buildAuth(false, credentials))
}

// ProxyAuthorization parses the Proxy-Authorization header from h
// (RFC 7235 Section 4.4).
// If h doesn't contain Proxy-Authorization, a zero Auth is returned.
func ProxyAuthorization(h http.Header) Auth {
	return parseCredentials(h.Get("Proxy-Authorization"))
}

// SetProxyAuthorization replaces the Proxy-Authorization header in h.
func SetProxyAuthorization(h http.Header, credentials Auth) {
	h.Set("Proxy-Authorization", buildAuth(false, credentials))
}

func parseChallenges(values []string) []Auth {
	if values == nil {
		return nil
	}
	challenges := make([]Auth, 0, estimateElems(values))
	for v, vs := iterElems("", values); v != ""; v, vs = iterElems(v, vs) {
		var challenge Auth
		challenge, v = consumeAuth(v, true)
		challenges = append(challenges, challenge)
	}
	return challenges
}

func parseCredentials(v string) Auth {
	var credentials Auth
	credentials, _ = consumeAuth(v, false)
	return credentials
}

func consumeAuth(v string, challenge bool) (Auth, string) {
	var auth Auth
	auth.Scheme, v = consumeItem(v)
	auth.Scheme = foldAuthScheme(auth.Scheme)
	maybeToken68 := true
ParamsLoop:
	for {
		v = skipWS(v)
		next := peek(v)
		if next == 0 {
			break
		}
		if next == ',' {
			if challenge {
				// This can be the next auth-param or the next challenge.
				if !detectAuthParam(v[1:]) {
					break ParamsLoop
				}
			}
			v = v[1:]
			maybeToken68 = false
			continue
		}

		if maybeToken68 {
			// This can be the first auth-param or the first and only token68.
			if token68, rest := consumeToken68(v); token68 != "" {
				auth.Token = token68
				v = rest
				break
			}
		}

		// Now this is definitely an auth-param.
		maybeToken68 = false
		var name, value string
		name, value, v = consumeParam(v)
		switch name {
		case "":
			break ParamsLoop
		case "realm":
			auth.Realm = value
		default:
			if auth.Params == nil {
				auth.Params = make(map[string]string)
			}
			auth.Params[name] = value
		}
	}
	return auth, v
}

func detectAuthParam(v string) bool {
	// An auth-param always has an equal sign after the first token,
	// but a challenge never does.
	for peek(v) == ' ' || peek(v) == '\t' || peek(v) == ',' {
		v = v[1:]
	}
	_, v = consumeItem(v)
	return peek(skipWS(v)) == '='
}

func consumeToken68(v string) (token68, newv string) {
	orig := v
	token68, v = consumeItem(v)
	// consumeItem didn't consume the trailing equal signs, if any.
	for peek(v) == '=' {
		token68 = orig[:len(token68)+1]
		v = v[1:]
	}
	v = skipWS(v)
	if v != "" && v[0] != ',' {
		// A token68 may only be followed by the next challenge,
		// so if we see anything else, it's an auth-param, not a token68.
		return "", orig
	}
	return token68, v
}

func setChallenges(h http.Header, name string, challenges []Auth) {
	if len(challenges) == 0 {
		h.Del(name)
		return
	}
	h.Set(name, buildAuth(true, challenges...))
}

func buildAuth(challenge bool, auths ...Auth) string {
	b := &strings.Builder{}
	for i, auth := range auths {
		if i > 0 {
			write(b, ", ")
		}
		write(b, unfoldAuthScheme(auth.Scheme))
		if auth.Token != "" {
			write(b, " ", auth.Token)
			continue
		}
		var wrote bool
		if auth.Realm != "" {
			// RFC 7235 page 6: ``For historical reasons, a sender MUST only
			// generate the quoted-string syntax.''
			write(b, " realm=")
			writeQuoted(b, auth.Realm)
			wrote = true
		}
		for name, value := range auth.Params {
			if strings.ToLower(name) == "realm" {
				continue
			}
			if wrote {
				write(b, ", ")
			} else {
				write(b, " ")
			}
			write(b, name, "=")
			if mustQuoteAuthParam(auth.Scheme, name, challenge) {
				writeQuoted(b, value)
			} else {
				writeTokenOrQuoted(b, value)
			}
			wrote = true
		}
	}
	return b.String()
}

func mustQuoteAuthParam(scheme, param string, challenge bool) bool {
	// RFC 7616 (pp. 9 and 10) requires that certain parameters always be quoted.
	// (It also requires that some parameters never be quoted, but we can't
	// do anything about that if the caller supplies a value that requires
	// quoting.) To make things even worse, the 'qop' parameter gets
	// both of these treatments, depending on whether it's in a challenge
	// or in credentials.
	if !strings.EqualFold(scheme, "Digest") {
		return false
	}
	switch strings.ToLower(param) {
	case "cnonce", "domain", "nonce", "opaque", "realm", "response", "uri", "username":
		return true
	case "qop":
		return challenge
	default:
		return false
	}
}

func foldAuthScheme(scheme string) string {
	// Most of the time, scheme will be in canonical spelling.
	// Look it up in the map first, to avoid allocating the lowercase spelling.
	if lower := knownSchemeFold[scheme]; lower != "" {
		return lower
	}
	return strings.ToLower(scheme)
}

func unfoldAuthScheme(scheme string) string {
	if !isLower(scheme) {
		// Preserve spelling supplied by the caller.
		return scheme
	}
	if canonical := knownSchemeUnfold[scheme]; canonical != "" {
		return canonical
	}
	return strings.Title(scheme)
}

var (
	knownSchemeFold   = make(map[string]string)
	knownSchemeUnfold = make(map[string]string)
)

func init() {
	// https://www.iana.org/assignments/http-authschemes/http-authschemes.xhtml
	registered := []string{
		"Basic",
		"Bearer",
		"Digest",
		"HOBA",
		"Mutual",
		"Negotiate",
		"OAuth",
		"SCRAM-SHA-1",
		"SCRAM-SHA-256",
		"vapid",
	}
	for _, canonical := range registered {
		lower := strings.ToLower(canonical)
		knownSchemeFold[canonical] = lower
		knownSchemeUnfold[lower] = canonical
	}
}
