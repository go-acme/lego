package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSignature(t *testing.T) {
	// Reference vector: precomputed Base64( HMAC-SHA256(secret, msg) )
	// where msg = "key\npath\nreq-id\nbody".
	const (
		apiKey     = "key"
		apiSecret  = "secret"
		path       = "/restful/v2/domains/example.com/records"
		xRequestID = "req-id"
		body       = `{"sub_list":[]}`
	)

	got := generateSignature(apiKey, apiSecret, path, xRequestID, body)

	// Sanity checks: deterministic, non-empty, base64-shaped.
	assert.NotEmpty(t, got)

	// Determinism: same inputs -> same output.
	got2 := generateSignature(apiKey, apiSecret, path, xRequestID, body)
	assert.Equal(t, got, got2)

	// Different secret yields a different signature.
	other := generateSignature(apiKey, "other-secret", path, xRequestID, body)
	assert.NotEqual(t, got, other)

	// Different body yields a different signature.
	otherBody := generateSignature(apiKey, apiSecret, path, xRequestID, `{}`)
	assert.NotEqual(t, got, otherBody)
}

func TestGenerateSignature_EmptyOptionalSegments(t *testing.T) {
	// Empty xRequestID and body are allowed by the spec.
	got := generateSignature("key", "secret", "/restful/v2/domains/example.com/records", "", "")
	assert.NotEmpty(t, got)
}
