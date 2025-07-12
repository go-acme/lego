package servermock

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
)

// QueryParameterLink validates query parameters in HTTP requests.
// The strict flag enforces exact matches with specified query parameters.
type QueryParameterLink struct {
	values     map[string]string
	regexes    map[string]*regexp.Regexp
	strict     bool
	statusCode int
}

func CheckQueryParameter() *QueryParameterLink {
	return &QueryParameterLink{
		values:     map[string]string{},
		regexes:    map[string]*regexp.Regexp{},
		statusCode: http.StatusBadRequest,
	}
}

func (l *QueryParameterLink) Bind(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()

		if l.strict {
			if len(query) != len(l.values)+len(l.regexes) {
				msg := fmt.Sprintf("invalid query parameters, got %v, want %v", query, l.values)
				http.Error(rw, msg, l.statusCode)
				return
			}
		}

		for k, v := range l.values {
			p := query.Get(k)
			if p != v {
				msg := fmt.Sprintf("invalid %q query parameter value, got %q, want %q", k, p, v)
				http.Error(rw, msg, l.statusCode)
				return
			}
		}

		for k, exp := range l.regexes {
			value := query.Get(k)
			if !exp.MatchString(value) {
				msg := fmt.Sprintf("invalid %q query parameter value, %q doesn't match to %q", k, value, exp)
				http.Error(rw, msg, l.statusCode)
				return
			}
		}

		next.ServeHTTP(rw, req)
	})
}

func (l *QueryParameterLink) Strict() *QueryParameterLink {
	l.strict = true

	return l
}

func (l *QueryParameterLink) With(name, value string) *QueryParameterLink {
	l.values[name] = value

	return l
}

func (l *QueryParameterLink) WithRegexp(name, exp string) *QueryParameterLink {
	l.regexes[name] = regexp.MustCompile(exp)

	return l
}

func (l *QueryParameterLink) WithValues(values url.Values) *QueryParameterLink {
	for k, v := range values {
		if len(v) != 1 {
			continue
		}

		l.values[k] = v[0]
	}

	return l
}

func (l *QueryParameterLink) WithStatusCode(status int) *QueryParameterLink {
	if l.statusCode >= http.StatusContinue {
		l.statusCode = status
	}

	return l
}
