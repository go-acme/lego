package internal

import (
	"context"
	"testing"
)

func mockContext(t *testing.T) context.Context {
	t.Helper()

	return context.WithValue(t.Context(), tokenKey, "secret")
}
