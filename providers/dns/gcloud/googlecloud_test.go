package gcloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/dns/v1"
)

const (
	envDomain = envNamespace + "DOMAIN"

	envServiceAccountFile = envNamespace + "SERVICE_ACCOUNT_FILE"
	envMetadataHost       = envNamespace + "METADATA_HOST"

	envGoogleApplicationCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
)

var envTest = tester.NewEnvTest(
	EnvProject,
	envServiceAccountFile,
	envGoogleApplicationCredentials,
	envMetadataHost,
	EnvServiceAccount,
	EnvImpersonateServiceAccount).
	WithDomain(envDomain).
	WithLiveTestExtra(func() bool {
		_, err := google.DefaultClient(context.Background(), dns.NdevClouddnsReadwriteScope)
		return err == nil
	})

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "invalid credentials",
			envVars: map[string]string{
				EnvProject:            "123",
				envServiceAccountFile: "",
				// as Travis run on GCE, we have to alter env
				envGoogleApplicationCredentials: "not-a-secret-file",
				envMetadataHost:                 "http://lego.wtf", // defined here to avoid the client cache.
			},
			// the error message varies according to the OS used.
			expected: "googlecloud: unable to get Google Cloud client: google: error getting credentials using GOOGLE_APPLICATION_CREDENTIALS environment variable: ",
		},
		{
			desc: "missing project",
			envVars: map[string]string{
				EnvProject:            "",
				envServiceAccountFile: "",
				// as Travis run on GCE, we have to alter env
				envMetadataHost: "http://lego.wtf",
			},
			expected: "googlecloud: project name missing",
		},
		{
			desc: "success key file",
			envVars: map[string]string{
				EnvProject:            "",
				envServiceAccountFile: "fixtures/gce_account_service_file.json",
			},
		},
		{
			desc: "success key",
			envVars: map[string]string{
				EnvProject:        "",
				EnvServiceAccount: `{"project_id": "A","type": "service_account","client_email": "foo@bar.com","private_key_id": "pki","private_key": "pk","token_uri": "/token","client_secret": "secret","client_id": "C","refresh_token": "D"}`,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		project  string
		expected string
	}{
		{
			desc:     "invalid project",
			project:  "123",
			expected: "googlecloud: unable to create Google Cloud DNS service: client is nil",
		},
		{
			desc:     "missing project",
			expected: "googlecloud: unable to create Google Cloud DNS service: client is nil",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			config := NewDefaultConfig()
			config.Project = test.project

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestPresentNoExistingRR(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	// getHostedZone: /manhattan/managedZones?alt=json&dnsName=lego.wtf.
	mux.HandleFunc("/dns/v1/projects/manhattan/managedZones", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		mzlrs := &dns.ManagedZonesListResponse{
			ManagedZones: []*dns.ManagedZone{
				{Name: "test", Visibility: "public"},
			},
		}

		err := json.NewEncoder(w).Encode(mzlrs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// findTxtRecords: /manhattan/managedZones/test/rrsets?alt=json&name=_acme-challenge.lego.wtf.&type=TXT
	mux.HandleFunc("/dns/v1/projects/manhattan/managedZones/test/rrsets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		rrslr := &dns.ResourceRecordSetsListResponse{
			Rrsets: []*dns.ResourceRecordSet{},
		}

		err := json.NewEncoder(w).Encode(rrslr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// applyChanges [Create]: /manhattan/managedZones/test/changes?alt=json
	mux.HandleFunc("/dns/v1/projects/manhattan/managedZones/test/changes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		var chgReq dns.Change
		if err := json.NewDecoder(r.Body).Decode(&chgReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		chgResp := chgReq
		chgResp.Status = changeStatusDone

		if err := json.NewEncoder(w).Encode(chgResp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	config := NewDefaultConfig()
	config.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	config.Project = "manhattan"

	p, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	p.client.BasePath = server.URL

	domain := "lego.wtf"

	err = p.Present(domain, "", "")
	require.NoError(t, err)
}

func TestPresentWithExistingRR(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	// getHostedZone: /manhattan/managedZones?alt=json&dnsName=lego.wtf.
	mux.HandleFunc("/dns/v1/projects/manhattan/managedZones", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		mzlrs := &dns.ManagedZonesListResponse{
			ManagedZones: []*dns.ManagedZone{
				{Name: "test", Visibility: "public"},
			},
		}

		err := json.NewEncoder(w).Encode(mzlrs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// findTxtRecords: /manhattan/managedZones/test/rrsets?alt=json&name=_acme-challenge.lego.wtf.&type=TXT
	mux.HandleFunc("/dns/v1/projects/manhattan/managedZones/test/rrsets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		rrslr := &dns.ResourceRecordSetsListResponse{
			Rrsets: []*dns.ResourceRecordSet{{
				Name:    "_acme-challenge.lego.wtf.",
				Rrdatas: []string{`"X7DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU"`, `"huji"`},
				Ttl:     120,
				Type:    "TXT",
			}},
		}

		err := json.NewEncoder(w).Encode(rrslr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// applyChanges [Create]: /manhattan/managedZones/test/changes?alt=json
	mux.HandleFunc("/dns/v1/projects/manhattan/managedZones/test/changes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		var chgReq dns.Change
		if err := json.NewDecoder(r.Body).Decode(&chgReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(chgReq.Additions) > 0 {
			sort.Strings(chgReq.Additions[0].Rrdatas)
		}

		var prevVal string
		for _, addition := range chgReq.Additions {
			for _, value := range addition.Rrdatas {
				if prevVal == value {
					http.Error(w, fmt.Sprintf("The resource %s already exists", value), http.StatusConflict)
					return
				}
				prevVal = value
			}
		}

		chgResp := chgReq
		chgResp.Status = changeStatusDone

		if err := json.NewEncoder(w).Encode(chgResp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	config := NewDefaultConfig()
	config.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	config.Project = "manhattan"

	p, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	p.client.BasePath = server.URL

	domain := "lego.wtf"

	err = p.Present(domain, "", "")
	require.NoError(t, err)
}

func TestPresentSkipExistingRR(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	// getHostedZone: /manhattan/managedZones?alt=json&dnsName=lego.wtf.
	mux.HandleFunc("/dns/v1/projects/manhattan/managedZones", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		mzlrs := &dns.ManagedZonesListResponse{
			ManagedZones: []*dns.ManagedZone{
				{Name: "test", Visibility: "public"},
			},
		}

		err := json.NewEncoder(w).Encode(mzlrs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// findTxtRecords: /manhattan/managedZones/test/rrsets?alt=json&name=_acme-challenge.lego.wtf.&type=TXT
	mux.HandleFunc("/dns/v1/projects/manhattan/managedZones/test/rrsets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		rrslr := &dns.ResourceRecordSetsListResponse{
			Rrsets: []*dns.ResourceRecordSet{{
				Name:    "_acme-challenge.lego.wtf.",
				Rrdatas: []string{`"47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU"`, `"X7DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU"`, `"huji"`},
				Ttl:     120,
				Type:    "TXT",
			}},
		}

		err := json.NewEncoder(w).Encode(rrslr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	config := NewDefaultConfig()
	config.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	config.Project = "manhattan"

	p, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	p.client.BasePath = server.URL

	domain := "lego.wtf"

	err = p.Present(domain, "", "")
	require.NoError(t, err)
}

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProviderCredentials(envTest.GetValue(EnvProject))
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLivePresentMultiple(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProviderCredentials(envTest.GetValue(EnvProject))
	require.NoError(t, err)

	// Check that we're able to create multiple entries
	err = provider.Present(envTest.GetDomain(), "1", "123d==")
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "2", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProviderCredentials(envTest.GetValue(EnvProject))
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
