---
title: "Scaleway"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: scaleway
dnsprovider:
  since:    "v3.4.0"
  code:     "scaleway"
  url:      "https://developers.scaleway.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/scaleway/scaleway.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Scaleway](https://developers.scaleway.com/).


<!--more-->

- Code: `scaleway`
- Since: v3.4.0


Here is an example bash command using the Scaleway provider:

```bash
SCW_SECRET_KEY=xxxxxxx-xxxxx-xxxx-xxx-xxxxxx \
lego --email you@example.com --dns scaleway -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `SCW_PROJECT_ID` | Project to use (optional) |
| `SCW_SECRET_KEY` | Secret key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `SCW_ACCESS_KEY` | Access key |
| `SCW_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `SCW_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 10) |
| `SCW_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 120) |
| `SCW_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://developers.scaleway.com/en/products/domain/dns/api/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/scaleway/scaleway.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
