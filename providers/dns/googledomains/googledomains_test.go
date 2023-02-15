package googledomains

// To be set by LEGO maintainers

// var envTest = tester.
// 	NewEnvTest("GOOGLE_DOMAINS_ACCESS_TOKEN").
// 	WithDomain("GOOGLE_DOMAINS_DOMAIN")

// func TestLivePresent(t *testing.T) {
// 	if !envTest.IsLiveTest() {
// 		t.Skip("skipping live test")
// 	}

// 	envTest.RestoreEnv()
// 	provider, err := NewDNSProvider()
// 	require.NoError(t, err)

// 	err = provider.Present(envTest.GetDomain(), "", "123d==")
// 	require.NoError(t, err)
// }

// func TestLiveCleanUp(t *testing.T) {
// 	if !envTest.IsLiveTest() {
// 		t.Skip("skipping live test")
// 	}

// 	envTest.RestoreEnv()
// 	provider, err := NewDNSProvider()
// 	require.NoError(t, err)

// 	time.Sleep(2 * time.Second)

// 	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
// 	require.NoError(t, err)
// }
