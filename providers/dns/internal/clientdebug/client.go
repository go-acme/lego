package clientdebug

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-acme/lego/v4/platform/config/env"
)

const replacement = "***"

type Option func(*DumpTransport)

func WithEnvKeys(keys ...string) Option {
	return func(d *DumpTransport) {
		for _, key := range keys {
			v := strings.TrimSpace(env.GetOrFile(key))
			if v == "" {
				continue
			}

			d.replacements = append(d.replacements, v, replacement)
		}
	}
}

func WithValues(values ...string) Option {
	return func(d *DumpTransport) {
		for _, value := range values {
			d.replacements = append(d.replacements, value, replacement)
		}
	}
}

func WithHeaders(keys ...string) Option {
	return func(d *DumpTransport) {
		d.regexps = append(d.regexps,
			regexp.MustCompile(fmt.Sprintf(`(?im)^(%s):.+$`, strings.Join(keys, "|"))))
	}
}

type DumpTransport struct {
	rt http.RoundTripper

	replacements []string
	replacer     *strings.Replacer

	regexps []*regexp.Regexp

	writer io.Writer
}

func NewDumpTransport(rt http.RoundTripper, opts ...Option) *DumpTransport {
	if rt == nil {
		rt = http.DefaultTransport
	}

	d := &DumpTransport{
		rt:     rt,
		writer: os.Stdout,
	}

	for _, opt := range opts {
		opt(d)
	}

	d.regexps = append(d.regexps,
		regexp.MustCompile(`(?im)^(Authorization):.+$`),
		regexp.MustCompile(`(?im)^(Token|X-Token):.+$`),
		regexp.MustCompile(`(?im)^(Auth-Token|X-Auth-Token):.+$`),
		regexp.MustCompile(`(?im)^(Api-Key|X-Api-Key|X-Api-Secret):.+$`),
	)

	if len(d.replacements) > 0 {
		d.replacer = strings.NewReplacer(d.replacements...)
	}

	return d
}

func (d *DumpTransport) RoundTrip(h *http.Request) (*http.Response, error) {
	data, _ := httputil.DumpRequestOut(h, true)

	_, _ = fmt.Fprintln(d.writer, "[HTTP Request]")
	_, _ = fmt.Fprintln(d.writer, d.redact(data))

	resp, err := d.rt.RoundTrip(h)
	if err != nil {
		return nil, err
	}

	data, _ = httputil.DumpResponse(resp, true)

	_, _ = fmt.Fprintln(d.writer, "[HTTP Response]")
	_, _ = fmt.Fprintln(d.writer, d.redact(data))

	return resp, err
}

func (d *DumpTransport) redact(content []byte) string {
	data := string(content)

	for _, r := range d.regexps {
		data = r.ReplaceAllString(data, "$1: "+replacement)
	}

	if d.replacer == nil {
		return data
	}

	return d.replacer.Replace(data)
}

// Wrap wraps an HTTP client Transport with the [DumpTransport].
func Wrap(client *http.Client, opts ...Option) *http.Client {
	val, found := os.LookupEnv("LEGO_DEBUG_DNS_API_HTTP_CLIENT")
	if !found {
		return client
	}

	if ok, _ := strconv.ParseBool(val); !ok {
		return client
	}

	client.Transport = NewDumpTransport(client.Transport, opts...)

	return client
}
