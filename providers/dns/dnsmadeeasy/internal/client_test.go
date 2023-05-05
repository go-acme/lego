package internal

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_sign(t *testing.T) {
	apiKey := "key"

	client := Client{apiKey: apiKey, apiSecret: "secret"}

	req, err := http.NewRequest(http.MethodGet, "", http.NoBody)
	require.NoError(t, err)

	timestamp := time.Date(2015, time.June, 2, 2, 36, 7, 0, time.UTC).Format(time.RFC1123)

	err = client.sign(req, timestamp)
	require.NoError(t, err)

	assert.Equal(t, apiKey, req.Header.Get("x-dnsme-apiKey"))
	assert.Equal(t, timestamp, req.Header.Get("x-dnsme-requestDate"))
	assert.Equal(t, "6b6c8432119c31e1d3776eb4cd3abd92fae4a71c", req.Header.Get("x-dnsme-hmac"))
}
