package httpheader_test

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"

	"github.com/vfaronov/httpheader"
)

func Example() {
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
}
