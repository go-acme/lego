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


{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `OTC_DOMAIN_NAME` | Domain name |
| `OTC_IDENTITY_ENDPOINT` | Identity endpoint URL |
| `OTC_PASSWORD` | Password |
| `OTC_PROJECT_NAME` | Project name |
| `OTC_USER_NAME` | User name |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `OTC_HTTP_TIMEOUT` | API request timeout |
| `OTC_POLLING_INTERVAL` | Time between DNS propagation check |
| `OTC_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `OTC_SEQUENCE_INTERVAL` | Time between sequential requests |
| `OTC_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://docs.otc.t-systems.com/domain-name-service/api-ref/index.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/otc/otc.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
