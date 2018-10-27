package acme

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
)

type nonceManager struct {
	nonces []string
	sync.Mutex
}

func (n *nonceManager) Pop() (string, bool) {
	n.Lock()
	defer n.Unlock()

	if len(n.nonces) == 0 {
		return "", false
	}

	nonce := n.nonces[len(n.nonces)-1]
	n.nonces = n.nonces[:len(n.nonces)-1]
	return nonce, true
}

func (n *nonceManager) Push(nonce string) {
	n.Lock()
	defer n.Unlock()
	n.nonces = append(n.nonces, nonce)
}

func getNonce(url string) (string, error) {
	resp, err := httpHead(url)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce from HTTP HEAD -> %v", err)
	}

	return getNonceFromResponse(resp)
}

func getNonceFromResponse(resp *http.Response) (string, error) {
	if resp == nil {
		return "", errors.New("nil response")
	}

	nonce := resp.Header.Get("Replay-Nonce")
	if nonce == "" {
		return "", fmt.Errorf("server did not respond with a proper nonce header")
	}

	return nonce, nil
}
