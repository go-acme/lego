---
title: "Google Domains"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: googledomains
dnsprovider:
  since:    "v4.11.0"
  code:     "googledomains"
  url:      "https://github.com/go-acme/lego/issues/2553"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/googledomains/googledomains.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

The Google Domains DNS provider has shut down.



<!--more-->

- Code: `googledomains`
- Since: v4.11.0


Here is an example bash command using the Google Domains provider:

```bash
GOOGLE_DOMAINS_ACCESS_TOKEN=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
lego --email you@example.com --dns googledomains -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `GOOGLE_DOMAINS_ACCESS_TOKEN` | Access token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `GOOGLE_DOMAINS_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `GOOGLE_DOMAINS_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `GOOGLE_DOMAINS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information


- [Go client](https://github.com/googleapis/google-api-go-client)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/googledomains/googledomains.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
