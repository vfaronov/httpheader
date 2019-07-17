# httpheader

[![GoDoc](https://godoc.org/github.com/vfaronov/httpheader?status.svg)](https://godoc.org/github.com/vfaronov/httpheader)
[![Travis build status](https://travis-ci.org/vfaronov/httpheader.svg?branch=master)](https://travis-ci.org/vfaronov/httpheader)
[![Codecov](https://codecov.io/gh/vfaronov/httpheader/branch/master/graph/badge.svg)](https://codecov.io/gh/vfaronov/httpheader)
[![Go Report Card](https://goreportcard.com/badge/github.com/vfaronov/httpheader)](https://goreportcard.com/report/github.com/vfaronov/httpheader)

This is a Go package to parse and generate standard HTTP headers correctly.
It knows about complex headers like 
[`Accept`](https://tools.ietf.org/html/rfc7231#section-5.3.2),
[`Prefer`](https://tools.ietf.org/html/rfc7240),
[`Link`](https://tools.ietf.org/html/rfc8288#section-3).
Unlike many other implementations, it handles all the tricky bits of syntax like
[quoted](https://tools.ietf.org/html/rfc7230#section-3.2.6) commas,
[multiple](https://tools.ietf.org/html/rfc7230#section-3.2.2) header lines,
[Unicode](https://tools.ietf.org/html/rfc8187) parameters.
It gives you convenient structures to work with, and can serialize them back 
into HTTP.

This package is distributed under the MIT license, and hosted on GitHub.
If it doesn't yet support a header that you need, feel free to open an issue there.


## Installation

	go get github.com/vfaronov/httpheader


## Example

	const request = `GET / HTTP/1.1
	Host: api.example.com
	User-Agent: MyApp/1.2.3 python-requests/2.22.0
	Accept: text/*, application/json;q=0.8
	Forwarded: for="198.51.100.30:14852";by="[2001:db8::ae:56]";proto=https
	
	`
	
	r, _ := http.ReadRequest(bufio.NewReader(strings.NewReader(request)))
	
	forwarded := httpheader.Forwarded(r.Header)
	fmt.Println("received request from user at", forwarded[0].For.IP)
	
	for _, product := range httpheader.UserAgent(r.Header) {
		if product.Name == "MyApp" && product.Version < "2.0" {
			fmt.Println("enabling compatibility mode for", product)
		}
	}
	
	accept := httpheader.Accept(r.Header)
	acceptJSON := httpheader.MatchAccept(accept, "application/json")
	acceptXML := httpheader.MatchAccept(accept, "text/xml")
	if acceptXML.Q > acceptJSON.Q {
		fmt.Println("responding with XML")
	}
	
	// Output: received request from user at 198.51.100.30
	// enabling compatibility mode for {MyApp 1.2.3 }
	// responding with XML
