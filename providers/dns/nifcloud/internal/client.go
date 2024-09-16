package internal

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const (
	defaultBaseURL = "https://dns.api.nifcloud.com"
	apiVersion     = "2012-12-12N2013-12-16"
	// XMLNs XML NS of Route53.
	XMLNs = "https://route53.amazonaws.com/doc/2012-12-12/"
)

// Client the API client for NIFCLOUD DNS.
type Client struct {
	accessKey string
	secretKey string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient Creates a new client of NIFCLOUD DNS.
func NewClient(accessKey, secretKey string) (*Client, error) {
	if accessKey == "" || secretKey == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		accessKey:  accessKey,
		secretKey:  secretKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// ChangeResourceRecordSets Call ChangeResourceRecordSets API and return response.
func (c *Client) ChangeResourceRecordSets(ctx context.Context, hostedZoneID string, input ChangeResourceRecordSetsRequest) (*ChangeResourceRecordSetsResponse, error) {
	endpoint := c.BaseURL.JoinPath(apiVersion, "hostedzone", hostedZoneID, "rrset")

	req, err := newXMLRequest(ctx, http.MethodPost, endpoint, input)
	if err != nil {
		return nil, err
	}

	output := &ChangeResourceRecordSetsResponse{}
	err = c.do(req, output)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// GetChange Call GetChange API and return response.
func (c *Client) GetChange(ctx context.Context, statusID string) (*GetChangeResponse, error) {
	endpoint := c.BaseURL.JoinPath(apiVersion, "change", statusID)

	req, err := newXMLRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	output := &GetChangeResponse{}
	err = c.do(req, output)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (c *Client) do(req *http.Request, result any) error {
	err := c.sign(req)
	if err != nil {
		return fmt.Errorf("an error occurred during the creation of the signature: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return parseError(req, resp)
	}

	if result == nil {
		return nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	err = xml.Unmarshal(raw, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return nil
}

func (c *Client) sign(req *http.Request) error {
	if req.Header.Get("Date") == "" {
		req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	}

	if req.URL.Path == "" {
		req.URL.Path += "/"
	}

	mac := hmac.New(sha1.New, []byte(c.secretKey))
	_, err := mac.Write([]byte(req.Header.Get("Date")))
	if err != nil {
		return err
	}

	hashed := mac.Sum(nil)
	signature := base64.StdEncoding.EncodeToString(hashed)

	auth := fmt.Sprintf("NIFTY3-HTTPS NiftyAccessKeyId=%s,Algorithm=HmacSHA1,Signature=%s", c.accessKey, signature)
	req.Header.Set("X-Nifty-Authorization", auth)

	return nil
}

func newXMLRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
	body := new(bytes.Buffer)

	if payload != nil {
		body.WriteString(xml.Header)
		err := xml.NewEncoder(body).Encode(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request XML body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	if payload != nil {
		req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	}

	return req, nil
}

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	errResp := &ErrorResponse{}
	err := xml.Unmarshal(raw, errResp)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return errResp.Error
}
