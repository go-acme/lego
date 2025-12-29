---
title: "Shellrent"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: shellrent
dnsprovider:
  since:    "v4.16.0"
  code:     "shellrent"
  url:      "https://www.shellrent.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/shellrent/shellrent.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Shellrent](https://www.shellrent.com/).


<!--more-->

- Code: `shellrent`
- Since: v4.16.0


Here is an example bash command using the Shellrent provider:

```bash
SHELLRENT_USERNAME=xxxx \
SHELLRENT_TOKEN=yyyy \
lego --dns shellrent -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `SHELLRENT_TOKEN` | Token |
| `SHELLRENT_USERNAME` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `SHELLRENT_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `SHELLRENT_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 10) |
| `SHELLRENT_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 300) |
| `SHELLRENT_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 3600) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api.shellrent.com/section/api2)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/shellrent/shellrent.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
