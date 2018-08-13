package egoscale

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Error formats a CloudStack error into a standard error
func (e ErrorResponse) Error() string {
	return fmt.Sprintf("API error %s %d (%s %d): %s", e.ErrorCode, e.ErrorCode, e.CSErrorCode, e.CSErrorCode, e.ErrorText)
}

// Error formats a CloudStack job response into a standard error
func (e booleanResponse) Error() error {
	if !e.Success {
		return fmt.Errorf("API error: %s", e.DisplayText)
	}

	return nil
}

func (client *Client) parseResponse(resp *http.Response, key string) (json.RawMessage, error) {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	m := map[string]json.RawMessage{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	response, ok := m[key]
	if !ok {
		if resp.StatusCode >= 400 {
			response, ok = m["errorresponse"]
		}
		if !ok {
			for k := range m {
				return nil, fmt.Errorf("malformed JSON response, %q was expected, got %q", key, k)
			}
		}
	}

	if resp.StatusCode >= 400 {
		errorResponse := new(ErrorResponse)
		if e := json.Unmarshal(response, errorResponse); e != nil && errorResponse.ErrorCode <= 0 {
			return nil, fmt.Errorf("%d %s", resp.StatusCode, b)
		}
		return nil, errorResponse
	}

	n := map[string]json.RawMessage{}
	if err := json.Unmarshal(response, &n); err != nil {
		return nil, err
	}

	if len(n) > 1 {
		return response, nil
	}

	if len(n) == 1 {
		for k := range n {
			// boolean response and asyncjob result may also contain
			// only one key
			if k == "success" || k == "jobid" {
				return response, nil
			}
			return n[k], nil
		}
	}

	return response, nil
}

// asyncRequest perform an asynchronous job with a context
func (client *Client) asyncRequest(ctx context.Context, asyncCommand AsyncCommand) (interface{}, error) {
	var err error

	resp := asyncCommand.asyncResponse()
	client.AsyncRequestWithContext(
		ctx,
		asyncCommand,
		func(j *AsyncJobResult, e error) bool {
			if e != nil {
				err = e
				return false
			}
			if j.JobStatus == Success {
				if r := j.Result(resp); r != nil {
					err = r
				}
				return false
			}
			return true
		},
	)
	return resp, err
}

// SyncRequestWithContext performs a sync request with a context
func (client *Client) SyncRequestWithContext(ctx context.Context, command Command) (interface{}, error) {
	body, err := client.request(ctx, command)
	if err != nil {
		return nil, err
	}

	response := command.response()
	b, ok := response.(*booleanResponse)
	if ok {
		m := make(map[string]interface{})
		if errUnmarshal := json.Unmarshal(body, &m); errUnmarshal != nil {
			return nil, errUnmarshal
		}

		b.DisplayText, _ = m["displaytext"].(string)

		if success, okSuccess := m["success"].(string); okSuccess {
			b.Success = success == "true"
		}

		if success, okSuccess := m["success"].(bool); okSuccess {
			b.Success = success
		}

		return b, nil
	}

	if err := json.Unmarshal(body, response); err != nil {
		errResponse := new(ErrorResponse)
		if e := json.Unmarshal(body, errResponse); e == nil && errResponse.ErrorCode > 0 {
			return errResponse, nil
		}
		return nil, err
	}

	return response, nil
}

// BooleanRequest performs the given boolean command
func (client *Client) BooleanRequest(command Command) error {
	resp, err := client.Request(command)
	if err != nil {
		return err
	}

	if b, ok := resp.(*booleanResponse); ok {
		return b.Error()
	}

	panic(fmt.Errorf("command %q is not a proper boolean response. %#v", client.APIName(command), resp))
}

// BooleanRequestWithContext performs the given boolean command
func (client *Client) BooleanRequestWithContext(ctx context.Context, command Command) error {
	resp, err := client.RequestWithContext(ctx, command)
	if err != nil {
		return err
	}

	if b, ok := resp.(*booleanResponse); ok {
		return b.Error()
	}

	panic(fmt.Errorf("command %q is not a proper boolean response. %#v", client.APIName(command), resp))
}

// Request performs the given command
func (client *Client) Request(command Command) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	return client.RequestWithContext(ctx, command)
}

// RequestWithContext preforms a command with a context
func (client *Client) RequestWithContext(ctx context.Context, command Command) (interface{}, error) {
	switch command.(type) {
	case AsyncCommand:
		return client.asyncRequest(ctx, command.(AsyncCommand))
	default:
		return client.SyncRequestWithContext(ctx, command)
	}
}

// SyncRequest performs the command as is
func (client *Client) SyncRequest(command Command) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	return client.SyncRequestWithContext(ctx, command)
}

// AsyncRequest performs the given command
func (client *Client) AsyncRequest(asyncCommand AsyncCommand, callback WaitAsyncJobResultFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	client.AsyncRequestWithContext(ctx, asyncCommand, callback)
}

// AsyncRequestWithContext preforms a request with a context
func (client *Client) AsyncRequestWithContext(ctx context.Context, asyncCommand AsyncCommand, callback WaitAsyncJobResultFunc) {
	result, err := client.SyncRequestWithContext(ctx, asyncCommand)
	if err != nil {
		if !callback(nil, err) {
			return
		}
	}

	jobResult, ok := result.(*AsyncJobResult)
	if !ok {
		callback(nil, fmt.Errorf("wrong type, AsyncJobResult was expected instead of %T", result))
	}

	// Successful response
	if jobResult.JobID == "" || jobResult.JobStatus != Pending {
		callback(jobResult, nil)
		// without a JobID, the next requests will only fail
		return
	}

	for iteration := 0; ; iteration++ {
		time.Sleep(client.RetryStrategy(int64(iteration)))

		req := &QueryAsyncJobResult{JobID: jobResult.JobID}
		resp, err := client.SyncRequestWithContext(ctx, req)
		if err != nil && !callback(nil, err) {
			return
		}

		result, ok := resp.(*AsyncJobResult)
		if !ok {
			if !callback(nil, fmt.Errorf("wrong type. AsyncJobResult expected, got %T", resp)) {
				return
			}
		}

		if result.JobStatus == Failure {
			if !callback(nil, result.Error()) {
				return
			}
		} else {
			if !callback(result, nil) {
				return
			}
		}
	}
}

// Payload builds the HTTP request from the given command
func (client *Client) Payload(command Command) (string, error) {
	params := url.Values{}
	err := prepareValues("", params, command)
	if err != nil {
		return "", err
	}
	if hookReq, ok := command.(onBeforeHook); ok {
		if err := hookReq.onBeforeSend(params); err != nil {
			return "", err
		}
	}
	params.Set("apikey", client.APIKey)
	params.Set("command", client.APIName(command))
	params.Set("response", "json")

	// This code is borrowed from net/url/url.go
	// The way it's encoded by net/url doesn't match
	// how CloudStack work.
	var buf bytes.Buffer
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		prefix := csEncode(k) + "="
		for _, v := range params[k] {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(prefix)
			buf.WriteString(csEncode(v))
		}
	}

	return buf.String(), nil
}

// Sign signs the HTTP request and return it
func (client *Client) Sign(query string) (string, error) {
	mac := hmac.New(sha1.New, []byte(client.apiSecret))
	_, err := mac.Write([]byte(strings.ToLower(query)))
	if err != nil {
		return "", err
	}

	signature := csEncode(base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	return fmt.Sprintf("%s&signature=%s", csQuotePlus(query), signature), nil
}

// request makes a Request while being close to the metal
func (client *Client) request(ctx context.Context, command Command) (json.RawMessage, error) {
	payload, err := client.Payload(command)
	if err != nil {
		return nil, err
	}
	query, err := client.Sign(payload)
	if err != nil {
		return nil, err
	}

	method := "GET"
	url := fmt.Sprintf("%s?%s", client.Endpoint, query)

	var body io.Reader
	// respect Internet Explorer limit of 2048
	if len(url) > 2048 {
		url = client.Endpoint
		method = "POST"
		body = strings.NewReader(query)
	}

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	request = request.WithContext(ctx)
	request.Header.Add("User-Agent", fmt.Sprintf("exoscale/egoscale (%v)", Version))

	if method == "POST" {
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Add("Content-Length", strconv.Itoa(len(query)))
	}

	resp, err := client.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // nolint: errcheck

	contentType := resp.Header.Get("content-type")

	if !strings.Contains(contentType, "application/json") {
		return nil, fmt.Errorf(`body content-type response expected "application/json", got %q`, contentType)
	}

	apiName := client.APIName(command)
	key := fmt.Sprintf("%sresponse", strings.ToLower(apiName))
	// XXX: addIpToNic, activateIp6 are kind of special
	if key == "addiptonicresponse" {
		key = "addiptovmnicresponse"
	} else if key == "activateip6response" {
		key = "activateip6nicresponse"
	}

	text, err := client.parseResponse(resp, key)
	if err != nil {
		return nil, err
	}

	return text, nil
}
