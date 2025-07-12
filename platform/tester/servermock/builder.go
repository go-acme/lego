package servermock

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

// Link represents a middleware interface, enabling middleware chaining.
type Link interface {
	Bind(next http.Handler) http.Handler
}

// LinkFunc defines a function type [Link].
type LinkFunc func(next http.Handler) http.Handler

func (f LinkFunc) Bind(next http.Handler) http.Handler {
	return f(next)
}

// ClientBuilder defines a function type for creating a client of type T based on a httptest.Server instance.
type ClientBuilder[T any] func(server *httptest.Server) (T, error)

// Builder is a type that facilitates the construction of testable HTTP clients and server.
// It allows defining routes, attaching middleware, and creating custom HTTP clients.
type Builder[T any] struct {
	mux   *http.ServeMux
	chain []Link

	clientBuilder ClientBuilder[T]
}

func NewBuilder[T any](clientBuilder ClientBuilder[T], chain ...Link) *Builder[T] {
	return &Builder[T]{
		mux:           http.NewServeMux(),
		chain:         chain,
		clientBuilder: clientBuilder,
	}
}

func (b *Builder[T]) Route(pattern string, handler http.Handler, chain ...Link) *Builder[T] {
	if handler == nil {
		handler = Noop()
	}

	for _, link := range slices.Backward(b.chain) {
		handler = link.Bind(handler)
	}

	for _, link := range slices.Backward(chain) {
		handler = link.Bind(handler)
	}

	b.mux.Handle(pattern, handler)

	return b
}

func (b *Builder[T]) Build(t *testing.T) T {
	t.Helper()

	server := httptest.NewServer(b.mux)
	t.Cleanup(server.Close)

	client, err := b.clientBuilder(server)
	require.NoError(t, err)

	return client
}
