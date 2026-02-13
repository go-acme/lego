package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient(server.URL, "user", "secret")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders(),
	)
}

func TestClient_AddDNSRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /jsonrpc10",
			servermock.ResponseFromFixture("addDNSRecord.json"),
			servermock.CheckRequestJSONBodyFromFixture("addDNSRecord-request.json"),
		).
		Build(t)

	record := DNSNode{
		Name: "_acme-challenge",
		Type: "TXT",
		Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:  300,
	}

	err := client.AddDNSRecord(t.Context(), "example.com", "viewA", record)
	require.NoError(t, err)
}

func TestClient_AddDNSRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /jsonrpc10",
			servermock.ResponseFromFixture("error.json"),
		).
		Build(t)

	record := DNSNode{
		Name: "_acme-challenge",
		Type: "TXT",
		Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:  300,
	}

	err := client.AddDNSRecord(t.Context(), "example.com", "viewA", record)
	require.EqualError(t, err, "code: Server.Keyfailure, "+
		"filename: service, line: 13, "+
		"message: Unknown keyname user, "+
		`detail: Traceback (most recent call last):   File "/usr/local/namesurfer/python/lib/python2.6/site-packages/ladon/server/dispatcher.py", line 159, in dispatch_request     result = self.call_method(method,req_dict,tc,export_dict,log_line)   File "/usr/local/namesurfer/python/lib/python2.6/site-packages/ladon/server/dispatcher.py", line 96, in call_method     result = getattr(service_class_instance,req_dict['methodname'])(*args)   File "/usr/local/namesurfer/python/lib/python2.6/site-packages/ladon/ladonizer/decorator.py", line 77, in injector     res = f(*args,**kw)   File "/usr/local/namesurfer/webui2/webui/service/service10/NSService_10.py", line 502, in addDNSRecord     key = validate_key(keyname, digest, [zonename, viewname, record.name, record.type, str(record.ttl), record.data])   File "/usr/local/namesurfer/webui2/webui/service/base/implementation.py", line 63, in validate_key     raise ApiFault('Server.Keyfailure', 'Unknown keyname %s' % keyname) ApiFault: service(13): Unknown keyname user `)
}

func TestClient_UpdateDNSHost(t *testing.T) {
	client := mockBuilder().
		Route("POST /jsonrpc10",
			servermock.ResponseFromFixture("updateDNSHost.json"),
			servermock.CheckRequestJSONBodyFromFixture("updateDNSHost-request.json"),
		).
		Build(t)

	record := DNSNode{
		Name: "_acme-challenge",
		Type: "TXT",
		Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:  300,
	}

	err := client.UpdateDNSHost(t.Context(), "example.com", "viewA", record, DNSNode{})
	require.NoError(t, err)
}

func TestClient_SearchDNSHosts(t *testing.T) {
	client := mockBuilder().
		Route("POST /jsonrpc10",
			servermock.ResponseFromFixture("searchDNSHosts.json"),
			servermock.CheckRequestJSONBodyFromFixture("searchDNSHosts-request.json"),
		).
		Build(t)

	records, err := client.SearchDNSHosts(t.Context(), "value")
	require.NoError(t, err)

	expected := []DNSNode{
		{Name: "foo", Type: "TXT", Data: "xxx", TTL: 300},
		{Name: "_acme-challenge", Type: "TXT", Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY", TTL: 300},
		{Name: "bar", Type: "A", Data: "yyy", TTL: 300},
	}

	assert.Equal(t, expected, records)
}

func TestClient_ListZones(t *testing.T) {
	client := mockBuilder().
		Route("POST /jsonrpc10",
			servermock.ResponseFromFixture("listZones.json"),
			servermock.CheckRequestJSONBodyFromFixture("listZones-request.json"),
		).
		Build(t)

	zones, err := client.ListZones(t.Context(), "value")
	require.NoError(t, err)

	expected := []DNSZone{
		{Name: "example.com", View: "viewA"},
		{Name: "example.org", View: "viewB"},
		{Name: "example.net", View: "viewC"},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_computeDigest(t *testing.T) {
	client, err := NewClient("https://test.example.com", "testkey", "testsecret")
	require.NoError(t, err)

	testCases := []struct {
		desc     string
		parts    []string
		expected string
	}{
		{
			desc:     "no parts",
			parts:    []string{},
			expected: "99b5dcdc19bfc0ce2af3fe848f4bcb6f7beb352e9599e8ba50544d86de567282",
		},
		{
			desc:     "parts",
			parts:    []string{"zone.example.com", "default"},
			expected: "94efef76383889b1ae620582a25d1c3aa9bd9ba9ac4bdccdf4aefbc3ae6e8329",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			digest := client.computeDigest(test.parts...)

			assert.Equal(t, test.expected, digest)
		})
	}
}
