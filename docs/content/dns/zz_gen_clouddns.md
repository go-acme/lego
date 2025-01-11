---
title: "CloudDNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: clouddns
dnsprovider:
  since:    "v3.6.0"
  code:     "clouddns"
  url:      "https://vshosting.eu/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/clouddns/clouddns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [CloudDNS](https://vshosting.eu/).


<!--more-->

- Code: `clouddns`
- Since: v3.6.0


Here is an example bash command using the CloudDNS provider:

```bash
CLOUDDNS_CLIENT_ID=bLsdFAks23429841238feb177a572aX \
CLOUDDNS_EMAIL=you@example.com \
CLOUDDNS_PASSWORD=b9841238feb177a84330f \
lego --email you@example.com --dns clouddns -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `CLOUDDNS_CLIENT_ID` | Client ID |
| `CLOUDDNS_EMAIL` | Account email |
| `CLOUDDNS_PASSWORD` | Account password |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `CLOUDDNS_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `CLOUDDNS_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 5) |
| `CLOUDDNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 120) |
| `CLOUDDNS_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 300) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://admin.vshosting.cloud/clouddns/swagger/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/clouddns/clouddns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
