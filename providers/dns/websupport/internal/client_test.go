package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, method, pattern string, status int, file string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc(pattern, func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		open, err := os.Open(file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() { _ = open.Close() }()

		rw.WriteHeader(status)
		_, err = io.Copy(rw, open)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	client, err := NewClient("apiKey", "secretKey")
	require.NoError(t, err)

	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func TestClient_GetUser(t *testing.T) {
	client := setupTest(t, http.MethodGet, "/v1/user/self", http.StatusOK, "./fixtures/get-user.json")

	user, err := client.GetUser(context.Background(), "self")
	require.NoError(t, err)

	expected := &User{
		ID:                      987654321,
		Login:                   "lego@example.com",
		Active:                  true,
		CreateTime:              1675237889,
		Group:                   "users",
		Email:                   "lego@example.com",
		Phone:                   "+123456789",
		ContactPerson:           "",
		AwaitingTosConfirmation: "1",
		UserLanguage:            "sk-SK",
		Credit:                  0,
		VerifyURL:               "https://rest.websupport.sk/v1/user/verify/key/xxx",
		Billing: []Billing{{
			ID:        1099970,
			Profile:   "default",
			IsDefault: true,
			Name:      "asdsdfs",
			City:      "Žilina",
			Street:    "asddfsdfsdf",
			Zip:       "01234",
			Country:   "sk",
		}},
		Market: Market{Name: "Slovakia", Identifier: "sk", Currency: "EUR"},
	}

	assert.Equal(t, expected, user)
}

func TestClient_ListRecords(t *testing.T) {
	client := setupTest(t, http.MethodGet, "/v1/user/self/zone/example.com/record", http.StatusOK, "./fixtures/list-records.json")

	resp, err := client.ListRecords(context.Background(), "example.com")
	require.NoError(t, err)

	expected := &ListResponse{
		Items: []Record{
			{
				ID:      1,
				Type:    "A",
				Name:    "@",
				Content: "37.9.169.99",
				TTL:     600,
			}, {
				ID:      2,
				Type:    "NS",
				Name:    "@",
				Content: "ns1.scaledo.com",
				TTL:     600,
			},
		},
		Pager: Pager{Page: 1, PageSize: 0, Items: 2},
	}

	assert.Equal(t, expected, resp)
}

func TestClient_AddRecord(t *testing.T) {
	client := setupTest(t, http.MethodPost, "/v1/user/self/zone/example.com/record", http.StatusCreated, "./fixtures/add-record.json")

	record := Record{
		Type:    "TXT",
		Name:    "_acme-challenge",
		Content: "txttxttxt",
		TTL:     600,
	}

	resp, err := client.AddRecord(context.Background(), "example.com", record)
	require.NoError(t, err)

	expected := &Response{
		Status: "success",
		Item: &Record{
			ID:      4,
			Type:    "A",
			Name:    "@",
			Content: "1.2.3.4",
			TTL:     600,
			Zone: &Zone{
				ID:         1,
				Name:       "example.com",
				UpdateTime: 1381169608,
			},
		},
		Errors: json.RawMessage("[]"),
	}

	assert.Equal(t, expected, resp)
}

func TestClient_AddRecord_error_400(t *testing.T) {
	client := setupTest(t, http.MethodPost, "/v1/user/self/zone/example.com/record", http.StatusBadRequest, "./fixtures/add-record-error-400.json")

	record := Record{
		Type:    "TXT",
		Name:    "_acme-challenge",
		Content: "txttxttxt",
		TTL:     600,
	}

	resp, err := client.AddRecord(context.Background(), "example.com", record)
	require.NoError(t, err)

	assert.Equal(t, "error", resp.Status)

	expectedRecord := &Record{
		ID:      0,
		Type:    "A",
		Name:    "something bad !@#$%^&*(",
		Content: "123.456.789.123",
		TTL:     600,
		Zone: &Zone{
			ID:         1,
			Name:       "scaledo.com",
			UpdateTime: 1381169608,
		},
	}
	assert.Equal(t, expectedRecord, resp.Item)

	expected := &Errors{Name: []string{"Invalid input."}, Content: []string{"Wrong IP address format"}}
	assert.Equal(t, expected, ParseError(resp))
}

func TestClient_AddRecord_error_404(t *testing.T) {
	client := setupTest(t, http.MethodPost, "/v1/user/self/zone/example.com/record", http.StatusNotFound, "./fixtures/add-record-error-404.json")

	record := Record{
		Type:    "TXT",
		Name:    "_acme-challenge",
		Content: "txttxttxt",
		TTL:     600,
	}

	resp, err := client.AddRecord(context.Background(), "example.com", record)
	require.Error(t, err)

	assert.Nil(t, resp)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, http.MethodDelete, "/v1/user/self/zone/example.com/record/123", http.StatusOK, "./fixtures/delete-record.json")

	resp, err := client.DeleteRecord(context.Background(), "example.com", 123)
	require.NoError(t, err)

	expected := &Response{
		Status: "success",
		Item: &Record{
			ID:      1,
			Type:    "A",
			Name:    "@",
			Content: "1.2.3.4",
			TTL:     600,
			Zone: &Zone{
				ID:         1,
				Name:       "scaledo.com",
				UpdateTime: 1381316081,
			},
		},
		Errors: json.RawMessage("[]"),
	}

	assert.Equal(t, expected, resp)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := setupTest(t, http.MethodDelete, "/v1/user/self/zone/example.com/record/123", http.StatusNotFound, "./fixtures/delete-record-error-404.json")

	resp, err := client.DeleteRecord(context.Background(), "example.com", 123)
	require.Error(t, err)

	assert.Nil(t, resp)
}
