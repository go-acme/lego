package yandexcloud

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"
const fakeToken = "ewogICJpZCI6ICJhYmNkZWZnaGlqa2xtbm9wcXJzdCIsCiAgInNlcnZpY2VfYWNjb3VudF9pZCI6ICJhYmNkZWZnaGlqa2xtbm9wcXJzdCIsCiAgImNyZWF0ZWRfYXQiOiAiMjAwMC0wMS0wMVQwMDowMDowMC4wMDAwMDAwMDBaIiwKICAia2V5X2FsZ29yaXRobSI6ICJSU0FfMjA0OCIsCiAgInB1YmxpY19rZXkiOiAiLS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS1cbk1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBa1ZGMkhqVHg0djlyR29mNU9IR09cbkdrYSs1WEpjK3B4Mktrekcwa0cySDBmdGFsOG4xTGFZMnJBUm1HcDFUMS9weDgwclIzYW1KOW1obm1CK2pINStcbnR3eFdyK3FWd1ZuSnJrbEJvemdFdGw2d1h6Qjd6TnFDM2tWNXJYWjRPbXZuNmRhS3VpY3pmZ0xMN04veVlRemtcblNLUllPQ3lnQmJQb3hWR1M1MFpMVmRDV1d0ejFpRmJObUVsbnNNNEtRam54V0JWUkR3UjJINU9JVTg0Tm9uVXpcbk5jSERrVkJYL2Q4cGtTZzdpQjROeUQxQXF2SnRGMXBTMDNOUW0zMm42OWJzZlJzSnhycVI2TEsvYXFsMzc5cmtcbmhnQTdTeXpNTEpjTGNrS3VnK0tmVENwa3Ryd3ppMkFwcFVQRDdrZUtKaWxPZmhTckNHUWdsTXI2UTNhbzAzU1pcbmNRSURBUUFCXG4tLS0tLUVORCBQVUJMSUMgS0VZLS0tLS0iLAogICJwcml2YXRlX2tleSI6ICItLS0tLUJFR0lOIFBSSVZBVEUgS0VZLS0tLS1cbk1JSUV2Z0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktnd2dnU2tBZ0VBQW9JQkFRQ1JVWFllTlBIaS8yc2FcbmgvazRjWTRhUnI3bGNsejZuSFlxVE1iU1FiWWZSKzFxWHlmVXRwamFzQkdZYW5WUFgrbkh6U3RIZHFZbjJhR2VcbllINk1mbjYzREZhdjZwWEJXY211U1VHak9BUzJYckJmTUh2TTJvTGVSWG10ZG5nNmErZnAxb3E2SnpOK0FzdnNcbjMvSmhET1JJcEZnNExLQUZzK2pGVVpMblJrdFYwSlphM1BXSVZzMllTV2V3emdwQ09mRllGVkVQQkhZZms0aFRcbnpnMmlkVE0xd2NPUlVGZjkzeW1SS0R1SUhnM0lQVUNxOG0wWFdsTFRjMUNiZmFmcjF1eDlHd25HdXBIb3NyOXFcbnFYZnYydVNHQUR0TExNd3Nsd3R5UXE2RDRwOU1LbVMydkRPTFlDbWxROFB1UjRvbUtVNStGS3NJWkNDVXl2cERcbmRxalRkSmx4QWdNQkFBRUNnZ0VBT3pHN3M4Sk5aZkkxWnJGTXk3azE4VzR3QkxiNU9QelRCWmdReFVVUE10N1Jcbnp5ckR4dG82bVpwdkVHOE5LakFmd3N2SWZXdlBjeHdyd1ovODdLMzZZQVllcWJvZEZvM0VvY0lsZ3A4bkRFSzJcbkJaQnlYWmdGQnhXMTR2c0hMb1VXQ3lMaGo4SzRMdlJrclREc1FxeEZzWEdBbmlGUGJnTkRKbDE4UWNsWWxyT3Jcbm5uOVpGN1cwdDJkMGpudXp3QjlrOEwxOFJxUllXb3ZDQWpuRkNTMHRYNXVRS3RqU1lEMEpSRzdDaUtxZDRydXZcbnRKMUdvNGJvK3JSY2FFYkZnRHlmOEJFVmE2dDlWSlgxTVZqTDJ4bTB0b1FVanRBK1pUdUFBZzRoQ2liRW9ydThcbllvNTUrUjY1SEhJOUI4blp4ZnAwa0VWeXpBaFFXb3Y5MUpiSHpoUmlBUUtCZ1FETTh5dUo0dERBUTUzUkRtREZcblg1ZXIyRjlUZUpvMkFSaUZCMkMrNGg5STg4akMxTEozS2dkMTYxTU8xbVkzU1ZmTk1IWFpjMHRwUkRyKzV4ZG5cblVOS3VWOEFTK084MEZhbjVlSlgyNDViSmlYcjdRNzN0VjFQalZ3Sm1Ya01UK0dhSVRxS3NHeU9acDFtczYxRWRcblAvWWFEZlM3YXoxS2VJR0tXbWtPNXhEYzJRS0JnUUMxZzlHNHdUckFhYVo4dVhCa205ODJPeTQ3aU1EeTRJZ1dcbmE0bUx5ZWRodkJoT0ZOU0d3TktmdzZ6QlgrUFBUMUZLTTl4SlgxZzFrYk5OaEgrVy95L1F4L3VOejdRY3NTdlFcbnNVVlJ3UFJtVWFyUHNJdURHdnFJajdrbjdIalFncUovaFRsbU9YUjNmVHJ2R1pxOE9ZeWhnRjZCcW93UEZTLzJcbnhWWU9MWHNpV1FLQmdRQ3BteGROelpsSmN1dDRaVGlxUGZpTGFzMUFpNDY2NEY5RlA1ek5ldDIvQnBmK3UveFFcbjUwUXpUcUoycGZFREVid0tmMjhYbS9VdFVSeXRjOXFIVW5oM2RRRHI4bndxRXorTnh6LzdoODV5VEVhdEJ4dDJcbi9ZemJsMWJTRm5IV1pmdWNFODlGTkZSYXhRWk9OcEx5N01xaU55aHZyVWlVaDNOVVpvdUluS24weVFLQmdFQXZcbkdvdWdHQ3hOcjRkTzgwVkFNTSsyWVlTL3VLcXBaclcyMU81UE9MaEFrTCtiY2dNc1Q4NGFuUTNMNEh3LzZkaTVcbk9kM2dEd3J5T0ZyaXpWTVJiVkVBUmgxQklzazZoT25JcFdCaFFJcWx1aWF5b01KOVdiWE1USWFuZ1prSmVIaHJcbkhYN2VOaWJDYTRKOHBWQ0ZjUXJ5bjNodVhCUkJRN0tZMlBNdWRlb1JBb0dCQUoxdmRCUVN1YWkzUklmeWo4WXJcbjRBcnRDVTFUNWJpY3AxMyttSk9EU2VSaEhNbmxLa21JNjR2d3JXNVBPRlhXeUpLUFlMa3VEazliRVlPeU5CT0FcbkJUc1V5YUpwM2p4Lzk0Mm9Fd1VSYzRUYjlhejdDcUVIYUNyV0hWSENqMUNqQ0VYL0ZzUmZkK3dZeXVHTHd3bHlcbndkcHFCV0JsNWlINzR0UkQ2YytyZ3VtYVxuLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLSIKfQ=="

var envTest = tester.NewEnvTest(EnvIamToken, EnvFolderId).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvIamToken: fakeToken,
				EnvFolderId: "folder_id",
			},
		},
		{
			desc: "missing iam token",
			envVars: map[string]string{
				EnvFolderId: "folder_id",
			},
			expected: "yandexcloud: some credentials information are missing: YANDEX_CLOUD_IAM_TOKEN",
		},
		{
			desc: "missing folder_id",
			envVars: map[string]string{
				EnvIamToken: fakeToken,
			},
			expected: "yandexcloud: some credentials information are missing: YANDEX_CLOUD_FOLDER_ID",
		},
		{
			desc: "malformed token (not base64)",
			envVars: map[string]string{
				EnvIamToken: "not_base64",
				EnvFolderId: "folder_id",
			},
			expected: "yandexcloud: iam token is malformed: illegal base64 data at input byte 3",
		},
		{
			desc: "malformed token (invalid json in bas64)",
			envVars: map[string]string{
				EnvIamToken: "aW52YWxpZCBqc29u",
				EnvFolderId: "folder_id",
			},
			expected: "yandexcloud: iam token is malformed: invalid character 'i' looking for beginning of value",
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
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		config   *Config
		expected string
	}{
		{
			desc: "success",
			config: &Config{
				IamToken: fakeToken,
				FolderID: "folder_id",
			},
		},
		{
			desc:     "nil config",
			config:   nil,
			expected: "yandexcloud: the configuration of the DNS provider is nil",
		},
		{
			desc: "missing token",
			config: &Config{
				FolderID: "folder_id",
			},
			expected: "yandexcloud: some credentials information are missing iam token",
		}, {
			desc: "missing token",
			config: &Config{
				IamToken: fakeToken,
			},
			expected: "yandexcloud: some credentials information are missing folder id",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			p, err := NewDNSProviderConfig(test.config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
