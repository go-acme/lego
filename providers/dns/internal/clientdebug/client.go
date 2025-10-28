package clientdebug

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
)

type DumpTransport struct {
	rt http.RoundTripper
}

func NewDumpTransport(rt http.RoundTripper) *DumpTransport {
	if rt == nil {
		rt = http.DefaultTransport
	}

	return &DumpTransport{rt: rt}
}

func (d *DumpTransport) RoundTrip(h *http.Request) (*http.Response, error) {
	dump, _ := httputil.DumpRequestOut(h, true)

	fmt.Println("[HTTP Request]")
	fmt.Println(string(dump))

	resp, err := d.rt.RoundTrip(h)

	dump, _ = httputil.DumpResponse(resp, true)

	fmt.Println("[HTTP Response]")
	fmt.Println(string(dump))

	return resp, err
}

// Wrap wraps an HTTP client Transport with the [DumpTransport].
func Wrap(client *http.Client) *http.Client {
	val, found := os.LookupEnv("LEGO_DEBUG_DNS_API_HTTP_CLIENT")
	if !found {
		return client
	}

	if ok, _ := strconv.ParseBool(val); !ok {
		return client
	}

	client.Transport = NewDumpTransport(client.Transport)

	return client
}
