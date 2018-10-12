package otc

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type OTCSuite struct {
	suite.Suite
	Mock *DNSServerMock
}

func (s *OTCSuite) SetupTest() {
	s.Mock = NewDNSServerMock(s.T())
	s.Mock.HandleAuthSuccessfully()
}

func (s *OTCSuite) TearDownSuite() {
	s.Mock.ShutdownServer()
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(OTCSuite))
}

func (s *OTCSuite) createDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.UserName = "test"
	config.Password = "test"
	config.DomainName = "test"
	config.ProjectName = "test"
	config.IdentityEndpoint = fmt.Sprintf("%s/v3/auth/token", s.Mock.GetServerURL())

	return NewDNSProviderConfig(config)
}

func (s *OTCSuite) TestLogin() {
	provider, err := s.createDNSProvider()
	s.Require().NoError(err)

	err = provider.loginRequest()
	s.Require().NoError(err)

	s.Equal(provider.baseURL, fmt.Sprintf("%s/v2", s.Mock.GetServerURL()))
	s.Equal(fakeOTCToken, provider.token)
}

func (s *OTCSuite) TestLoginEnv() {
	defer os.Clearenv()

	os.Setenv("OTC_DOMAIN_NAME", "unittest1")
	os.Setenv("OTC_USER_NAME", "unittest2")
	os.Setenv("OTC_PASSWORD", "unittest3")
	os.Setenv("OTC_PROJECT_NAME", "unittest4")
	os.Setenv("OTC_IDENTITY_ENDPOINT", "unittest5")

	provider, err := NewDNSProvider()
	s.Require().NoError(err)

	s.Equal(provider.config.DomainName, "unittest1")
	s.Equal(provider.config.UserName, "unittest2")
	s.Equal(provider.config.Password, "unittest3")
	s.Equal(provider.config.ProjectName, "unittest4")
	s.Equal(provider.config.IdentityEndpoint, "unittest5")

	os.Setenv("OTC_IDENTITY_ENDPOINT", "")

	provider, err = NewDNSProvider()
	s.Require().NoError(err)

	s.Equal(provider.config.IdentityEndpoint, "https://iam.eu-de.otc.t-systems.com:443/v3/auth/tokens")
}

func (s *OTCSuite) TestLoginEnvEmpty() {
	defer os.Clearenv()

	_, err := NewDNSProvider()
	s.EqualError(err, "otc: some credentials information are missing: OTC_DOMAIN_NAME,OTC_USER_NAME,OTC_PASSWORD,OTC_PROJECT_NAME")
}

func (s *OTCSuite) TestDNSProvider_Present() {
	s.Mock.HandleListZonesSuccessfully()
	s.Mock.HandleListRecordsetsSuccessfully()

	provider, err := s.createDNSProvider()
	s.Require().NoError(err)

	err = provider.Present("example.com", "", "foobar")
	s.Require().NoError(err)
}

func (s *OTCSuite) TestDNSProvider_Present_EmptyZone() {
	s.Mock.HandleListZonesEmpty()
	s.Mock.HandleListRecordsetsSuccessfully()

	provider, err := s.createDNSProvider()
	s.Require().NoError(err)

	err = provider.Present("example.com", "", "foobar")
	s.NotNil(err)
}

func (s *OTCSuite) TestDNSProvider_CleanUp() {
	s.Mock.HandleListZonesSuccessfully()
	s.Mock.HandleListRecordsetsSuccessfully()
	s.Mock.HandleDeleteRecordsetsSuccessfully()

	provider, err := s.createDNSProvider()
	s.Require().NoError(err)

	err = provider.CleanUp("example.com", "", "foobar")
	s.Require().NoError(err)
}

func (s *OTCSuite) TestDNSProvider_CleanUp_EmptyRecordset() {
	s.Mock.HandleListZonesSuccessfully()
	s.Mock.HandleListRecordsetsEmpty()

	provider, err := s.createDNSProvider()
	s.Require().NoError(err)

	err = provider.CleanUp("example.com", "", "foobar")
	s.Require().Error(err)
}
