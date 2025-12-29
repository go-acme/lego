---
title: "Open Telekom Cloud"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: otc
dnsprovider:
  since:    "v0.4.1"
  code:     "otc"
  url:      "https://cloud.telekom.de/en"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/otc/otc.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Open Telekom Cloud](https://cloud.telekom.de/en).


<!--more-->

- Code: `otc`
- Since: v0.4.1


Here is an example bash command using the Open Telekom Cloud provider:

```bash
OTC_DOMAIN_NAME=domain_name \
OTC_USER_NAME=user_name \
OTC_PASSWORD=password \
OTC_PROJECT_NAME=project_name \
lego --dns otc -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `OTC_DOMAIN_NAME` | Domain name |
| `OTC_PASSWORD` | Password |
| `OTC_PROJECT_NAME` | Project name |
| `OTC_USER_NAME` | User name |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `OTC_HTTP_TIMEOUT` | API request timeout in seconds (Default: 10) |
| `OTC_IDENTITY_ENDPOINT` | Identity endpoint URL (default: https://iam.eu-de.otc.t-systems.com:443/v3/auth/tokens) |
| `OTC_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `OTC_PRIVATE_ZONE` | Set to true to use private zones only (default: use public zones only) |
| `OTC_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `OTC_SEQUENCE_INTERVAL` | Time between sequential requests in seconds (Default: 60) |
| `OTC_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 300) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://docs.otc.t-systems.com/domain-name-service/api-ref/index.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/otc/otc.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
