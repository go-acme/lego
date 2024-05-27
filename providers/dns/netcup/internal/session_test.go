package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockContext() context.Context {
	return context.WithValue(context.Background(), sessionIDKey, "session-id")
}

func TestClient_Login(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		raw, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if string(bytes.TrimSpace(raw)) != `{"action":"login","param":{"customernumber":"a","apikey":"b","apipassword":"c"}}` {
			http.Error(rw, fmt.Sprintf("invalid request body: %s", string(raw)), http.StatusBadRequest)
			return
		}

		response := `
		{
		    "serverrequestid": "srv-request-id",
		    "clientrequestid": "",
		    "action": "login",
		    "status": "success",
		    "statuscode": 2000,
		    "shortmessage": "Login successful",
		    "longmessage": "Session has been created successful.",
		    "responsedata": {
		        "apisessionid": "api-session-id"
		    }
		}
		`
		_, err = rw.Write([]byte(response))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	sessionID, err := client.login(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "api-session-id", sessionID)
}

func TestClient_Login_errors(t *testing.T) {
	testCases := []struct {
		desc    string
		handler func(rw http.ResponseWriter, req *http.Request)
	}{
		{
			desc: "HTTP error",
			handler: func(rw http.ResponseWriter, _ *http.Request) {
				http.Error(rw, "error message", http.StatusInternalServerError)
			},
		},
		{
			desc: "API error",
			handler: func(rw http.ResponseWriter, _ *http.Request) {
				response := `
					{
						"serverrequestid":"YxTr4EzdbJ101T211zR4yzUEMVE",
						"clientrequestid":"",
						"action":"login",
						"status":"error",
						"statuscode":4013,
						"shortmessage":"Validation Error.",
						"longmessage":"Message is empty.",
						"responsedata":""
					}`
				_, err := rw.Write([]byte(response))
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
					return
				}
			},
		},
		{
			desc: "responsedata marshaling error",
			handler: func(rw http.ResponseWriter, _ *http.Request) {
				response := `
							{
								"serverrequestid": "srv-request-id",
								"clientrequestid": "",
								"action": "login",
								"status": "success",
								"statuscode": 2000,
								"shortmessage": "Login successful",
								"longmessage": "Session has been created successful.",
								"responsedata": ""
							}`
				_, err := rw.Write([]byte(response))
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
					return
				}
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client, mux := setupTest(t)

			mux.HandleFunc("/", test.handler)

			sessionID, err := client.login(context.Background())
			assert.Error(t, err)
			assert.Equal(t, "", sessionID)
		})
	}
}

func TestClient_Logout(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		raw, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if string(bytes.TrimSpace(raw)) != `{"action":"logout","param":{"customernumber":"a","apikey":"b","apisessionid":"session-id"}}` {
			http.Error(rw, fmt.Sprintf("invalid request body: %s", string(raw)), http.StatusBadRequest)
			return
		}

		response := `
			{
				"serverrequestid": "request-id",
				"clientrequestid": "",
				"action": "logout",
				"status": "success",
				"statuscode": 2000,
				"shortmessage": "Logout successful",
				"longmessage": "Session has been terminated successful.",
				"responsedata": ""
			}`
		_, err = rw.Write([]byte(response))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	err := client.Logout(mockContext())
	require.NoError(t, err)
}

func TestClient_Logout_errors(t *testing.T) {
	testCases := []struct {
		desc    string
		handler func(rw http.ResponseWriter, req *http.Request)
	}{
		{
			desc: "HTTP error",
			handler: func(rw http.ResponseWriter, _ *http.Request) {
				http.Error(rw, "error message", http.StatusInternalServerError)
			},
		},
		{
			desc: "API error",
			handler: func(rw http.ResponseWriter, _ *http.Request) {
				response := `
					{
						"serverrequestid":"YxTr4EzdbJ101T211zR4yzUEMVE",
						"clientrequestid":"",
						"action":"logout",
						"status":"error",
						"statuscode":4013,
						"shortmessage":"Validation Error.",
						"longmessage":"Message is empty.",
						"responsedata":""
					}`
				_, err := rw.Write([]byte(response))
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
					return
				}
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client, mux := setupTest(t)

			mux.HandleFunc("/", test.handler)

			err := client.Logout(context.Background())
			require.Error(t, err)
		})
	}
}

func TestLiveClientAuth(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	// Setup
	envTest.RestoreEnv()

	client, err := NewClient(
		envTest.GetValue("NETCUP_CUSTOMER_NUMBER"),
		envTest.GetValue("NETCUP_API_KEY"),
		envTest.GetValue("NETCUP_API_PASSWORD"))
	require.NoError(t, err)

	for i := range 4 {
		t.Run("Test_"+strconv.Itoa(i+1), func(t *testing.T) {
			t.Parallel()

			ctx, err := client.CreateSessionContext(context.Background())
			require.NoError(t, err)

			err = client.Logout(ctx)
			require.NoError(t, err)
		})
	}
}
