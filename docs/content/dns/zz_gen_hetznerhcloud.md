---
title: "Hetzner Cloud"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: hetznerhcloud
dnsprovider:
  since:    "v4.28.0"
  code:     "hetznerhcloud"
  url:      "https://docs.hetzner.cloud/reference/dns"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hetznerhcloud/hetznerhcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Hetzner Cloud](https://docs.hetzner.cloud/reference/dns).


<!--more-->

- Code: `hetznerhcloud`
- Since: v4.28.0


Here is an example bash command using the Hetzner Cloud provider:

```bash
HCLOUD_TOKEN="xxxxxx" \
lego --email you@example.com --dns hetznerhcloud -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `HCLOUD_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `HCLOUD_BASE_URL` | Override the Hetzner Cloud API base URL (Default: https://api.hetzner.cloud) |
| `HCLOUD_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `HCLOUD_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 5) |
| `HCLOUD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 120) |
| `HCLOUD_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://docs.hetzner.cloud/reference/dns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hetznerhcloud/hetznerhcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
