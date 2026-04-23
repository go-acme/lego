---
title: "Rage4"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: rage4
dnsprovider:
  since:    "v5.0.0"
  code:     "rage4"
  url:      "https://rage4.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/rage4/rage4.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Rage4](https://rage4.com/).


<!--more-->

- Code: `rage4`
- Since: v5.0.0


Here is an example bash command using the Rage4 provider:

```bash
RAGE4_USERNAME="xxx" \
RAGE4_PASSWORD="yyy" \
lego --dns rage4 -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `RAGE4_PASSWORD` | Password |
| `RAGE4_USERNAME` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `RAGE4_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `RAGE4_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `RAGE4_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `RAGE4_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://rage4.com/swagger/index.html?urls.primaryName=dns-legacy#/DNS)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/rage4/rage4.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
