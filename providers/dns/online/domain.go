package online

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type OnlinedomainAPI struct {
	Token  string
	client *http.Client
	API    string
}

const (
	Endpoint = "https://api.online.net"
)

// create a new api
func NewOnlinedomainAPI(token string) (*OnlinedomainAPI, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: os.Getenv("ONLINEDOMAIN_INSECURE") != "",
		},
	}

	d := &OnlinedomainAPI{
		Token:  token,
		client: &http.Client{Transport: tr},
		API:    Endpoint,
	}

	if apiURL := os.Getenv("ONLINE_DOMAIN_API"); apiURL != "" {
		d.API = apiURL
	}

	return d, nil
}

// GetResponsePaginate fetchs all resources and returns an http.Response object for the requested resource
func (d *OnlinedomainAPI) GetResponse(apiURL, resource string, values url.Values) (*http.Response, error) {
	trimAPIURL := strings.TrimRight(apiURL, "/")

	var resp *http.Response
	var err error

	if len(values) == 0 {
		resp, err = d.response("GET", fmt.Sprintf("%s/%s", trimAPIURL, resource), nil)
	} else {
		resp, err = d.response("GET", fmt.Sprintf("%s/%s?%s", trimAPIURL, resource, values.Encode()), nil)
	}

	if err != nil {
		return nil, err
	}

	return resp, err
}

func (d *OnlinedomainAPI) response(method, uri string, content io.Reader) (resp *http.Response, err error) {
	var (
		req *http.Request
	)

	req, err = http.NewRequest(method, uri, content)
	if err != nil {
		err = fmt.Errorf("response %s %s", method, uri)
		return
	}
	req.Header.Set("Authorization", "Bearer "+d.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = d.client.Do(req)

	return
}

// PatchResponse returns an http.Response object for the updated resource
func (d *OnlinedomainAPI) PatchResponse(apiURL, resource string, data interface{}) (*http.Response, error) {
	payload := new(bytes.Buffer)
	if err := json.NewEncoder(payload).Encode(data); err != nil {
		return nil, err
	}
	fmt.Printf("%s", payload.String())
	fmt.Printf("%s/%s\n", strings.TrimRight(apiURL, "/"), resource)
	return d.response("PATCH", fmt.Sprintf("%s/%s", strings.TrimRight(apiURL, "/"), resource), payload)
}

// handleHTTPError checks the statusCode and displays the error
func (d *OnlinedomainAPI) handleHTTPError(goodStatusCode []int, resp *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusInternalServerError {
		return nil, fmt.Errorf(string(body))
	}
	good := false
	for _, code := range goodStatusCode {
		if code == resp.StatusCode {
			good = true
		}
	}
	if !good {
		return nil, fmt.Errorf("unexpected status: %v", resp.StatusCode)
	}
	return body, nil
}
