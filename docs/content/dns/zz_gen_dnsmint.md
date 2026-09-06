---
title: "DNSMint"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dnsmint
dnsprovider:
  since:    "v5.5.0"
  code:     "dnsmint"
  url:      "https://dnsmint.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnsmint/dnsmint.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [DNSMint](https://dnsmint.com/).


<!--more-->

- Code: `dnsmint`
- Since: v5.5.0


Here is an example bash command using the DNSMint provider:

```bash
DNSMINT_API_KEY="xxx" \
lego --dns dnsmint -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DNSMINT_API_KEY` | API key with the dns01:write scope |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DNSMINT_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `DNSMINT_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `DNSMINT_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `DNSMINT_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://dnsmint.com/api-reference)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnsmint/dnsmint.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
