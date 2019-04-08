package sakuracloud

import (
	"fmt"
	"sync"
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

func TestSacloudClient_AddTXTRecord_concurrent(t *testing.T) {
	dummyRecordCount := 10

	fakeZone := sacloud.CreateNewDNS("example.com")
	fakeZone.ID = 123456789012
	fakeClient := &fakeClient{fakeValues: map[int64]*sacloud.DNS{fakeZone.ID: fakeZone}}

	var testClients []*sacloudClient
	for i := 0; i < dummyRecordCount; i++ {
		testClients = append(testClients, &sacloudClient{client: fakeClient})
	}

	var wg sync.WaitGroup
	wg.Add(len(testClients))

	for i, testClient := range testClients {
		go func(index int, client *sacloudClient) {
			err := client.AddTXTRecord(fmt.Sprintf("test%d.example.com", index), "example.com", "dummyValue", 10)
			require.NoError(t, err)
			wg.Done()
		}(i, testClient)
	}

	wg.Wait()

	updZone, err := testClients[0].getHostedZone("example.com")
	require.NoError(t, err)
	require.NotNil(t, updZone)

	require.Len(t, updZone.Settings.DNS.ResourceRecordSets, dummyRecordCount)
}

func TestSacloudClient_CleanupTXTRecord_concurrent(t *testing.T) {
	dummyRecordCount := 10
	fakeZone := sacloud.CreateNewDNS("example.com")
	for i := 0; i < dummyRecordCount; i++ {
		fakeZone.AddRecord(fakeZone.CreateNewRecord(fmt.Sprintf("test%d", i), "TXT", "dummyValue", 10))
	}
	fakeZone.ID = 123456789012
	fakeClient := &fakeClient{fakeValues: map[int64]*sacloud.DNS{fakeZone.ID: fakeZone}}
	var testClients []*sacloudClient
	for i := 0; i < dummyRecordCount; i++ {
		testClients = append(testClients, &sacloudClient{client: fakeClient})
	}

	var wg sync.WaitGroup
	wg.Add(len(testClients))

	for i, testClient := range testClients {
		go func(index int, client *sacloudClient) {
			err := client.CleanupTXTRecord(fmt.Sprintf("test%d.example.com", index), "example.com")
			require.NoError(t, err)
			wg.Done()
		}(i, testClient)
	}

	wg.Wait()

	updZone, err := testClients[0].getHostedZone("example.com")
	require.NoError(t, err)
	require.NotNil(t, updZone)

	require.Len(t, updZone.Settings.DNS.ResourceRecordSets, 0)
}
