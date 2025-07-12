package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("user", "secret")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithContentTypeFromURLEncoded())
}

func TestClientAddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domain/",
			servermock.ResponseFromFixture("adddnsrecord.json").
				WithHeader("Content-Type", "application/json", "Charset=gb2312"),
			servermock.CheckQueryParameter().Strict().
				With("act", "adddnsrecord"),
			servermock.CheckForm().UsePostForm().Strict().
				With("domain", "example.com").
				With("host", "@").
				With("ttl", "60").
				With("type", "TXT").
				With("value", "txtTXTtxt").
				// With("act", "adddnsrecord").
				With("username", "user").
				WithRegexp("time", `\d+`).
				WithRegexp("token", `[a-z0-9]{32}`),
		).
		Build(t)

	record := Record{
		Domain: "example.com",
		Host:   "@",
		Type:   "TXT",
		Value:  "txtTXTtxt",
		TTL:    60,
	}

	id, err := client.AddRecord(t.Context(), record)
	require.NoError(t, err)

	assert.Equal(t, 123456, id)
}

func TestClientAddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /domain/",
			servermock.ResponseFromFixture("error.json").
				WithHeader("Content-Type", "application/json", "Charset=gb2312"),
			servermock.CheckQueryParameter().Strict().
				With("act", "adddnsrecord")).
		Build(t)

	record := Record{
		Domain: "example.com",
		Host:   "@",
		Type:   "TXT",
		Value:  "txtTXTtxt",
		TTL:    60,
	}

	_, err := client.AddRecord(t.Context(), record)
	require.Error(t, err)

	require.EqualError(t, err, "10000: username,time,token必传 (500)")
}

func TestClientDeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domain/",
			servermock.ResponseFromFixture("deldnsrecord.json").
				WithHeader("Content-Type", "application/json", "Charset=gb2312"),
			servermock.CheckQueryParameter().Strict().
				With("act", "deldnsrecord"),
			servermock.CheckForm().UsePostForm().Strict().
				With("id", "123").
				With("domain", "example.com").
				With("username", "user").
				WithRegexp("time", `\d+`).
				WithRegexp("token", `[a-z0-9]{32}`),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", 123)
	require.NoError(t, err)
}

func TestClientDeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /domain/",
			servermock.ResponseFromFixture("error.json").
				WithHeader("Content-Type", "application/json", "Charset=gb2312"),
			servermock.CheckQueryParameter().Strict().
				With("act", "deldnsrecord"),
		).
		Build(t)
	err := client.DeleteRecord(t.Context(), "example.com", 123)
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
