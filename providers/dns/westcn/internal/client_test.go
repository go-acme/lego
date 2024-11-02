package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/simplifiedchinese"
)

type formExpectation func(values url.Values) error

func setupTest(t *testing.T, filename string, expectations ...formExpectation) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("POST /", func(rw http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		commons := []formExpectation{
			expectValue("username", "user"),
			expectNotEmpty("time"),
			expectNotEmpty("token"),
		}

		for _, common := range commons {
			err = common(req.Form)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
				return
			}
		}

		for _, expectation := range expectations {
			err = expectation(req.Form)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
				return
			}
		}

		rw.Header().Set("Content-Type", "application/json; Charset=gb2312")

		file, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() { _ = file.Close() }()

		rw.WriteHeader(http.StatusOK)
		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	client, err := NewClient("user", "secret")
	require.NoError(t, err)

	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func expectValue(key, value string) formExpectation {
	return func(values url.Values) error {
		if values.Get(key) != value {
			return fmt.Errorf("expected %s, got %s", value, values.Get(key))
		}

		return nil
	}
}

func expectNotEmpty(key string) formExpectation {
	return func(values url.Values) error {
		if values.Get(key) == "" {
			return fmt.Errorf("%s missing", key)
		}

		return nil
	}
}

func noop() formExpectation {
	return func(_ url.Values) error {
		return nil
	}
}

func TestClientAddRecord(t *testing.T) {
	expectValue("act", "adddnsrecord")

	client := setupTest(t, "adddnsrecord.json",
		expectValue("act", "adddnsrecord"),
		expectValue("domain", "example.com"),
		expectValue("host", "@"),
		expectValue("type", "TXT"),
		expectValue("value", "txtTXTtxt"),
		expectValue("ttl", "60"),
	)

	record := Record{
		Domain: "example.com",
		Host:   "@",
		Type:   "TXT",
		Value:  "txtTXTtxt",
		TTL:    60,
	}

	id, err := client.AddRecord(context.Background(), record)
	require.NoError(t, err)

	assert.Equal(t, 123456, id)
}

func TestClientAddRecord_error(t *testing.T) {
	client := setupTest(t, "error.json", noop())

	record := Record{
		Domain: "example.com",
		Host:   "@",
		Type:   "TXT",
		Value:  "txtTXTtxt",
		TTL:    60,
	}

	_, err := client.AddRecord(context.Background(), record)
	require.Error(t, err)

	require.EqualError(t, err, "10000: username,time,token必传 (500)")
}

func TestClientDeleteRecord(t *testing.T) {
	client := setupTest(t, "deldnsrecord.json",
		expectValue("act", "deldnsrecord"),
		expectValue("domain", "example.com"),
	)

	err := client.DeleteRecord(context.Background(), "example.com", 123)
	require.NoError(t, err)
}

func TestClientDeleteRecord_error(t *testing.T) {
	client := setupTest(t, "error.json", noop())

	err := client.DeleteRecord(context.Background(), "example.com", 123)
	require.Error(t, err)

	require.EqualError(t, err, "10000: username,time,token必传 (500)")
}

func Test_convertURLValues(t *testing.T) {
	client, err := NewClient("user", "secret")
	require.NoError(t, err)

	key := "你好abc"
	value := "世界def"

	form := url.Values{}
	form.Set(key, value)

	values, err := client.convertURLValues(form)
	require.NoError(t, err)

	encoder := simplifiedchinese.GBK.NewEncoder()

	k, err := encoder.String(key)
	require.NoError(t, err)

	v, err := encoder.String(value)
	require.NoError(t, err)

	assert.Equal(t, v, values.Get(k))

	decoder := simplifiedchinese.GBK.NewDecoder()

	decValue, err := decoder.String(values.Get(k))
	require.NoError(t, err)

	assert.Equal(t, value, decValue)
}

func TestClient_sign(t *testing.T) {
	client, err := NewClient("zhangsan", "5dh232kfg!*")
	require.NoError(t, err)

	form := url.Values{}

	client.sign(form, time.UnixMilli(1554691950854))

	assert.Equal(t, "zhangsan", form.Get("username"))
	assert.Equal(t, "1554691950854", form.Get("time"))
	assert.Equal(t, "f17581fb2535b2a7ee4468eb3f96a2a9", form.Get("token"))
}
