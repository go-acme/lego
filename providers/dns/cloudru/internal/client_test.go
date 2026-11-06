package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder(hc *http.Client) *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient(hc)
			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("Bearer "+
				"eyJhbGciOiJSUzI1NiIsImtpZCI6ImEyMjM3ZDhhLWQ0ZDQtNDA5Yi04ZTMxLWM3NGJhYTZhM2NjYiIsInR5cCI6IkpXVCJ9."+
				"eyJhdWQiOlsiaWFtIl0sImF1dGhfdGltZSI6MTY5MDA1Mzg1MiwiYXpwIjoiOWE0M2I0OTM1ZDRhMDc5NmRkYjE0Mjk0NjUxYjk2NzciLCJlbWFpbCI6ImxlZ29AOTlmN2I5NzItZmZlYS00OTkyLTgyN2EtY2M4MDYzOTg1MmNhLmlhbS5zYmVyY2xvdWQucnUiLCJleHAiOjE2OTAwNTc0NTIsImlhdCI6MTY5MDA1Mzg1MiwiaXNzIjoiaHR0cHM6Ly9hdXRoLmlhbS5zYmVyY2xvdWQucnUvYXV0aC9zeXN0ZW0iLCJqdGkiOiJlYzk0ZWJhNC03NzU2LTRjNjQtYmNmMC0zMzYxODIwNWM5ODkiLCJuYmYiOjE2OTAwNTM4NTIsIm5vbmNlIjoiIiwicmVhbG1fYWNjZXNzIjpudWxsLCJyZXNvdXJjZV9hY2Nlc3MiOnt9LCJzY29wZSI6Im9wZW5pZCBwcm9maWxlIGVtYWlsIHJvbGVzIiwic3ViIjoiZmVhMzRmYTUtZmE5ZS00OTdkLTk5MDAtNjdhYWJjOWNkZWJjIiwic3ViX2lkIjoiZmVhMzRmYTUtZmE5ZS00OTdkLTk5MDAtNjdhYWJjOWNkZWJjIiwic3ViX3R5cGUiOiJzZXJ2aWNlX2FjY291bnQiLCJ0eXAiOiJCZWFyZXIifQ."+
				"hhPr-Xr_NbyRwrqGoqeepthWfpfmD47RjzHUwo2lVPkeMiL8AMWzDPRxs-8gns9eTSHZCoAH0RjyrBnTaOrztInM72h8_rIIFr0MMPIIqrUkp2id_alya9eoiSWg_69PzNZ2CKWJDylL8o4Vi9_cSBYp-6H1xNcOAvO4a9xkNCoGGiogjHWNFq64qnS_P6fYY-pl9leuprCeq1GAKPODevHwzmc4gkEZIj_15SUh_ofJRJICgyLmkELQ8a0wDGYmZcdNKiGQDpd7rHaGrOvO1k8IJHfgs5aCMyuHXybTg6AMlodpYs8MBdk6K_VFY-cxSRB8ocq_Q7Hgt9qaRADg2Q",
			),
	)
}

func TestClient_GetZones(t *testing.T) {
	client := mockBuilder(mockOAuthClient(t)).
		Route("GET /zones",
			servermock.ResponseFromFixture("zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("parentId", "xxx"),
		).
		Build(t)

	zones, err := client.GetZones(t.Context(), "xxx")
	require.NoError(t, err)

	expected := []Zone{
		{
			ID:        "59556fcd-95ff-451f-b49b-9732f21f944a",
			ParentID:  "2d7b6194-2b83-4f71-86fd-a1e727e347b2",
			Name:      "example.com",
			Valid:     true,
			Delegated: true,
			CreatedAt: time.Date(2023, 7, 23, 8, 12, 41, 0, time.UTC),
			UpdatedAt: time.Date(2023, 7, 24, 5, 50, 28, 0, time.UTC),
		},
	}
	assert.Equal(t, expected, zones)
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder(mockOAuthClient(t)).
		Route("GET /zones/zzz/records",
			servermock.ResponseFromFixture("records.json"),
		).
		Build(t)

	records, err := client.GetRecords(t.Context(), "zzz")
	require.NoError(t, err)

	expected := []Record{
		{
			ZoneID: "59556fcd-95ff-451f-b49b-9732f21f944a",
			Name:   "example.com.",
			Type:   "SOA",
			Values: []string{
				"cdns-ns01.sbercloud.ru. mail.sbercloud.ru 1 120 3600 604800 3600",
			},
			TTL:     "3600",
			Enables: true,
		},
		{
			ZoneID: "59556fcd-95ff-451f-b49b-9732f21f944a",
			Name:   "example.com.",
			Type:   "NS",
			Values: []string{
				"cdns-ns01.sbercloud.ru.",
				"cdns-ns02.sbercloud.ru.",
			},
			TTL:     "3600",
			Enables: true,
		},
		{
			ZoneID: "59556fcd-95ff-451f-b49b-9732f21f944a",
			Name:   "www.example.com.",
			Type:   "A",
			Values: []string{
				"8.8.8.8",
			},
			TTL:     "3600",
			Enables: true,
		},
	}
	assert.Equal(t, expected, records)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder(mockOAuthClient(t)).
		Route("POST /zones/zzz/records",
			servermock.ResponseFromFixture("record.json"),
			servermock.CheckRequestJSONBodyFromFixture("record-request.json"),
		).
		Build(t)

	recordReq := Record{
		Name:   "_acme-challenge.example.com.",
		Type:   "TXT",
		Values: []string{"ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"},
		TTL:    "120",
	}

	record, err := client.CreateRecord(t.Context(), "zzz", recordReq)
	require.NoError(t, err)

	expected := &Record{
		ZoneID: "59556fcd-95ff-451f-b49b-9732f21f944a",
		Name:   "_acme-challenge.example.com.",
		Type:   "TXT",
		Values: []string{
			"ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		},
		TTL:     "120",
		Enables: true,
	}
	assert.Equal(t, expected, record)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder(mockOAuthClient(t)).
		Route("DELETE /zones/zzz/records/example.com/TXT",
			servermock.ResponseFromFixture("record.json"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "zzz", "example.com", "TXT")
	require.NoError(t, err)
}
