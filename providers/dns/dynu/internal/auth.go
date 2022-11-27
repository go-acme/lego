package internal

import (
	"errors"
	"net/http"
)

const apiKeyHeader = "Api-Key"

// TokenTransport HTTP transport for API authentication.
type TokenTransport struct {
	apiKey string

	// Transport is the underlying HTTP transport to use when making requests.
	// It will default to http.DefaultTransport if nil.
	Transport http.RoundTripper
}

// NewTokenTransport Creates a HTTP transport for API authentication.
func NewTokenTransport(apiKey string) (*TokenTransport, error) {
	if apiKey == "" {
		return nil, errors.New("credentials missing: API key")
	}

	return &TokenTransport{apiKey: apiKey}, nil
}

// RoundTrip executes a single HTTP transaction.
func (t *TokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	enrichedReq := &http.Request{}
	*enrichedReq = *req

	enrichedReq.Header = make(http.Header, len(req.Header))
	for k, s := range req.Header {
		enrichedReq.Header[k] = append([]string(nil), s...)
	}

	if t.apiKey != "" {
		enrichedReq.Header.Set(apiKeyHeader, t.apiKey)
	}

	return t.transport().RoundTrip(enrichedReq)
}

func (t *TokenTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

// Client Creates a new HTTP client.
func (t *TokenTransport) Client() *http.Client {
	return &http.Client{Transport: t}
}

// Wrap wraps an HTTP client Transport with the TokenTransport.
func (t *TokenTransport) Wrap(client *http.Client) *http.Client {
	backup := client.Transport
	t.Transport = backup
	client.Transport = t

	return client
}
