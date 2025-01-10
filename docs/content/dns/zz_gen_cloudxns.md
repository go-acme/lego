---
title: "CloudXNS (Deprecated)"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: cloudxns
dnsprovider:
  since:    "v0.5.0"
  code:     "cloudxns"
  url:      "https://github.com/go-acme/lego/issues/2323"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/cloudxns/cloudxns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

The CloudXNS DNS provider has shut down.



<!--more-->

- Code: `cloudxns`
- Since: v0.5.0


Here is an example bash command using the CloudXNS (Deprecated) provider:

```bash
CLOUDXNS_API_KEY=xxxx \
CLOUDXNS_SECRET_KEY=yyyy \
lego --email you@example.com --dns cloudxns -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `CLOUDXNS_API_KEY` | The API key |
| `CLOUDXNS_SECRET_KEY` | The API secret key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `CLOUDXNS_HTTP_TIMEOUT` | API request timeout in seconds (Default: ) |
| `CLOUDXNS_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: ) |
| `CLOUDXNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: ) |
| `CLOUDXNS_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: ) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).





<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/cloudxns/cloudxns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
