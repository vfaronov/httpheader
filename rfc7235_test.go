package httpheader

import (
	"net/http"
	"os"
	"testing"
)

func TestWWWAuthenticate(t *testing.T) {
	tests := []struct {
		header http.Header
		result []Auth
	}{
		// Valid headers.
		{
			http.Header{"Www-Authenticate": {""}},
			[]Auth{},
		},
		{
			http.Header{"Www-Authenticate": {","}},
			[]Auth{},
		},
		{
			http.Header{"Www-Authenticate": {"\t,\t,"}},
			[]Auth{},
		},
		{
			http.Header{"Www-Authenticate": {"Foo"}},
			[]Auth{{Scheme: "foo"}},
		},
		{
			http.Header{"Www-Authenticate": {"Foo \t"}},
			[]Auth{{Scheme: "foo"}},
		},
		{
			http.Header{"Www-Authenticate": {"Foo,Bar"}},
			[]Auth{{Scheme: "foo"}, {Scheme: "bar"}},
		},
		{
			http.Header{"Www-Authenticate": {"Foo, Bar"}},
			[]Auth{{Scheme: "foo"}, {Scheme: "bar"}},
		},
		{
			http.Header{"Www-Authenticate": {"Foo ,\t, Bar"}},
			[]Auth{{Scheme: "foo"}, {Scheme: "bar"}},
		},
		{
			http.Header{"Www-Authenticate": {`Foo bar="baz", Qux`}},
			[]Auth{
				{Scheme: "foo", Params: map[string]string{"bar": "baz"}},
				{Scheme: "qux"},
			},
		},
		{
			http.Header{"Www-Authenticate": {"Foo bar = baz, Qux"}},
			[]Auth{
				{Scheme: "foo", Params: map[string]string{"bar": "baz"}},
				{Scheme: "qux"},
			},
		},
		{
			http.Header{"Www-Authenticate": {"Foo bar = baz,\t, Qux"}},
			[]Auth{
				{Scheme: "foo", Params: map[string]string{"bar": "baz"}},
				{Scheme: "qux"},
			},
		},
		{
			http.Header{"Www-Authenticate": {"Foo bar = baz, Qux = xyzzy"}},
			[]Auth{
				{
					Scheme: "foo",
					Params: map[string]string{
						"bar": "baz",
						"qux": "xyzzy",
					},
				},
			},
		},
		{
			http.Header{"Www-Authenticate": {"Foo bar = baz,\t, Qux =xyzzy"}},
			[]Auth{
				{
					Scheme: "foo",
					Params: map[string]string{
						"bar": "baz",
						"qux": "xyzzy",
					},
				},
			},
		},
		{
			http.Header{"Www-Authenticate": {"Foo bar = baz, Qux, xyzzy"}},
			[]Auth{
				{Scheme: "foo", Params: map[string]string{"bar": "baz"}},
				{Scheme: "qux"},
				{Scheme: "xyzzy"},
			},
		},
		{
			http.Header{"Www-Authenticate": {"Foo bar, baz"}},
			[]Auth{{Scheme: "foo", Token: "bar"}, {Scheme: "baz"}},
		},
		{
			http.Header{"Www-Authenticate": {"Foo bar , baz"}},
			[]Auth{{Scheme: "foo", Token: "bar"}, {Scheme: "baz"}},
		},
		{
			http.Header{"Www-Authenticate": {"Foo bar=baz, Qux xyzzy"}},
			[]Auth{
				{Scheme: "foo", Params: map[string]string{"bar": "baz"}},
				{Scheme: "qux", Token: "xyzzy"},
			},
		},
		{
			http.Header{"Www-Authenticate": {"Foo bar=baz,\tQux  xyzzy"}},
			[]Auth{
				{Scheme: "foo", Params: map[string]string{"bar": "baz"}},
				{Scheme: "qux", Token: "xyzzy"},
			},
		},
		{
			http.Header{"Www-Authenticate": {
				`Digest realm="http-auth@example.org", qop="auth, auth-int", algorithm=SHA-256, nonce="7ypf/xlj9XXwfDPE", opaque="FQhe/qaU925kfnzj",  Digest realm="http-auth@example.org", qop="auth, auth-int", algorithm=MD5, nonce="7ypf/xlj9XXwfDPE", opaque="FQhe/qaU925kfnzj`,
			}},
			[]Auth{
				{
					Scheme: "digest",
					Realm:  "http-auth@example.org",
					Params: map[string]string{
						"qop":       "auth, auth-int",
						"algorithm": "SHA-256",
						"nonce":     "7ypf/xlj9XXwfDPE",
						"opaque":    "FQhe/qaU925kfnzj",
					},
				},
				{
					Scheme: "digest",
					Realm:  "http-auth@example.org",
					Params: map[string]string{
						"qop":       "auth, auth-int",
						"algorithm": "MD5",
						"nonce":     "7ypf/xlj9XXwfDPE",
						"opaque":    "FQhe/qaU925kfnzj",
					},
				},
			},
		},
		{
			http.Header{"Www-Authenticate": {
				`Bearer realm="Experimental API",scope="urn:example:channel=HBO&urn:example:rating=G,PG-13",error="invalid_token",error_description="Sorry, your token has expired. Please \"renew\" it.",`,
			}},
			[]Auth{
				{
					Scheme: "bearer",
					Realm:  "Experimental API",
					Params: map[string]string{
						"scope":             "urn:example:channel=HBO&urn:example:rating=G,PG-13",
						"error":             "invalid_token",
						"error_description": `Sorry, your token has expired. Please "renew" it.`,
					},
				},
			},
		},
		{
			http.Header{"Www-Authenticate": {"Foo bar="}},
			[]Auth{{Scheme: "foo", Token: "bar="}},
		},
		{
			http.Header{"Www-Authenticate": {"Foo bar=, baz"}},
			[]Auth{{Scheme: "foo", Token: "bar="}, {Scheme: "baz"}},
		},
		{
			http.Header{"Www-Authenticate": {"Foo bar= baz"}},
			[]Auth{{Scheme: "foo", Params: map[string]string{"bar": "baz"}}},
		},
		{
			http.Header{"Www-Authenticate": {"Foo bar===, baz"}},
			[]Auth{{Scheme: "foo", Token: "bar==="}, {Scheme: "baz"}},
		},
		{
			http.Header{"Www-Authenticate": {"Foo bar=, Baz qux="}},
			[]Auth{
				{Scheme: "foo", Token: "bar="},
				{Scheme: "baz", Token: "qux="},
			},
		},
		{
			http.Header{"Www-Authenticate": {`Newauth realm="apps", type=1, title="Login to \"apps\"", Basic realm="simple"`}},
			[]Auth{
				{
					Scheme: "newauth",
					Realm:  "apps",
					Params: map[string]string{
						"type":  "1",
						"title": `Login to "apps"`,
					},
				},
				{
					Scheme: "basic",
					Realm:  "simple",
				},
			},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Www-Authenticate": {"Basic realm=Hello"}},
			[]Auth{{Scheme: "basic", Realm: "Hello"}},
		},
		{
			http.Header{"Www-Authenticate": {`Basic; realm="Hello"`}},
			[]Auth{{Scheme: "basic", Realm: "Hello"}},
		},
		{
			http.Header{"Www-Authenticate": {`Basic; realm="Hello", Bearer`}},
			[]Auth{{Scheme: "basic", Realm: "Hello"}, {Scheme: "bearer"}},
		},
		{
			http.Header{"Www-Authenticate": {`Basic, realm="Hello", Bearer`}},
			[]Auth{{Scheme: "basic", Realm: "Hello"}, {Scheme: "bearer"}},
		},
		{
			http.Header{"Www-Authenticate": {`Basic="Hello"`}},
			[]Auth{{Scheme: "basic"}},
		},
		{
			http.Header{"Www-Authenticate": {`Foo bar, baz=qux`}},
			[]Auth{{Scheme: "foo", Token: "bar"}, {Scheme: "baz"}},
		},
		{
			http.Header{"Www-Authenticate": {`Foo bar=baz qux=xyzzy`}},
			[]Auth{
				{
					Scheme: "foo",
					Params: map[string]string{"bar": "baz", "qux": "xyzzy"},
				},
			},
		},
		{
			http.Header{"Www-Authenticate": {`Foo , =bar`}},
			[]Auth{{Scheme: "foo"}},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, WWWAuthenticate(test.header))
		})
	}
}

func BenchmarkWWWAuthenticateSimple(b *testing.B) {
	header := http.Header{"Www-Authenticate": {`Basic realm="protected space"`}}
	for i := 0; i < b.N; i++ {
		WWWAuthenticate(header)
	}
}

func BenchmarkWWWAuthenticateComplex(b *testing.B) {
	header := http.Header{"Www-Authenticate": {`Basic realm="example.com API", Bearer realm="example.com API", scope="urn:example:channel=HBO&urn:example:rating=G,PG-13", Digest realm="http-auth@example.org", qop="auth, auth-int", algorithm=MD5, nonce="7ypf/xlj9XXwfDPEoM4URrv/xwf94BcCAzFZH4GiTo0v", opaque="FQhe/qaU925kfnzjCev0ciny7QMkPqMAFRtzCUYo5tdS"`}}
	for i := 0; i < b.N; i++ {
		WWWAuthenticate(header)
	}
}

func TestSetWWWAuthenticate(t *testing.T) {
	tests := []struct {
		input  []Auth
		result http.Header
	}{
		{
			[]Auth{{Scheme: "basic"}},
			http.Header{"Www-Authenticate": {`Basic`}},
		},
		{
			[]Auth{{Scheme: "oauth"}, {Scheme: "hoba"}},
			http.Header{"Www-Authenticate": {`OAuth, HOBA`}},
		},
		{
			// RFC 7235 page 6: ``For historical reasons, a sender MUST only
			// generate the quoted-string syntax [for the realm parameter].''
			// RFC 7616 page 9: ``For historical reasons, a sender MUST only
			// generate the quoted string syntax values for the following
			// parameters: realm, domain, nonce, opaque, and qop.''
			[]Auth{
				{
					Scheme: "Digest",
					Realm:  "Hello1",
					Params: map[string]string{"domain": "ururu"},
				},
				{
					Scheme: "Digest",
					Realm:  "Hello2",
					Params: map[string]string{"Nonce": "7YSRY5UV"},
				},
				{
					Scheme: "Digest",
					Realm:  "Hello3",
					Params: map[string]string{"opaque": "JW6jJuz4"},
				},
				{
					Scheme: "Digest",
					Realm:  "Hello4",
					Params: map[string]string{"QOP": "auth"},
				},
			},
			http.Header{"Www-Authenticate": {`Digest realm="Hello1", domain="ururu", Digest realm="Hello2", Nonce="7YSRY5UV", Digest realm="Hello3", opaque="JW6jJuz4", Digest realm="Hello4", QOP="auth"`}},
		},
		{
			[]Auth{{Scheme: "Basic", Realm: `Hello, "lovely" world!`}},
			http.Header{"Www-Authenticate": {`Basic realm="Hello, \"lovely\" world!"`}},
		},
		{
			[]Auth{
				{
					Scheme: "DIGEST",
					Realm:  "TEST",
					Params: map[string]string{
						"QOP":   "AUTH, AUTH-INT",
						"REALM": "IGNORE ME",
					},
				},
			},
			http.Header{"Www-Authenticate": {`DIGEST realm="TEST", QOP="AUTH, AUTH-INT"`}},
		},
		{
			[]Auth{{Scheme: "foobar", Token: "XpL+OI2ydaLvA1/fTmpdwXrb="}},
			http.Header{"Www-Authenticate": {`Foobar XpL+OI2ydaLvA1/fTmpdwXrb=`}},
		},
		{
			[]Auth{
				{
					Scheme: "digest",
					Params: map[string]string{"qop": "auth, auth-int"},
				},
			},
			http.Header{"Www-Authenticate": {`Digest qop="auth, auth-int"`}},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			header := http.Header{}
			SetWWWAuthenticate(header, test.input)
			checkGenerate(t, test.input, test.result, header)
		})
	}
}

func TestWWWAuthenticateFuzz(t *testing.T) {
	checkFuzz(t, "Www-Authenticate", WWWAuthenticate, SetWWWAuthenticate)
}

func TestWWWAuthenticateRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetWWWAuthenticate, WWWAuthenticate,
		[]Auth{
			{
				Scheme: "lower token",
				Token:  "token68",
			},
			{
				Scheme: "lower token",
				Realm:  "quotable",
				Params: map[string]string{
					"lower token": "token | quotable | empty",
				},
			},
		},
	)
}

func ExampleSetWWWAuthenticate() {
	header := http.Header{}
	SetWWWAuthenticate(header, []Auth{{
		Scheme: "Bearer",
		Realm:  "api.example.com",
		Params: map[string]string{"scope": "profile"},
	}})
	header.Write(os.Stdout)
	// Output: Www-Authenticate: Bearer realm="api.example.com", scope=profile
}

func TestAuthorization(t *testing.T) {
	tests := []struct {
		header http.Header
		result Auth
	}{
		// Valid headers.
		{
			http.Header{"Authorization": {""}},
			Auth{},
		},
		{
			http.Header{"Authorization": {"Foo"}},
			Auth{Scheme: "foo"},
		},
		{
			http.Header{"Authorization": {"Foo \t,"}},
			Auth{Scheme: "foo"},
		},
		{
			http.Header{"Authorization": {"Foo ,Bar=baz"}},
			Auth{Scheme: "foo", Params: map[string]string{"bar": "baz"}},
		},
		{
			http.Header{"Authorization": {"Foo   Bar"}},
			Auth{Scheme: "foo", Token: "Bar"},
		},
		{
			http.Header{"Authorization": {`Foo bar="baz, qux", xyzzy = 123`}},
			Auth{
				Scheme: "foo",
				Params: map[string]string{
					"bar":   "baz, qux",
					"xyzzy": "123",
				},
			},
		},
		{
			http.Header{"Authorization": {"Foo bar = baz,\t, Qux =xyzzy"}},
			Auth{
				Scheme: "foo",
				Params: map[string]string{
					"bar": "baz",
					"qux": "xyzzy",
				},
			},
		},
		{
			http.Header{"Authorization": {"Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ=="}},
			Auth{Scheme: "basic", Token: "QWxhZGRpbjpvcGVuIHNlc2FtZQ=="},
		},
		{
			http.Header{"Authorization": {
				`Digest username="Mufasa", realm="http-auth@example.org", uri="/dir/index.html", algorithm=MD5, nonce="7ypf/xlj9XXwfDPE", nc=00000001, cnonce="f2/wE4q74E6zIJEt", qop=auth, response="8ca523f5e9506fed4657c9700eebdbec", opaque="FQhe/qaU925kfnzj"`,
			}},
			Auth{
				Scheme: "digest",
				Realm:  "http-auth@example.org",
				Params: map[string]string{
					"username":  "Mufasa",
					"uri":       "/dir/index.html",
					"algorithm": "MD5",
					"nonce":     "7ypf/xlj9XXwfDPE",
					"nc":        "00000001",
					"cnonce":    "f2/wE4q74E6zIJEt",
					"qop":       "auth",
					"response":  "8ca523f5e9506fed4657c9700eebdbec",
					"opaque":    "FQhe/qaU925kfnzj",
				},
			},
		},

		// Invalid headers.
		// Precise outputs on them are not a guaranteed part of the API.
		// They may change as convenient for the parsing code.
		{
			http.Header{"Authorization": {"Basic j.doe:secret123"}},
			Auth{
				Scheme: "basic",
				Token:  "j.doe:secret123",
			},
		},
		{
			http.Header{"Authorization": {"Foo bar=baz, qux, xyzzy"}},
			Auth{
				Scheme: "foo",
				Params: map[string]string{"bar": "baz", "qux": "", "xyzzy": ""},
			},
		},
		{
			http.Header{"Authorization": {`Foo; password="Hello"`}},
			Auth{
				Scheme: "foo",
				Params: map[string]string{"password": "Hello"},
			},
		},
		{
			http.Header{"Authorization": {"Basic=XpLOI2ydaLvA1z"}},
			Auth{Scheme: "basic"},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			checkParse(t, test.header, test.result, Authorization(test.header))
		})
	}
}

func ExampleSetAuthorization() {
	header := http.Header{}
	SetAuthorization(header, Auth{Scheme: "Bearer", Token: "wvmjsa9N0iyLjnFo"})
	header.Write(os.Stdout)
	// Output: Authorization: Bearer wvmjsa9N0iyLjnFo
}

func TestSetAuthorization(t *testing.T) {
	tests := []struct {
		input  Auth
		result http.Header
	}{
		// RFC 7616 page 10: ``For historical reasons, a sender MUST only generate
		// the quoted string syntax for the following parameters: username, realm,
		// nonce, uri, response, cnonce, and opaque.''
		{
			Auth{
				Scheme: "Digest",
				Realm:  "Hello",
				Params: map[string]string{"username": "john.doe"},
			},
			http.Header{"Authorization": {`Digest realm="Hello", username="john.doe"`}},
		},
		{
			Auth{
				Scheme: "Digest",
				Realm:  "Hello",
				Params: map[string]string{"nonce": "YVtLPys2"},
			},
			http.Header{"Authorization": {`Digest realm="Hello", nonce="YVtLPys2"`}},
		},
		{
			Auth{
				Scheme: "Digest",
				Realm:  "Hello",
				Params: map[string]string{"URI": "foo-bar"},
			},
			http.Header{"Authorization": {`Digest realm="Hello", URI="foo-bar"`}},
		},
		{
			Auth{
				Scheme: "Digest",
				Realm:  "Hello",
				Params: map[string]string{"response": "D5HBzl62aK17KWJd"},
			},
			http.Header{"Authorization": {`Digest realm="Hello", response="D5HBzl62aK17KWJd"`}},
		},
		{
			Auth{
				Scheme: "DIGEST",
				Realm:  "Hello",
				Params: map[string]string{"CNonce": "00000912"},
			},
			http.Header{"Authorization": {`DIGEST realm="Hello", CNonce="00000912"`}},
		},
		{
			Auth{
				Scheme: "Digest",
				Realm:  "Hello",
				Params: map[string]string{"opaque": "FooBar"},
			},
			http.Header{"Authorization": {`Digest realm="Hello", opaque="FooBar"`}},
		},
		// RFC 7616 page 10: ``For historical reasons, a sender MUST NOT generate
		// the quoted string syntax for the following parameters: algorithm, qop,
		// and nc.'' But we only generate the quoted string syntax when the
		// supplied value doesn't fit otherwise, so the only thing we check here
		// is that the special handling of qop for WWW-Authenticate (where it is,
		// conversely, required to be quoted) doesn't affect Authorization.
		{
			Auth{
				Scheme: "Digest",
				Realm:  "Hello",
				Params: map[string]string{"qop": "auth"},
			},
			http.Header{"Authorization": {`Digest realm="Hello", qop=auth`}},
		},
		{
			// No need to do any of the above stuff for non-Digest credentials.
			Auth{
				Scheme: "Foo",
				Params: map[string]string{"nonce": "abcde"},
			},
			http.Header{"Authorization": {"Foo nonce=abcde"}},
		},
		{
			// As documented, we can use EncodeExtValue manually
			// to send 'username*'.
			Auth{
				Scheme: "Digest",
				Params: map[string]string{
					"username*": EncodeExtValue("Васян", "ru"),
				},
			},
			http.Header{"Authorization": {`Digest username*=UTF-8'ru'%D0%92%D0%B0%D1%81%D1%8F%D0%BD`}},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			header := http.Header{}
			SetAuthorization(header, test.input)
			checkGenerate(t, test.input, test.result, header)
		})
	}
}

func TestAuthorizationFuzz(t *testing.T) {
	checkFuzz(t, "Authorization", Authorization, SetAuthorization)
}

func TestAuthorizationRoundTrip(t *testing.T) {
	checkRoundTrip(t, SetAuthorization, Authorization,
		Auth{
			Scheme: "lower token",
			Token:  "token68",
		},
	)
	checkRoundTrip(t, SetAuthorization, Authorization,
		Auth{
			Scheme: "lower token",
			Realm:  "quotable",
			Params: map[string]string{
				"lower token": "token | quotable | empty",
			},
		},
	)
}
