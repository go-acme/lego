package servermock

import (
	"fmt"
	"net/http"
	"regexp"
	"slices"
)

const (
	authorizationHeader = "Authorization"
	contentTypeHeader   = "Content-Type"
	acceptHeader        = "Accept"
)

const (
	applicationJSONMimeType = "application/json"
	applicationFormMimeType = "application/x-www-form-urlencoded"
)

type basicAuth struct {
	username, password string
}

// HeaderLink validates HTTP request headers.
type HeaderLink struct {
	values     http.Header
	regexes    map[string]*regexp.Regexp
	json       bool
	basicAuth  *basicAuth
	statusCode int
}

func CheckHeader() *HeaderLink {
	return &HeaderLink{
		values:     http.Header{},
		regexes:    map[string]*regexp.Regexp{},
		statusCode: http.StatusBadRequest,
	}
}

func (l *HeaderLink) Bind(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		for k, v := range l.values {
			err := checkHeader(req, k, v)
			if err != nil {
				http.Error(rw, err.Error(), l.statusCode)
				return
			}
		}

		for k, exp := range l.regexes {
			value := req.Header.Get(k)

			if !exp.MatchString(value) {
				msg := fmt.Sprintf("invalid %q header value, %q doesn't match to %q", k, value, exp)
				http.Error(rw, msg, l.statusCode)

				return
			}
		}

		if l.json && !l.checkJSONHeaders(rw, req) {
			return
		}

		if l.basicAuth != nil && !l.checkBasicAuth(rw, req) {
			return
		}

		next.ServeHTTP(rw, req)
	})
}

func (l *HeaderLink) With(name, value string, values ...string) *HeaderLink {
	for _, v := range slices.Concat([]string{value}, values) {
		l.values.Add(name, v)
	}

	return l
}

func (l *HeaderLink) WithRegexp(name, exp string) *HeaderLink {
	l.regexes[name] = regexp.MustCompile(exp)

	return l
}

func (l *HeaderLink) WithJSONHeaders() *HeaderLink {
	l.json = true

	return l
}

func (l *HeaderLink) WithContentTypeFromURLEncoded() *HeaderLink {
	l.values.Set(contentTypeHeader, applicationFormMimeType)

	return l
}

func (l *HeaderLink) WithContentType(value string) *HeaderLink {
	l.values.Set(contentTypeHeader, value)

	return l
}

func (l *HeaderLink) WithAccept(value string) *HeaderLink {
	l.values.Set(acceptHeader, value)

	return l
}

func (l *HeaderLink) WithAuthorization(value string) *HeaderLink {
	l.values.Set(authorizationHeader, value)

	return l
}

func (l *HeaderLink) WithStatusCode(status int) *HeaderLink {
	if l.statusCode >= http.StatusContinue {
		l.statusCode = status
	}

	return l
}

func (l *HeaderLink) WithBasicAuth(username, password string) *HeaderLink {
	l.basicAuth = &basicAuth{username: username, password: password}

	return l
}

func (l *HeaderLink) checkBasicAuth(rw http.ResponseWriter, req *http.Request) bool {
	usr, pwd, ok := req.BasicAuth()
	if !ok {
		http.Error(rw, "missing Basic auth", l.statusCode)

		return false
	}

	if usr != l.basicAuth.username || pwd != l.basicAuth.password {
		msg := fmt.Sprintf("invalid credentials: got [username: %q, password: %q], want [username: %q, password: %q]",
			usr, pwd, l.basicAuth.username, l.basicAuth.password)
		http.Error(rw, msg, l.statusCode)

		return false
	}

	return true
}

func (l *HeaderLink) checkJSONHeaders(rw http.ResponseWriter, req *http.Request) bool {
	err := checkHeader(req, acceptHeader, []string{applicationJSONMimeType})
	if err != nil {
		http.Error(rw, err.Error(), l.statusCode)

		return false
	}

	if req.ContentLength > 0 {
		err = checkHeader(req, contentTypeHeader, []string{applicationJSONMimeType})
		if err != nil {
			http.Error(rw, err.Error(), l.statusCode)

			return false
		}
	}

	return true
}

func checkHeader(req *http.Request, k string, v []string) error {
	if !slices.Equal(req.Header[k], v) {
		return fmt.Errorf("invalid %q header value, got %q, want %q", k, req.Header[k], v)
	}

	return nil
}
