package internal

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const securityTokenHeader = "x-cns-security-token"

// TokenTransport HTTP transport for API authentication.
type TokenTransport struct {
	apiKey    string
	secretKey string

	// Transport is the underlying HTTP transport to use when making requests.
	// It will default to http.DefaultTransport if nil.
	Transport http.RoundTripper
}

// NewTokenTransport Creates an HTTP transport for API authentication.
func NewTokenTransport(apiKey, secretKey string) (*TokenTransport, error) {
	if apiKey == "" {
		return nil, errors.New("credentials missing: API key")
	}

	if secretKey == "" {
		return nil, errors.New("credentials missing: secret key")
	}

	return &TokenTransport{apiKey: apiKey, secretKey: secretKey}, nil
}

// RoundTrip executes a single HTTP transaction.
func (t *TokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	enrichedReq := &http.Request{}
	*enrichedReq = *req

	enrichedReq.Header = make(http.Header, len(req.Header))
	for k, s := range req.Header {
		enrichedReq.Header[k] = append([]string(nil), s...)
	}

	if t.apiKey != "" && t.secretKey != "" {
		securityToken := createCnsSecurityToken(t.apiKey, t.secretKey)
		enrichedReq.Header.Set(securityTokenHeader, securityToken)
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

func createCnsSecurityToken(apiKey, secretKey string) string {
	timestamp := time.Now().Round(time.Millisecond).UnixNano() / int64(time.Millisecond)

	hm := encodedHmac(timestamp, secretKey)
	requestDate := strconv.FormatInt(timestamp, 10)

	return fmt.Sprintf("%s:%s:%s", apiKey, hm, requestDate)
}

func encodedHmac(message int64, secret string) string {
	h := hmac.New(sha1.New, []byte(secret))
	_, _ = h.Write([]byte(strconv.FormatInt(message, 10)))

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
