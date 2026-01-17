package nonces

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-acme/lego/v5/acme/api/internal/sender"
	"github.com/go-jose/go-jose/v4"
)

// Manager Manages nonces.
type Manager struct {
	sync.Mutex

	do       *sender.Doer
	nonceURL string
	nonces   []string
}

// NewManager Creates a new Manager.
func NewManager(do *sender.Doer, nonceURL string) *Manager {
	return &Manager{
		do:       do,
		nonceURL: nonceURL,
	}
}

// Push Pushes nonce.
func (n *Manager) Push(nonce string) {
	n.Lock()
	defer n.Unlock()

	n.nonces = append(n.nonces, nonce)
}

// Pop Pops a nonce.
func (n *Manager) Pop() (string, bool) {
	n.Lock()
	defer n.Unlock()

	if len(n.nonces) == 0 {
		return "", false
	}

	nonce := n.nonces[len(n.nonces)-1]
	n.nonces = n.nonces[:len(n.nonces)-1]

	return nonce, true
}

func (n *Manager) getNonce(ctx context.Context) (string, error) {
	if nonce, ok := n.Pop(); ok {
		return nonce, nil
	}

	resp, err := n.do.Head(ctx, n.nonceURL)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce from HTTP HEAD: %w", err)
	}

	return GetFromResponse(resp)
}

// GetFromResponse Extracts nonce from an HTTP response.
func GetFromResponse(resp *http.Response) (string, error) {
	if resp == nil {
		return "", errors.New("nil response")
	}

	nonce := resp.Header.Get("Replay-Nonce")
	if nonce == "" {
		return "", errors.New("server did not respond with a proper nonce header")
	}

	return nonce, nil
}

var _ jose.NonceSource = (*NonceSource)(nil)

// NonceSource implements [jose.NonceSource].
//
//nolint:containedctx // This is the only way to use the context in this case.
type NonceSource struct {
	ctx context.Context
	m   *Manager
}

// NewNonceSource Creates a new NonceSource.
func NewNonceSource(ctx context.Context, manager *Manager) *NonceSource {
	return &NonceSource{ctx: ctx, m: manager}
}

func (n *NonceSource) Nonce() (string, error) {
	return n.m.getNonce(n.ctx)
}
