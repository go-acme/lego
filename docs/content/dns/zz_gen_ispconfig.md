---
title: "ISPConfig 3"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: ispconfig
dnsprovider:
  since:    "v4.31.0"
  code:     "ispconfig"
  url:      "https://www.ispconfig.org/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ispconfig/ispconfig.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [ISPConfig 3](https://www.ispconfig.org/).


<!--more-->

- Code: `ispconfig`
- Since: v4.31.0


Here is an example bash command using the ISPConfig 3 provider:

```bash
ISPCONFIG_SERVER_URL="https://example.com:8080/remote/json.php" \
ISPCONFIG_USERNAME="xxx" \
ISPCONFIG_PASSWORD="yyy" \
lego --dns ispconfig -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ISPCONFIG_PASSWORD` | Password |
| `ISPCONFIG_SERVER_URL` | Server URL |
| `ISPCONFIG_USERNAME` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ISPCONFIG_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `ISPCONFIG_INSECURE_SKIP_VERIFY` | Whether to verify the API certificate |
| `ISPCONFIG_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `ISPCONFIG_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `ISPCONFIG_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/remoting_client/API-docs/index.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ispconfig/ispconfig.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
