package nearlyfreespeech

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	apiURL    = "api.nearlyfreespeech.net"
	saltBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type TXTRecord struct {
	Name string
	Type string
	Data string
	TTL  int
}

type ClientConfig struct {
	domain string
}

func (d *DNSProvider) authHeader(reqURI string, body []byte) string {
	// Generate salt.
	// This is the only part of this that needs to be serialized.
	salt := make([]byte, 16)
	d.rngMu.Lock()
	for i := 0; i < 16; i++ {
		salt[i] = saltBytes[rand.Intn(len(saltBytes))]
	}
	d.rngMu.Unlock()

	return genAuthHeader(d.config.Login, time.Now(), salt, d.config.APIKey, reqURI, body)
}

func genAuthHeader(user string, t time.Time, salt []byte, apiKey, reqURI string, body []byte) string {
	// Header is "login;timestamp;salt;hash".
	// hash is SHA1("login;timestamp;salt;api-key;request-uri;body-hash")
	// and body-hash is SHA1(body).
	bodyHash := sha1.Sum(body)
	ts := strconv.FormatInt(t.Unix(), 10)
	hashInput := fmt.Sprintf("%s;%s;%s;%s;%s;%02x", user, ts, salt, apiKey, reqURI, bodyHash)

	return fmt.Sprintf("%s;%s;%s;%02x", user, ts, salt, sha1.Sum([]byte(hashInput)))
}

func (d *DNSProvider) do(path string, body interface{}) error {
	hdr := make(http.Header)

	var encBody []byte
	switch x := body.(type) {
	case nil:
	case []byte:
		encBody = x
	case url.Values:
		encBody = []byte(x.Encode())
		hdr.Set("Content-Type", "application/x-www-form-urlencoded")
	default:
		panic(fmt.Sprintf("invalid body type %T", x))
	}

	req := &http.Request{
		Method: "POST",
		URL: &url.URL{
			Scheme: "https",
			Host:   apiURL,
			Path:   path,
		},
		Header: hdr,
		Body:   io.NopCloser(bytes.NewReader(encBody)),
	}
	req.Header.Set("X-NFSN-Authentication", d.authHeader(req.URL.Path, encBody))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("%s: %s", resp.Status, respBody)
	}

	// Success
	return nil
}

func (d *DNSProvider) AddRR(cc ClientConfig, rr TXTRecord) error {
	args := url.Values{
		"name": []string{rr.Name},
		"type": []string{rr.Type},
		"data": []string{rr.Data},
	}
	if rr.TTL > 0 {
		args["ttl"] = []string{strconv.Itoa(rr.TTL)}
	}
	err := d.do("/dns/"+url.PathEscape(cc.domain)+"/addRR", args)
	return err
}

func (d *DNSProvider) DeleteRR(cc ClientConfig, rr TXTRecord) error {
	args := url.Values{
		"name": []string{rr.Name},
		"type": []string{rr.Type},
		"data": []string{rr.Data},
	}
	err := d.do("/dns/"+url.PathEscape(cc.domain)+"/removeRR", args)
	return err
}
