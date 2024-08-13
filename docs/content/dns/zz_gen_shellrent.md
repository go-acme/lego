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
lego --email you@example.com --dns shellrent --domains my.example.org run
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
| `SHELLRENT_HTTP_TIMEOUT` | API request timeout |
| `SHELLRENT_POLLING_INTERVAL` | Time between DNS propagation check |
| `SHELLRENT_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `SHELLRENT_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api.shellrent.com/section/api2)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/shellrent/shellrent.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
