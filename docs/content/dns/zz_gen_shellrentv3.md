---
title: "Shellrent v3"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: shellrentv3
dnsprovider:
  since:    "v5.8.0"
  code:     "shellrentv3"
  url:      "https://www.shellrent.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/shellrentv3/shellrentv3.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Shellrent v3](https://www.shellrent.com/).


<!--more-->

- Code: `shellrentv3`
- Since: v5.8.0


Here is an example bash command using the Shellrent v3 provider:

```bash
SHELLRENTV3_CLIENT_ID="xxxxxxxxxxxxxxxxxxxxx" \
SHELLRENTV3_CLIENT_SECRET="yyyyyyyyyyyyyyyyyyyyy" \
lego run --dns shellrentv3 -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `SHELLRENTV3_CLIENT_ID` | OAuth2 client ID |
| `SHELLRENTV3_CLIENT_SECRET` | OAuth2 client secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `SHELLRENTV3_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `SHELLRENTV3_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `SHELLRENTV3_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

The OAuth2 client requires the scopes: `purchases:read`, `domains:read` and `domains:write`.



## More information

- [API documentation](https://guide.shellrent.com/api)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/shellrentv3/shellrentv3.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
