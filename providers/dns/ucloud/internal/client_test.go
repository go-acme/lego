package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/go-acme/lego/v5/internal/useragent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			credential := auth.NewCredential()
			credential.PublicKey = "pubkey"
			credential.PrivateKey = "privkey"

			cfg := ucloud.NewConfig()
			cfg.UserAgent = useragent.Get()
			cfg.BaseUrl = server.URL

			client := NewClient(&cfg, &credential)
			client.SetTransport(server.Client().Transport)

			return client, nil
		},
		servermock.CheckHeader().
			WithRegexp("U-Timestamp-Ms", `\d+`).
			WithContentTypeFromURLEncoded(),
	)
}

func TestClient_DomainDNSAdd(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture("udnrDomainDNSAdd.json"),
			servermock.CheckQueryParameter().Strict().
				With("Action", "UdnrDomainDNSAdd"),
			servermock.CheckForm().Strict().
				WithRegexp("Action", "UdnrDomainDNSAdd").
				With("Dn", "example.com").
				With("RecordName", "_acme-challenge.example.com").
				With("TTL", "600").
				With("DnsType", "TXT").
				With("Content", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("PublicKey", "pubkey").
				WithRegexp("Signature", ".+"),
		).
		Build(t)

	request := client.NewDomainDNSAddRequest()
	request.Dn = ucloud.String("example.com")
	request.RecordName = ucloud.String("_acme-challenge.example.com")
	request.DnsType = ucloud.String("TXT")
	request.Content = ucloud.String("ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY")
	request.TTL = ucloud.String("600")

	_, err := client.DomainDNSAdd(request)
	require.NoError(t, err)
}

func TestClient_DomainDNSQuery(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture("udnrDomainDNSQuery.json"),
			servermock.CheckQueryParameter().Strict().
				With("Action", "UdnrDomainDNSQuery"),
			servermock.CheckForm().Strict().
				WithRegexp("Action", "UdnrDomainDNSQuery").
				With("Dn", "example.com").
				With("PublicKey", "pubkey").
				WithRegexp("Signature", ".+"),
		).
		Build(t)

	request := client.NewDomainDNSQueryRequest()
	request.Dn = ucloud.String("example.com")

	domains, err := client.DomainDNSQuery(request)
	require.NoError(t, err)

	expected := []DomainDNSRecord{{
		Type:     "TXT",
		Name:     "_acme-challenge.example.com",
		Content:  "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		Priority: "",
		TTL:      "600",
	}}

	assert.Equal(t, expected, domains.Data)
}

func TestClient_DeleteDNSRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture("udnrDeleteDNSRecord.json"),
			servermock.CheckQueryParameter().Strict().
				With("Action", "UdnrDeleteDnsRecord"),
			servermock.CheckForm().Strict().
				WithRegexp("Action", "UdnrDeleteDnsRecord").
				With("Dn", "example.com").
				With("RecordName", "_acme-challenge.example.com").
				With("DnsType", "TXT").
				With("Content", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("PublicKey", "pubkey").
				WithRegexp("Signature", ".+"),
		).
		Build(t)

	request := client.NewDeleteDNSRecordRequest()
	request.Dn = ucloud.String("example.com")
	request.RecordName = ucloud.String("_acme-challenge.example.com")
	request.DnsType = ucloud.String("TXT")
	request.Content = ucloud.String("ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY")

	_, err := client.DeleteDNSRecord(request)
	require.NoError(t, err)
}
