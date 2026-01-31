package errutils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

const legoDebugClientVerboseError = "LEGO_DEBUG_CLIENT_VERBOSE_ERROR"

// HTTPDoError uses with `(http.Client).Do` error.
type HTTPDoError struct {
	req *http.Request
	err error
}

// NewHTTPDoError creates a new HTTPDoError.
func NewHTTPDoError(req *http.Request, err error) *HTTPDoError {
	return &HTTPDoError{req: req, err: err}
}

func (h HTTPDoError) Error() string {
	msg := "unable to communicate with the API server:"

	if ok, _ := strconv.ParseBool(os.Getenv(legoDebugClientVerboseError)); ok {
		msg += fmt.Sprintf(" [request: %s %s]", h.req.Method, h.req.URL)
	}

	if h.err == nil {
		return msg
	}

	return msg + fmt.Sprintf(" error: %v", h.err)
}

func (h HTTPDoError) Unwrap() error {
	return h.err
}

// ReadResponseError use with `io.ReadAll` when reading response body.
type ReadResponseError struct {
	req        *http.Request
	StatusCode int
	err        error
}

// NewReadResponseError creates a new ReadResponseError.
func NewReadResponseError(req *http.Request, statusCode int, err error) *ReadResponseError {
	return &ReadResponseError{req: req, StatusCode: statusCode, err: err}
}

func (r ReadResponseError) Error() string {
	msg := "unable to read response body:"

	if ok, _ := strconv.ParseBool(os.Getenv(legoDebugClientVerboseError)); ok {
		msg += fmt.Sprintf(" [request: %s %s]", r.req.Method, r.req.URL)
	}

	msg += fmt.Sprintf(" [status code: %d]", r.StatusCode)

	if r.err == nil {
		return msg
	}

	return msg + fmt.Sprintf(" error: %v", r.err)
}

func (r ReadResponseError) Unwrap() error {
	return r.err
}

// UnmarshalError uses with `json.Unmarshal` or `xml.Unmarshal` when reading response body.
type UnmarshalError struct {
	req        *http.Request
	StatusCode int
	Body       []byte
	err        error
}

// NewUnmarshalError creates a new UnmarshalError.
func NewUnmarshalError(req *http.Request, statusCode int, body []byte, err error) *UnmarshalError {
	return &UnmarshalError{req: req, StatusCode: statusCode, Body: bytes.TrimSpace(body), err: err}
}

func (u UnmarshalError) Error() string {
	msg := "unable to unmarshal response:"

	if ok, _ := strconv.ParseBool(os.Getenv(legoDebugClientVerboseError)); ok {
		msg += fmt.Sprintf(" [request: %s %s]", u.req.Method, u.req.URL)
	}

	msg += fmt.Sprintf(" [status code: %d] body: %s", u.StatusCode, string(u.Body))

	if u.err == nil {
		return msg
	}

	return msg + fmt.Sprintf(" error: %v", u.err)
}

func (u UnmarshalError) Unwrap() error {
	return u.err
}

// UnexpectedStatusCodeError use when the status of the response is unexpected but there is no API error type.
type UnexpectedStatusCodeError struct {
	req        *http.Request
	StatusCode int
	Body       []byte
}

// NewUnexpectedStatusCodeError creates a new UnexpectedStatusCodeError.
func NewUnexpectedStatusCodeError(req *http.Request, statusCode int, body []byte) *UnexpectedStatusCodeError {
	return &UnexpectedStatusCodeError{req: req, StatusCode: statusCode, Body: bytes.TrimSpace(body)}
}

func NewUnexpectedResponseStatusCodeError(req *http.Request, resp *http.Response) *UnexpectedStatusCodeError {
	raw, _ := io.ReadAll(resp.Body)
	return &UnexpectedStatusCodeError{req: req, StatusCode: resp.StatusCode, Body: bytes.TrimSpace(raw)}
}

func (u UnexpectedStatusCodeError) Error() string {
	msg := "unexpected status code:"

	if ok, _ := strconv.ParseBool(os.Getenv(legoDebugClientVerboseError)); ok {
		msg += fmt.Sprintf(" [request: %s %s]", u.req.Method, u.req.URL)
	}

	return msg + fmt.Sprintf(" [status code: %d] body: %s", u.StatusCode, string(u.Body))
}
