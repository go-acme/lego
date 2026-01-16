package nonces

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-acme/lego/v5/acme/api/internal/sender"
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

// Push Pushes a nonce.
func (n *Manager) Push(nonce string) {
	n.Lock()
	defer n.Unlock()

	n.nonces = append(n.nonces, nonce)
}

// Nonce implement jose.NonceSource.
func (n *Manager) Nonce() (string, error) {
	if nonce, ok := n.Pop(); ok {
		return nonce, nil
	}

	// TODO(ldez): the Nonce method signature cannot be changed because it must implement jose.NonceSource.
	// Maybe use a dirty context struct field in this case.
	return n.getNonce(context.Background())
}

func (n *Manager) getNonce(ctx context.Context) (string, error) {
	resp, err := n.do.Head(ctx, n.nonceURL)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce from HTTP HEAD: %w", err)
	}

	return GetFromResponse(resp)
}

// GetFromResponse Extracts a nonce from an HTTP response.
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
