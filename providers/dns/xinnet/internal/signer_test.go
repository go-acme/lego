package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSigner_Sign(t *testing.T) {
	signer, err := NewSigner("9f3b6760952b41eda7ddad84d755b3944cbb7929277947c294b7a578128e1170", "agent12345")
	require.NoError(t, err)

	signer.clock = func() time.Time {
		return time.Date(2024, 6, 5, 8, 55, 45, 0, time.UTC)
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/domain/check/", http.NoBody)

	err = signer.Sign(req)
	require.NoError(t, err)

	assert.Equal(t, "20240605T085545Z", req.Header.Get("Timestamp"))
	assert.Equal(t,
		"HMAC-SHA256 Access=agent12345, Signature=bbcebdf1d93fbd8dfbd18313b29e645014c98e307d8f3a88b7517199bad564a6",
		req.Header.Get("Authorization"),
	)
}
