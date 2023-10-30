package otc

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/providers/dns/otc/internal"
	"github.com/stretchr/testify/suite"
)

type OTCSuite struct {
	suite.Suite

	mock    *internal.DNSServerMock
	envTest *tester.EnvTest
}

func (s *OTCSuite) SetupTest() {
	s.mock = internal.NewDNSServerMock(s.T())
	s.mock.HandleAuthSuccessfully()
	s.envTest = tester.NewEnvTest(
		EnvDomainName,
		EnvUserName,
		EnvPassword,
		EnvProjectName,
		EnvIdentityEndpoint,
	)
}

func (s *OTCSuite) TearDownTest() {
	s.envTest.RestoreEnv()
	s.mock.ShutdownServer()
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(OTCSuite))
}

func (s *OTCSuite) createDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.UserName = "UserName"
	config.Password = "Password"
	config.DomainName = "DomainName"
	config.ProjectName = "ProjectName"
	config.IdentityEndpoint = fmt.Sprintf("%s/v3/auth/token", s.mock.GetServerURL())

	return NewDNSProviderConfig(config)
}

func (s *OTCSuite) TestLoginEnv() {
	s.envTest.ClearEnv()

	s.envTest.Apply(map[string]string{
		EnvDomainName:       "unittest1",
		EnvUserName:         "unittest2",
		EnvPassword:         "unittest3",
		EnvProjectName:      "unittest4",
		EnvIdentityEndpoint: "unittest5",
	})

	provider, err := NewDNSProvider()
	s.Require().NoError(err)

	s.Equal("unittest1", provider.config.DomainName)
	s.Equal("unittest2", provider.config.UserName)
	s.Equal("unittest3", provider.config.Password)
	s.Equal("unittest4", provider.config.ProjectName)
	s.Equal("unittest5", provider.config.IdentityEndpoint)

	os.Setenv(EnvIdentityEndpoint, "")

	provider, err = NewDNSProvider()
	s.Require().NoError(err)

	s.Equal("https://iam.eu-de.otc.t-systems.com:443/v3/auth/tokens", provider.config.IdentityEndpoint)
}

func (s *OTCSuite) TestLoginEnvEmpty() {
	s.envTest.ClearEnv()

	_, err := NewDNSProvider()
	s.EqualError(err, "otc: some credentials information are missing: OTC_DOMAIN_NAME,OTC_USER_NAME,OTC_PASSWORD,OTC_PROJECT_NAME")
}

func (s *OTCSuite) TestDNSProvider_Present() {
	s.mock.HandleListZonesSuccessfully()
	s.mock.HandleListRecordsetsSuccessfully()

	provider, err := s.createDNSProvider()
	s.Require().NoError(err)

	err = provider.Present("example.com", "", "foobar")
	s.Require().NoError(err)
}

func (s *OTCSuite) TestDNSProvider_Present_EmptyZone() {
	s.mock.HandleListZonesEmpty()
	s.mock.HandleListRecordsetsSuccessfully()

	provider, err := s.createDNSProvider()
	s.Require().NoError(err)

	err = provider.Present("example.com", "", "foobar")
	s.Error(err)
}

func (s *OTCSuite) TestDNSProvider_CleanUp() {
	s.mock.HandleListZonesSuccessfully()
	s.mock.HandleListRecordsetsSuccessfully()
	s.mock.HandleDeleteRecordsetsSuccessfully()

	provider, err := s.createDNSProvider()
	s.Require().NoError(err)

	err = provider.CleanUp("example.com", "", "foobar")
	s.Require().NoError(err)
}

func (s *OTCSuite) TestDNSProvider_CleanUp_EmptyRecordset() {
	s.mock.HandleListZonesSuccessfully()
	s.mock.HandleListRecordsetsEmpty()

	provider, err := s.createDNSProvider()
	s.Require().NoError(err)

	err = provider.CleanUp("example.com", "", "foobar")
	s.Require().Error(err)
}
