package acme

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type preCheckDNSFunc func(domain, fqdn string) bool

var preCheckDNS preCheckDNSFunc = checkDNS

var preCheckDNSFallbackCount = 5

// DNSProvider represents a service for managing DNS records.
type DNSProvider interface {
	CreateTXTRecord(fqdn, value string, ttl int) error
	RemoveTXTRecord(fqdn, value string, ttl int) error
}

// dnsChallenge implements the dns-01 challenge according to ACME 7.5
type dnsChallenge struct {
	jws      *jws
	provider DNSProvider
}

func (s *dnsChallenge) Solve(chlng challenge, domain string) error {

	logf("[INFO] acme: Trying to solve DNS-01")

	// Generate the Key Authorization for the challenge
	keyAuth, err := getKeyAuthorization(chlng.Token, &s.jws.privKey.PublicKey)
	if err != nil {
		return err
	}

	keyAuthShaBytes := sha256.Sum256([]byte(keyAuth))
	// base64URL encoding without padding
	keyAuthSha := base64.URLEncoding.EncodeToString(keyAuthShaBytes[:sha256.Size])
	keyAuthSha = strings.TrimRight(keyAuthSha, "=")

	fqdn := fmt.Sprintf("_acme-challenge.%s.", domain)
	if err = s.provider.CreateTXTRecord(fqdn, keyAuthSha, 120); err != nil {
		return err
	}

	preCheckDNS(domain, fqdn)

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

	if err = s.provider.RemoveTXTRecord(fqdn, keyAuthSha, 120); err != nil {
		logf("[WARN] acme: Failed to cleanup DNS record. -> %v ", err)
	}

	return nil
}

func checkDNS(domain, fqdn string) bool {
	// check if the expected DNS entry was created. If not wait for some time and try again.
	m := new(dns.Msg)
	m.SetQuestion(domain+".", dns.TypeSOA)
	c := new(dns.Client)
	in, _, err := c.Exchange(m, "8.8.8.8:53")
	if err != nil {
		return false
	}

	var authorativeNS string
	for _, answ := range in.Answer {
		soa := answ.(*dns.SOA)
		authorativeNS = soa.Ns
	}

	fallbackCnt := 0
	for fallbackCnt < preCheckDNSFallbackCount {
		m.SetQuestion(fqdn, dns.TypeTXT)
		in, _, err = c.Exchange(m, authorativeNS+":53")
		if err != nil {
			return false
		}

		if len(in.Answer) > 0 {
			return true
		}

		fallbackCnt++
		if fallbackCnt >= preCheckDNSFallbackCount {
			return false
		}

		time.Sleep(time.Second * time.Duration(fallbackCnt))
	}

	return false
}
