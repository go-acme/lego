---
title: "United-Domains"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: uniteddomains
dnsprovider:
  since:    "v4.29.0"
  code:     "uniteddomains"
  url:      "https://www.united-domains.de/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/uniteddomains/uniteddomains.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [United-Domains](https://www.united-domains.de/).


<!--more-->

- Code: `uniteddomains`
- Since: v4.29.0


Here is an example bash command using the United-Domains provider:

```bash
UNITEDDOMAINS_API_KEY=xxxxxxxx \
lego --dns uniteddomains -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `UNITEDDOMAINS_API_KEY` | API key `<prefix>.<secret>` https://www.united-domains.de/help/faq-article/getting-started-with-the-united-domains-dns-api/ |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `UNITEDDOMAINS_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `UNITEDDOMAINS_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `UNITEDDOMAINS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 900) |
| `UNITEDDOMAINS_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 300) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.united-domains.de/dns-apidoc/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/uniteddomains/uniteddomains.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
