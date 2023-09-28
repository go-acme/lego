package liquidweb

var testNewDNSProvider_testdata = []struct {
	desc     string
	envVars  map[string]string
	expected string
}{
	{
		desc: "minimum-success",
		envVars: map[string]string{
			EnvPrefix + EnvUsername: "blars",
			EnvPrefix + EnvPassword: "tacoman",
		},
	},
	{
		desc: "set-everything",
		envVars: map[string]string{
			EnvPrefix + EnvURL:      "https://storm.com",
			EnvPrefix + EnvUsername: "blars",
			EnvPrefix + EnvPassword: "tacoman",
			EnvPrefix + EnvZone:     "blars.com",
		},
	},
	{
		desc:     "missing credentials",
		envVars:  map[string]string{},
		expected: "liquidweb: username and password are missing, set LWAPI_USERNAME and LWAPI_PASSWORD",
	},
	{
		desc: "missing username",
		envVars: map[string]string{
			EnvPrefix + EnvPassword: "tacoman",
			EnvPrefix + EnvZone:     "blars.com",
		},
		expected: "liquidweb: username is missing, set LWAPI_USERNAME",
	},
	{
		desc: "missing password",
		envVars: map[string]string{
			EnvPrefix + EnvUsername: "blars",
			EnvPrefix + EnvZone:     "blars.com",
		},
		expected: "liquidweb: password is missing, set LWAPI_PASSWORD",
	},
}
