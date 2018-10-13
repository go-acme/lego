package otc

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	Mock *DNSMock
}

func (s *TestSuite) TearDownSuite() {
	s.Mock.ShutdownServer()
}

func (s *TestSuite) SetupTest() {
	s.Mock = NewDNSMock(s.T())
	s.Mock.Setup()
	s.Mock.HandleAuthSuccessfully()
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) createDNSProvider() (*DNSProvider, error) {
	url := fmt.Sprintf("%s/v3/auth/token", s.Mock.Server.URL)

	config := NewDefaultConfig()
	config.UserName = fakeOTCUserName
	config.Password = fakeOTCPassword
	config.DomainName = fakeOTCDomainName
	config.ProjectName = fakeOTCProjectName
	config.IdentityEndpoint = url

	return NewDNSProviderConfig(config)
}

func (s *TestSuite) TestLogin() {
	provider, err := s.createDNSProvider()
	require.NoError(s.T(), err)

	err = provider.loginRequest()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), provider.baseURL, fmt.Sprintf("%s/v2", s.Mock.Server.URL))
	assert.Equal(s.T(), fakeOTCToken, provider.token)
}

func (s *TestSuite) TestLoginEnv() {
	defer os.Clearenv()

	os.Setenv("OTC_DOMAIN_NAME", "unittest1")
	os.Setenv("OTC_USER_NAME", "unittest2")
	os.Setenv("OTC_PASSWORD", "unittest3")
	os.Setenv("OTC_PROJECT_NAME", "unittest4")
	os.Setenv("OTC_IDENTITY_ENDPOINT", "unittest5")

	provider, err := NewDNSProvider()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), provider.config.DomainName, "unittest1")
	assert.Equal(s.T(), provider.config.UserName, "unittest2")
	assert.Equal(s.T(), provider.config.Password, "unittest3")
	assert.Equal(s.T(), provider.config.ProjectName, "unittest4")
	assert.Equal(s.T(), provider.config.IdentityEndpoint, "unittest5")

	os.Setenv("OTC_IDENTITY_ENDPOINT", "")

	provider, err = NewDNSProvider()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), provider.config.IdentityEndpoint, "https://iam.eu-de.otc.t-systems.com:443/v3/auth/tokens")
}

func (s *TestSuite) TestLoginEnvEmpty() {
	defer os.Clearenv()

	_, err := NewDNSProvider()
	assert.EqualError(s.T(), err, "otc: some credentials information are missing: OTC_DOMAIN_NAME,OTC_USER_NAME,OTC_PASSWORD,OTC_PROJECT_NAME")
}

func (s *TestSuite) TestDNSProvider_Present() {
	s.Mock.HandleListZonesSuccessfully()
	s.Mock.HandleListRecordsetsSuccessfully()

	provider, err := s.createDNSProvider()
	require.NoError(s.T(), err)

	err = provider.Present("example.com", "", "foobar")
	require.NoError(s.T(), err)
}

func (s *TestSuite) TestDNSProvider_Present_EmptyZone() {
	s.Mock.HandleListZonesEmpty()
	s.Mock.HandleListRecordsetsSuccessfully()

	provider, err := s.createDNSProvider()
	require.NoError(s.T(), err)

	err = provider.Present("example.com", "", "foobar")
	assert.NotNil(s.T(), err)
}

func (s *TestSuite) TestDNSProvider_CleanUp() {
	s.Mock.HandleListZonesSuccessfully()
	s.Mock.HandleListRecordsetsSuccessfully()
	s.Mock.HandleDeleteRecordsetsSuccessfully()

	provider, err := s.createDNSProvider()
	require.NoError(s.T(), err)

	err = provider.CleanUp("example.com", "", "foobar")
	require.NoError(s.T(), err)
}

func (s *TestSuite) TestDNSProvider_CleanUp_EmptyRecordset() {
	s.Mock.HandleListZonesSuccessfully()
	s.Mock.HandleListRecordsetsEmpty()

	provider, err := s.createDNSProvider()
	require.NoError(s.T(), err)

	err = provider.CleanUp("example.com", "", "foobar")
	require.Error(s.T(), err)
}
