---
title: "Ionos Cloud"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: ionoscloud
dnsprovider:
  since:    "v4.30.0"
  code:     "ionoscloud"
  url:      "https://cloud.ionos.de/network/cloud-dns"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ionoscloud/ionoscloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Ionos Cloud](https://cloud.ionos.de/network/cloud-dns).


<!--more-->

- Code: `ionoscloud`
- Since: v4.30.0


Here is an example bash command using the Ionos Cloud provider:

```bash
IONOSCLOUD_API_TOKEN="xxxxxxxxxxxxxxxxxxxxx" \
lego --email you@example.com --dns ionoscloud -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `IONOSCLOUD_API_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `IONOSCLOUD_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `IONOSCLOUD_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `IONOSCLOUD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 120) |
| `IONOSCLOUD_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api.ionos.com/docs/dns/v1/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ionoscloud/ionoscloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
