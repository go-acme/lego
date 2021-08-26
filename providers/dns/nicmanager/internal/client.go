package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/log"
	"github.com/pquerna/otp/totp"
)

const (
	headerTOTPToken = "X-Auth-Token"
)

type NicManagerClient struct {
	Account  *string
	Username *string

	Email *string

	Password string
	OTP      *string

	Mode string

	baseURL string
	c       *http.Client
}

// NewNicManagerClient create a new client.
func NewNicManagerClient(c *http.Client) *NicManagerClient {
	return &NicManagerClient{
		Mode:    "anycast",
		baseURL: "https://api.nicmanager.com/v1",
		c:       c,
	}
}

// SetAccount Use account-based login.
func (n *NicManagerClient) SetAccount(account, username string) {
	n.Account = &account
	n.Username = &username
}

// SetEmail Use email-based login.
func (n *NicManagerClient) SetEmail(email string) {
	n.Email = &email
}

// SetOTP Set the TOTP Secret to use 2fa.
func (n *NicManagerClient) SetOTP(otp string) {
	n.OTP = &otp
}

// handleError Output Request body in error case.
func (n *NicManagerClient) handleError(res *http.Response, err error) error {
	b, er := ioutil.ReadAll(res.Body)
	if er != nil {
		log.Printf("nicmanager: failed to read response: %s", er.Error())
		b = []byte{}
	}
	log.Printf("nicmanager: error response: %s", string(b))
	return fmt.Errorf("HTTP Error '%w' during request '%s %s': \"%s\"", err, res.Request.Method, res.Request.URL.Path, string(b))
}

// Request Wrapper for all API Requests.
func (n *NicManagerClient) Request(method, url string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonValue, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonValue)
	}
	// https://api.nicmanager.com/docs/v1/
	r, err := http.NewRequest(method, fmt.Sprintf("%s%s", n.baseURL, url), reqBody)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Content-Type", "application/json")
	if n.Account != nil && n.Username != nil {
		r.SetBasicAuth(fmt.Sprintf("%s.%s", *n.Account, *n.Username), n.Password)
	} else {
		r.SetBasicAuth(*n.Email, n.Password)
	}
	if n.OTP != nil {
		tan, err := totp.GenerateCode(*n.OTP, time.Now())
		if err != nil {
			return nil, err
		}
		r.Header.Set(headerTOTPToken, tan)
	}
	return n.c.Do(r)
}

func (n *NicManagerClient) ZoneInfo(name string) (*Zone, error) {
	res, err := n.Request("GET", fmt.Sprintf("/%s/%s", n.Mode, name), nil)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, n.handleError(res, fmt.Errorf("nicmanager: failed to get zone info for %s", name))
	}
	var zone *Zone
	err = json.NewDecoder(res.Body).Decode(&zone)
	if err != nil {
		return nil, err
	}
	return zone, nil
}

func (n *NicManagerClient) ResourceRecordCreate(zone string, req RecordCreateUpdate) error {
	res, err := n.Request("POST", fmt.Sprintf("/%s/%s/records", n.Mode, zone), req)
	if err != nil {
		return err
	}
	if res.StatusCode != 202 {
		return n.handleError(res, fmt.Errorf("nicmanager: records create should've returned 202 but returned %d", res.StatusCode))
	}
	return nil
}

func (n *NicManagerClient) ResourceRecordDelete(zone string, record int) error {
	res, err := n.Request("DELETE", fmt.Sprintf("/%s/%s/records/%d", n.Mode, zone, record), nil)
	if err != nil {
		return err
	}
	if res.StatusCode != 202 {
		return n.handleError(res, fmt.Errorf("nicmanager: records delete should've returned 202 but returned %d", res.StatusCode))
	}
	return nil
}
