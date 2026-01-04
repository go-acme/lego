package internal

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const algorithm = "HMAC-SHA256"

type Signer struct {
	secret  string
	agentID string
	clock   func() time.Time
}

// NewSigner creates a new Signer instance.
func NewSigner(secret, agentID string) (*Signer, error) {
	if secret == "" || agentID == "" {
		return nil, errors.New("credentials missing")
	}

	return &Signer{
		secret:  secret,
		agentID: agentID,
		clock:   time.Now,
	}, nil
}

// Sign signs the request.
// https://apidoc.xin.cn/doc-7283837
// https://apidoc.xin.cn/doc-7283838
func (s *Signer) Sign(req *http.Request) error {
	reqBody, err := drainBody(req)
	if err != nil {
		return err
	}

	timestamp := s.clock().UTC().Format("20060102T150405Z")

	stringToSign := algorithm + "\n" + timestamp + "\n" + req.Method + "\n" + req.URL.Path + "\n" + string(reqBody)

	h := hmac.New(sha256.New, []byte(s.secret))
	h.Write([]byte(stringToSign))

	signature := hex.EncodeToString(h.Sum(nil))

	req.Header.Set("Timestamp", timestamp)
	req.Header.Set("Authorization", fmt.Sprintf("%s Access=%s, Signature=%s", algorithm, s.agentID, signature))

	return nil
}

func drainBody(req *http.Request) ([]byte, error) {
	if req.Body == nil || req.Body == http.NoBody {
		return []byte(`{}`), nil
	}

	all, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	req.Body = io.NopCloser(bytes.NewBuffer(all))

	return all, nil
}
