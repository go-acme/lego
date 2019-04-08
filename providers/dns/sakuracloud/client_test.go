package sakuracloud

import (
	"testing"

	"github.com/sacloud/libsacloud/api"
	"github.com/sacloud/libsacloud/sacloud"
	"github.com/stretchr/testify/require"
)

type fakeClient struct {
	fakeValues map[int64]*sacloud.DNS
}

func (f *fakeClient) update(id int64, value *sacloud.DNS) (*sacloud.DNS, error) {
	f.fakeValues[id] = value
	return value, nil
}

func (f *fakeClient) find(zoneName string) (*api.SearchDNSResponse, error) {

	res := &api.SearchDNSResponse{}
	for _, zone := range f.fakeValues {
		res.CommonServiceDNSItems = append(res.CommonServiceDNSItems, *zone)
	}
	res.Total = len(f.fakeValues)
	res.Count = len(f.fakeValues)
	return res, nil
}

func TestSacloudClient_AddTXTRecord(t *testing.T) {
	fakeZone := sacloud.CreateNewDNS("example.com")
	fakeZone.ID = 123456789012
	fakeClient := &fakeClient{fakeValues: map[int64]*sacloud.DNS{fakeZone.ID: fakeZone}}
	testClient := &sacloudClient{client: fakeClient}

	err := testClient.AddTXTRecord("test.example.com", "example.com", "dummyValue", 10)
	require.NoError(t, err)

	updZone, err := testClient.getHostedZone("example.com")
	require.NoError(t, err)
	require.NotNil(t, updZone)

	require.Len(t, updZone.Settings.DNS.ResourceRecordSets, 1)
}

func TestSacloudClient_CleanupTXTRecord(t *testing.T) {
	fakeZone := sacloud.CreateNewDNS("example.com")
	fakeZone.AddRecord(fakeZone.CreateNewRecord("test", "TXT", "dummyValue", 10))
	fakeZone.ID = 123456789012
	fakeClient := &fakeClient{fakeValues: map[int64]*sacloud.DNS{fakeZone.ID: fakeZone}}
	testClient := &sacloudClient{client: fakeClient}

	err := testClient.CleanupTXTRecord("test.example.com", "example.com")
	require.NoError(t, err)

	updZone, err := testClient.getHostedZone("example.com")
	require.NoError(t, err)
	require.NotNil(t, updZone)

	require.Len(t, updZone.Settings.DNS.ResourceRecordSets, 0)
}
