---
title: "VegaDNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: vegadns
dnsprovider:
  since:    "v1.1.0"
  code:     "vegadns"
  url:      "https://github.com/shupp/VegaDNS-API"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/vegadns/vegadns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [VegaDNS](https://github.com/shupp/VegaDNS-API).


<!--more-->

- Code: `vegadns`
- Since: v1.1.0


{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `SECRET_VEGADNS_KEY` | API key |
| `SECRET_VEGADNS_SECRET` | API secret |
| `VEGADNS_URL` | API endpoint URL |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `VEGADNS_POLLING_INTERVAL` | Time between DNS propagation check |
| `VEGADNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `VEGADNS_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://github.com/shupp/VegaDNS-API)
- [Go client](https://github.com/OpenDNS/vegadns2client)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/vegadns/vegadns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
