package acme

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type HttpChallengeMethod interface {
	PresentToken(domain, token, keyAuth string, checkSolvedFunc func() error) (err error)
}

type httpChallenge struct {
	jws    *jws
	method HttpChallengeMethod
}

func (s *httpChallenge) Solve(chlng challenge, domain string) error {

	logf("[INFO] acme: Trying to solve HTTP-01")

	// Generate the Key Authorization for the challenge
	keyAuth, err := getKeyAuthorization(chlng.Token, &s.jws.privKey.PublicKey)
	if err != nil {
		return err
	}

	return s.method.PresentToken(domain, chlng.Token, keyAuth, func() error {
		return s.checkSolved(chlng.URI, chlng.Type, chlng.Type, keyAuth)
	})
}

func (s *httpChallenge) checkSolved(cURI, cType, cToken, keyAuth string) error {

	jsonBytes, err := json.Marshal(challenge{Resource: "challenge", Type: cType, Token: cToken, KeyAuthorization: keyAuth})
	if err != nil {
		return errors.New("Failed to marshal network message...")
	}

	// Tell the server we handle HTTP-01
	resp, err := s.jws.post(cURI, jsonBytes)
	if err != nil {
		return fmt.Errorf("Failed to post JWS message. -> %v", err)
	}

	// After the path is sent, the ACME server will access our server.
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
		resp, err = http.Get(cURI)
	}

	return nil
}
