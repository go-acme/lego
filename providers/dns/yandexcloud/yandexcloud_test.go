package yandexcloud

import (
	"encoding/base64"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

const fakeIAMToken = `
{
  "id": "abcdefghijklmnopqrst",
  "service_account_id": "abcdefghijklmnopqrst",
  "created_at": "2000-01-01T00:00:00.000000000Z",
  "key_algorithm": "RSA_2048",
  "public_key": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAkVF2HjTx4v9rGof5OHGO\nGka+5XJc+px2KkzG0kG2H0ftal8n1LaY2rARmGp1T1/px80rR3amJ9mhnmB+jH5+\ntwxWr+qVwVnJrklBozgEtl6wXzB7zNqC3kV5rXZ4Omvn6daKuiczfgLL7N/yYQzk\nSKRYOCygBbPoxVGS50ZLVdCWWtz1iFbNmElnsM4KQjnxWBVRDwR2H5OIU84NonUz\nNcHDkVBX/d8pkSg7iB4NyD1AqvJtF1pS03NQm32n69bsfRsJxrqR6LK/aql379rk\nhgA7SyzMLJcLckKug+KfTCpktrwzi2AppUPD7keKJilOfhSrCGQglMr6Q3ao03SZ\ncQIDAQAB\n-----END PUBLIC KEY-----",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCRUXYeNPHi/2sa\nh/k4cY4aRr7lclz6nHYqTMbSQbYfR+1qXyfUtpjasBGYanVPX+nHzStHdqYn2aGe\nYH6Mfn63DFav6pXBWcmuSUGjOAS2XrBfMHvM2oLeRXmtdng6a+fp1oq6JzN+Asvs\n3/JhDORIpFg4LKAFs+jFUZLnRktV0JZa3PWIVs2YSWewzgpCOfFYFVEPBHYfk4hT\nzg2idTM1wcORUFf93ymRKDuIHg3IPUCq8m0XWlLTc1Cbfafr1ux9GwnGupHosr9q\nqXfv2uSGADtLLMwslwtyQq6D4p9MKmS2vDOLYCmlQ8PuR4omKU5+FKsIZCCUyvpD\ndqjTdJlxAgMBAAECggEAOzG7s8JNZfI1ZrFMy7k18W4wBLb5OPzTBZgQxUUPMt7R\nzyrDxto6mZpvEG8NKjAfwsvIfWvPcxwrwZ/87K36YAYeqbodFo3EocIlgp8nDEK2\nBZByXZgFBxW14vsHLoUWCyLhj8K4LvRkrTDsQqxFsXGAniFPbgNDJl18QclYlrOr\nnn9ZF7W0t2d0jnuzwB9k8L18RqRYWovCAjnFCS0tX5uQKtjSYD0JRG7CiKqd4ruv\ntJ1Go4bo+rRcaEbFgDyf8BEVa6t9VJX1MVjL2xm0toQUjtA+ZTuAAg4hCibEoru8\nYo55+R65HHI9B8nZxfp0kEVyzAhQWov91JbHzhRiAQKBgQDM8yuJ4tDAQ53RDmDF\nX5er2F9TeJo2ARiFB2C+4h9I88jC1LJ3Kgd161MO1mY3SVfNMHXZc0tpRDr+5xdn\nUNKuV8AS+O80Fan5eJX245bJiXr7Q73tV1PjVwJmXkMT+GaITqKsGyOZp1ms61Ed\nP/YaDfS7az1KeIGKWmkO5xDc2QKBgQC1g9G4wTrAaaZ8uXBkm982Oy47iMDy4IgW\na4mLyedhvBhOFNSGwNKfw6zBX+PPT1FKM9xJX1g1kbNNhH+W/y/Qx/uNz7QcsSvQ\nsUVRwPRmUarPsIuDGvqIj7kn7HjQgqJ/hTlmOXR3fTrvGZq8OYyhgF6BqowPFS/2\nxVYOLXsiWQKBgQCpmxdNzZlJcut4ZTiqPfiLas1Ai4664F9FP5zNet2/Bpf+u/xQ\n50QzTqJ2pfEDEbwKf28Xm/UtURytc9qHUnh3dQDr8nwqEz+Nxz/7h85yTEatBxt2\n/Yzbl1bSFnHWZfucE89FNFRaxQZONpLy7MqiNyhvrUiUh3NUZouInKn0yQKBgEAv\nGougGCxNr4dO80VAMM+2YYS/uKqpZrW21O5POLhAkL+bcgMsT84anQ3L4Hw/6di5\nOd3gDwryOFrizVMRbVEARh1BIsk6hOnIpWBhQIqluiayoMJ9WbXMTIangZkJeHhr\nHX7eNibCa4J8pVCFcQryn3huXBRBQ7KY2PMudeoRAoGBAJ1vdBQSuai3RIfyj8Yr\n4ArtCU1T5bicp13+mJODSeRhHMnlKkmI64vwrW5POFXWyJKPYLkuDk9bEYOyNBOA\nBTsUyaJp3jx/942oEwURc4Tb9az7CqEHaCrWHVHCj1CjCEX/FsRfd+wYyuGLwwly\nwdpqBWBl5iH74tRD6c+rguma\n-----END PRIVATE KEY-----"
}
`

var envTest = tester.NewEnvTest(EnvIamToken, EnvFolderID).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvIamToken: base64.StdEncoding.EncodeToString([]byte(fakeIAMToken)),
				EnvFolderID: "folder_id",
			},
		},
		{
			desc: "missing iam token",
			envVars: map[string]string{
				EnvFolderID: "folder_id",
			},
			expected: "yandexcloud: some credentials information are missing: YANDEX_CLOUD_IAM_TOKEN",
		},
		{
			desc: "missing folder_id",
			envVars: map[string]string{
				EnvIamToken: base64.StdEncoding.EncodeToString([]byte(fakeIAMToken)),
			},
			expected: "yandexcloud: some credentials information are missing: YANDEX_CLOUD_FOLDER_ID",
		},
		{
			desc: "malformed token (not base64)",
			envVars: map[string]string{
				EnvIamToken: fakeIAMToken,
				EnvFolderID: "folder_id",
			},
			expected: "yandexcloud: iam token is malformed: illegal base64 data at input byte 1",
		},
		{
			desc: "malformed token (invalid json in bas64)",
			envVars: map[string]string{
				EnvIamToken: "aW52YWxpZCBqc29u",
				EnvFolderID: "folder_id",
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
				IamToken: base64.StdEncoding.EncodeToString([]byte(fakeIAMToken)),
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
			expected: "yandexcloud: some credentials information are missing IAM token",
		},
		{
			desc: "missing folder id",
			config: &Config{
				IamToken: base64.StdEncoding.EncodeToString([]byte(fakeIAMToken)),
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
