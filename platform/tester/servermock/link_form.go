package servermock

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"slices"
)

// FormLink is a type used for validating and processing form data in HTTP requests.
// It supports strict validation, predefined values, and regex-based checks to ensure form compliance.
type FormLink struct {
	values      url.Values
	regexes     map[string]*regexp.Regexp
	strict      bool
	usePostForm bool
	statusCode  int
}

func CheckForm() *FormLink {
	return &FormLink{
		values:     url.Values{},
		regexes:    map[string]*regexp.Regexp{},
		statusCode: http.StatusBadRequest,
	}
}

func (l *FormLink) Bind(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			http.Error(rw, err.Error(), l.statusCode)
			return
		}

		form := req.Form
		if l.usePostForm {
			form = req.PostForm
		}

		if l.strict {
			if len(form) != len(l.values)+len(l.regexes) {
				msg := fmt.Sprintf("invalid query parameters, got %v, want %v", req.Form, l.values)
				http.Error(rw, msg, l.statusCode)
				return
			}
		}

		for k, v := range l.values {
			value := form[k]
			if !slices.Equal(v, value) {
				msg := fmt.Sprintf("invalid %q form value, got %q, want %q", k, value, v)
				http.Error(rw, msg, l.statusCode)
				return
			}
		}

		for k, exp := range l.regexes {
			value := form.Get(k)
			if !exp.MatchString(value) {
				msg := fmt.Sprintf("invalid %q form value, %q doesn't match to %q", k, value, exp)
				http.Error(rw, msg, l.statusCode)
				return
			}
		}

		next.ServeHTTP(rw, req)
	})
}

func (l *FormLink) Strict() *FormLink {
	l.strict = true

	return l
}

func (l *FormLink) UsePostForm() *FormLink {
	l.usePostForm = true

	return l
}

func (l *FormLink) With(name, value string) *FormLink {
	l.values.Set(name, value)

	return l
}

func (l *FormLink) WithRegexp(name, exp string) *FormLink {
	l.regexes[name] = regexp.MustCompile(exp)

	return l
}
