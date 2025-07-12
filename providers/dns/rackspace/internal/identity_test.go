package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIdentifier(server *httptest.Server) (*Identifier, error) {
	return NewIdentifier(server.Client(), server.URL), nil
}

func TestIdentifier_Login(t *testing.T) {
	identifier := servermock.NewBuilder[*Identifier](setupIdentifier, servermock.CheckHeader().WithJSONHeaders()).
		Route("POST /", servermock.ResponseFromFixture("tokens.json")).
		Build(t)

	identity, err := identifier.Login(t.Context(), "user", "secret")
	require.NoError(t, err)

	expected := &Identity{
		Access: Access{
			Token: Token{
				ID:                     "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				Expires:                "2014-11-24T22:05:39.115Z",
				Tenant:                 Tenant{ID: "110011", Name: "110011"},
				RAXAUTHAuthenticatedBy: []string{"APIKEY"},
			},
			ServiceCatalog: []ServiceCatalog{
				{
					Name: "cloudDatabases",
					Type: "rax:database",
					Endpoints: []Endpoint{
						{PublicURL: "https://syd.databases.api.rackspacecloud.com/v1.0/110011", Region: "SYD", TenantID: "110011", InternalURL: ""},
						{PublicURL: "https://dfw.databases.api.rackspacecloud.com/v1.0/110011", Region: "DFW", TenantID: "110011", InternalURL: ""},
						{PublicURL: "https://ord.databases.api.rackspacecloud.com/v1.0/110011", Region: "ORD", TenantID: "110011", InternalURL: ""},
						{PublicURL: "https://iad.databases.api.rackspacecloud.com/v1.0/110011", Region: "IAD", TenantID: "110011", InternalURL: ""},
						{PublicURL: "https://hkg.databases.api.rackspacecloud.com/v1.0/110011", Region: "HKG", TenantID: "110011", InternalURL: ""},
					},
				},
				{
					Name:      "cloudDNS",
					Type:      "rax:dns",
					Endpoints: []Endpoint{{PublicURL: "https://dns.api.rackspacecloud.com/v1.0/110011", Region: "", TenantID: "110011", InternalURL: ""}},
				},
				{
					Name:      "rackCDN",
					Type:      "rax:cdn",
					Endpoints: []Endpoint{{PublicURL: "https://global.cdn.api.rackspacecloud.com/v1.0/110011", Region: "", TenantID: "110011", InternalURL: "https://global.cdn.api.rackspacecloud.com/v1.0/110011"}},
				},
			},
			User: User{
				ID: "123456",
				Roles: []Role{
					{Description: "A Role that allows a user access to keystone Service methods", ID: "6", Name: "compute:default", TenantID: "110011"},
					{Description: "User Admin Role.", ID: "3", Name: "identity:user-admin", TenantID: ""},
				},
				Name:                 "jsmith",
				RAXAUTHDefaultRegion: "ORD",
			},
		},
	}

	assert.Equal(t, expected, identity)
}
