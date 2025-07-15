package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient(OAuthStaticAccessToken(server.Client(), "secret"))
			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("Bearer secret"))
}

func TestClient_GetDomainIDByName(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/domains",
			servermock.JSONEncode(DomainListingResponse{
				Embedded: EmbeddedDomainList{Domains: []*Domain{
					{ID: 1, Name: "test.com"},
					{ID: 2, Name: "test.org"},
				}},
			})).
		Build(t)

	id, err := client.GetDomainIDByName(t.Context(), "test.com")
	require.NoError(t, err)

	assert.Equal(t, 1, id)
}

func TestClient_CheckNameservers(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/domains/1/nameservers",
			servermock.JSONEncode(NameserverResponse{
				Nameservers: []*Nameserver{
					{Name: ns1},
					{Name: ns2},
					// {Name: "ns.fake.de"},
				},
			})).
		Build(t)

	err := client.CheckNameservers(t.Context(), 1)
	require.NoError(t, err)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /v1/domains/1/nameservers/records", nil,
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json")).
		Build(t)

	record := &Record{
		Name:  "test.com",
		TTL:   300,
		Type:  "TXT",
		Value: "value",
	}

	err := client.CreateRecord(t.Context(), 1, record)
	require.NoError(t, err)
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	domainName := "lego.test"
	recordValue := "test"

	client := mockBuilder().
		Route("GET /v1/domains/",
			servermock.JSONEncode(DomainResponse{
				ID:   1,
				Name: domainName,
			})).
		Route("GET /v1/domains/1/nameservers",
			servermock.JSONEncode(NameserverResponse{
				Nameservers: []*Nameserver{{Name: ns1}, {Name: ns2}},
			})).
		Route("GET /v1/domains/1/nameservers/records",
			servermock.JSONEncode(RecordListingResponse{
				Embedded: EmbeddedRecordList{
					Records: []*Record{
						{
							Name:  "_acme-challenge",
							Value: recordValue,
							Type:  "TXT",
						},
						{
							Name:  "_acme-challenge",
							Value: recordValue,
							Type:  "A",
						},
						{
							Name:  "foobar",
							Value: recordValue,
							Type:  "TXT",
						},
					},
				},
			})).
		Route("PUT /v1/domains/1/nameservers/records", nil,
			servermock.CheckRequestJSONBodyFromFixture("delete_txt_record-request.json")).
		Build(t)

	info := dns01.GetChallengeInfo(domainName, "abc")
	err := client.DeleteTXTRecord(t.Context(), 1, info.EffectiveFQDN, recordValue)
	require.NoError(t, err)
}
