package dmapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockContext(sessionID string) context.Context {
	if sessionID == "" {
		sessionID = "xxx"
	}

	return context.WithValue(context.Background(), sessionIDKey, sessionID)
}

func TestClient_login_apikey(t *testing.T) {
	testCases := []struct {
		desc               string
		apiKey             string
		expectedError      bool
		expectedStatusCode int
		expectedAuthSid    string
	}{
		{
			desc:               "correct key",
			apiKey:             correctAPIKey,
			expectedStatusCode: 0,
			expectedAuthSid:    correctAPIKey,
		},
		{
			desc:               "incorrect key",
			apiKey:             incorrectAPIKey,
			expectedStatusCode: 2200,
			expectedError:      true,
		},
		{
			desc:               "server error",
			apiKey:             serverErrorAPIKey,
			expectedStatusCode: -500,
			expectedError:      true,
		},
		{
			desc:               "non-ok status code",
			apiKey:             "333",
			expectedStatusCode: 2202,
			expectedError:      true,
		},
	}

	mux, serverURL := setupTest(t)

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)

		switch r.FormValue("api-key") {
		case correctAPIKey:
			_, _ = io.WriteString(w, "Status-Code: 0\nStatus-Text: OK\nAuth-Sid: 123\n\ncom\nnet")
		case incorrectAPIKey:
			_, _ = io.WriteString(w, "Status-Code: 2200\nStatus-Text: Authentication error")
		case serverErrorAPIKey:
			http.NotFound(w, r)
		default:
			_, _ = io.WriteString(w, "Status-Code: 2202\nStatus-Text: OK\n\ncom\nnet")
		}
	})

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := NewClient(AuthInfo{APIKey: test.apiKey})
			client.BaseURL = serverURL

			response, err := client.login(context.Background())
			if test.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, test.expectedStatusCode, response.StatusCode)
				assert.Equal(t, test.expectedAuthSid, response.AuthSid)
			}
		})
	}
}

func TestClient_login_username(t *testing.T) {
	testCases := []struct {
		desc               string
		username           string
		password           string
		expectedError      bool
		expectedStatusCode int
		expectedAuthSid    string
	}{
		{
			desc:               "correct username and password",
			username:           correctUsername,
			password:           "go-acme",
			expectedError:      false,
			expectedStatusCode: 0,
			expectedAuthSid:    correctAPIKey,
		},
		{
			desc:               "incorrect username",
			username:           incorrectUsername,
			password:           "go-acme",
			expectedStatusCode: 2200,
			expectedError:      true,
		},
		{
			desc:               "server error",
			username:           serverErrorUsername,
			password:           "go-acme",
			expectedStatusCode: -500,
			expectedError:      true,
		},
		{
			desc:               "non-ok status code",
			username:           "random",
			password:           "go-acme",
			expectedStatusCode: 2202,
			expectedError:      true,
		},
	}

	mux, serverURL := setupTest(t)

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)

		switch r.FormValue("username") {
		case correctUsername:
			_, _ = io.WriteString(w, "Status-Code: 0\nStatus-Text: OK\nAuth-Sid: 123\n\ncom\nnet")
		case incorrectUsername:
			_, _ = io.WriteString(w, "Status-Code: 2200\nStatus-Text: Authentication error")
		case serverErrorUsername:
			http.NotFound(w, r)
		default:
			_, _ = io.WriteString(w, "Status-Code: 2202\nStatus-Text: OK\n\ncom\nnet")
		}
	})

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := NewClient(AuthInfo{Username: test.username, Password: test.password})
			client.BaseURL = serverURL

			response, err := client.login(context.Background())
			if test.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, test.expectedStatusCode, response.StatusCode)
				assert.Equal(t, test.expectedAuthSid, response.AuthSid)
			}
		})
	}
}

func TestClient_logout(t *testing.T) {
	testCases := []struct {
		desc               string
		authSid            string
		expectedError      bool
		expectedStatusCode int
	}{
		{
			desc:               "correct auth-sid",
			authSid:            correctAPIKey,
			expectedStatusCode: 0,
		},
		{
			desc:               "incorrect auth-sid",
			authSid:            incorrectAPIKey,
			expectedStatusCode: 2200,
		},
		{
			desc:          "already logged out",
			authSid:       "",
			expectedError: true,
		},
		{
			desc:          "server error",
			authSid:       serverErrorAPIKey,
			expectedError: true,
		},
	}

	mux, serverURL := setupTest(t)

	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)

		switch r.FormValue("auth-sid") {
		case correctAPIKey:
			_, _ = io.WriteString(w, "Status-Code: 0\nStatus-Text: OK\n")
		case incorrectAPIKey:
			_, _ = io.WriteString(w, "Status-Code: 2200\nStatus-Text: Authentication error")
		default:
			http.NotFound(w, r)
		}
	})

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := NewClient(AuthInfo{APIKey: "12345"})
			client.BaseURL = serverURL
			client.token = &Token{SessionID: test.authSid}

			response, err := client.Logout(mockContext(test.authSid))
			if test.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, test.expectedStatusCode, response.StatusCode)
			}
		})
	}
}

func TestClient_CreateAuthenticatedContext(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	id := atomic.Int32{}
	id.Add(100)

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)

		switch r.FormValue("username") {
		case correctUsername:
			_, _ = fmt.Fprintf(w, "Status-Code: 0\nStatus-Text: OK\nAuth-Sid: %d\n\ncom\nnet", id.Load())
			id.Add(100)

		default:
			_, _ = io.WriteString(w, "Status-Code: 2200\nStatus-Text: Authentication error")
		}
	})

	client := NewClient(AuthInfo{Username: correctUsername, Password: "secret"})
	client.HTTPClient = server.Client()
	client.BaseURL = server.URL

	ctx, err := client.CreateAuthenticatedContext(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "100", getSessionID(ctx))

	// the token is not expired then we use the "cache".
	client.muToken.Lock()
	client.token.SessionID = "cache"
	client.muToken.Unlock()

	ctx, err = client.CreateAuthenticatedContext(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "cache", getSessionID(ctx))

	// force the expiration of the token
	client.muToken.Lock()
	client.token.ExpireAt = time.Now().UTC().Add(-1 * time.Hour)
	client.muToken.Unlock()

	ctx, err = client.CreateAuthenticatedContext(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "200", getSessionID(ctx))
}
