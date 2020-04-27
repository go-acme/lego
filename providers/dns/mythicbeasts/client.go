package mythicbeasts

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type authResponse struct {
	// The bearer token for use in API requests
	Token string `json:"access_token"`

	// The maximum lifetime of the token in seconds
	Lifetime int `json:"expires_in"`

	// The token type (must be 'bearer')
	TokenType string `json:"token_type"`
}

type authResponseError struct {
	// The error
	Error string `json:"error"`
	// A description of the error
	ErrorDescription string `json:"error_description"`
}

// Logs into mythic beasts and acquires a bearer token for use in future
// API calls
func (d *DNSProvider) login() error {
	if d.token != "" {
		// Already authenticated, stop now
		return nil
	}
	sendbody := strings.NewReader("grant_type=client_credentials")

	req, err := http.NewRequest("POST", d.config.AuthAPIEndpoint.String(), sendbody)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	req.SetBasicAuth(d.config.UserName, d.config.Password)

	resp, err := d.config.HTTPClient.Do(req)

	if err != nil {
		return err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, readErr := ioutil.ReadAll(resp.Body)

	if readErr != nil {
		return fmt.Errorf("login: %w", readErr)
	}

	if resp.StatusCode != 200 {
		if resp.StatusCode >= 400 && resp.StatusCode <= 499 {
			// Returned body should be a JSON thing
			err := authResponseError{}
			jsonErr := json.Unmarshal(body, &err)
			if jsonErr != nil {
				return fmt.Errorf("login: Error parsing error: %w", jsonErr)
			}
			return fmt.Errorf("login: %d: %s: %s", resp.StatusCode, err.Error, err.ErrorDescription)
		}
		return fmt.Errorf("login: Unknown error in auth API: %d", resp.StatusCode)
	}

	authresp := authResponse{}
	jsonErr := json.Unmarshal(body, &authresp)
	if jsonErr != nil {
		return fmt.Errorf("login: Error parsing response: %w", jsonErr)
	}

	if authresp.TokenType != "bearer" {
		return fmt.Errorf("login: Received unexpected token type: %s", authresp.TokenType)
	}

	d.token = authresp.Token
	return nil // Success
}

func (d *DNSProvider) createTXTRecord(zone string, leaf string, value string) error {
	return fmt.Errorf("mythicbeasts: createTXTRecord() not implemented")
}

func (d *DNSProvider) removeTXTRecord(zone string, leaf string, value string) error {
	return fmt.Errorf("mythicbeasts: removeTXTRecord() not implemented")
}
