package webroot

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-acme/lego/v5/challenge/http01"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPProvider(t *testing.T) {
	webroot := t.TempDir()

	domain := "domain"
	token := "token"
	keyAuth := "keyAuth"

	require.NoError(t, os.MkdirAll(filepath.Join(webroot, filepath.FromSlash(http01.PathPrefix)), 0o777))

	provider, err := NewHTTPProvider(webroot)
	require.NoError(t, err)

	err = provider.Present(t.Context(), domain, token, keyAuth)
	require.NoError(t, err)

	challengeFilePath := filepath.Join(webroot, filepath.FromSlash(http01.PathPrefix), token)

	require.FileExists(t, challengeFilePath, "Challenge file was not created in webroot")

	var data []byte

	data, err = os.ReadFile(challengeFilePath)
	require.NoError(t, err)

	dataStr := string(data)
	assert.Equal(t, keyAuth, dataStr)

	err = provider.CleanUp(t.Context(), domain, token, keyAuth)
	require.NoError(t, err)
}
