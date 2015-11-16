package acme

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	dnsTemplate = "_acme-challenge.%s. 300 IN TXT \"%s\""
)

// dnsChallenge implements the dns-01 challenge according to ACME 7.5
type dnsChallenge struct {
	jws *jws
}

func (s *dnsChallenge) Solve(chlng challenge, domain string) error {

	logf("[INFO] acme: Trying to solve DNS-01")

	// Generate the Key Authorization for the challenge
	keyAuth, err := getKeyAuthorization(chlng.Token, &s.jws.privKey.PublicKey)
	if err != nil {
		return err
	}

	keyAuthShaBytes := sha256.Sum256([]byte(keyAuth))
	// FIXME: Currently boulder does not conform to the spec as in it uses hex encoding instead
	// of the base64 encoding mentioned by the spec. Fix this if either the spec or boulder changes!
	keyAuthSha := hex.EncodeToString(keyAuthShaBytes[:sha256.Size])

	dnsRecord := fmt.Sprintf(dnsTemplate, domain, keyAuthSha)
	logf("[DEBUG] acme: DNS Record: %s", dnsRecord)

	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')

	jsonBytes, err := json.Marshal(challenge{Resource: "challenge", Type: chlng.Type, Token: chlng.Token, KeyAuthorization: keyAuth})
	if err != nil {
		return errors.New("Failed to marshal network message...")
	}

	// Tell the server we handle DNS-01
	resp, err := s.jws.post(chlng.URI, jsonBytes)
	if err != nil {
		return fmt.Errorf("Failed to post JWS message. -> %v", err)
	}

	// Repeatedly check the server for an updated status on our request.
	var challengeResponse challenge
Loop:
	for {
		if resp.StatusCode >= http.StatusBadRequest {
			return handleHTTPError(resp)
		}

		err = json.NewDecoder(resp.Body).Decode(&challengeResponse)
		resp.Body.Close()
		if err != nil {
			return err
		}

		switch challengeResponse.Status {
		case "valid":
			logf("The server validated our request")
			break Loop
		case "pending":
			break
		case "invalid":
			return errors.New("The server could not validate our request.")
		default:
			return errors.New("The server returned an unexpected state.")
		}

		time.Sleep(1 * time.Second)
		resp, err = http.Get(chlng.URI)
	}

	return nil
}
