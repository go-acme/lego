package nonces

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/xenolf/lego/le/api/internal/sender"
)

type Manager struct {
	do       *sender.Do
	nonceURL string
	nonces   []string
	sync.Mutex
}

func NewManager(do *sender.Do, nonceURL string) *Manager {
	return &Manager{
		do:       do,
		nonceURL: nonceURL,
	}
}

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

func (n *Manager) Push(nonce string) {
	n.Lock()
	defer n.Unlock()
	n.nonces = append(n.nonces, nonce)
}

// Nonce implement jose.NonceSource
func (n *Manager) Nonce() (string, error) {
	if nonce, ok := n.Pop(); ok {
		return nonce, nil
	}
	return n.getNonce()
}

func (n *Manager) getNonce() (string, error) {
	resp, err := n.do.Head(n.nonceURL)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce from HTTP HEAD -> %v", err)
	}

	return GetFromResponse(resp)
}

func GetFromResponse(resp *http.Response) (string, error) {
	if resp == nil {
		return "", errors.New("nil response")
	}

	nonce := resp.Header.Get("Replay-Nonce")
	if nonce == "" {
		return "", fmt.Errorf("server did not respond with a proper nonce header")
	}

	return nonce, nil
}
