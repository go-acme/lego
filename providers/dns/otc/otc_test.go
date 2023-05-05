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

	s.Equal(provider.config.DomainName, "unittest1")
	s.Equal(provider.config.UserName, "unittest2")
	s.Equal(provider.config.Password, "unittest3")
	s.Equal(provider.config.ProjectName, "unittest4")
	s.Equal(provider.config.IdentityEndpoint, "unittest5")

	os.Setenv(EnvIdentityEndpoint, "")

	provider, err = NewDNSProvider()
	s.Require().NoError(err)

	s.Equal(provider.config.IdentityEndpoint, "https://iam.eu-de.otc.t-systems.com:443/v3/auth/tokens")
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
	s.NotNil(err)
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
