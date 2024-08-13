---
title: "NIFCloud"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: nifcloud
dnsprovider:
  since:    "v1.1.0"
  code:     "nifcloud"
  url:      "https://www.nifcloud.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nifcloud/nifcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [NIFCloud](https://www.nifcloud.com/).


<!--more-->

- Code: `nifcloud`
- Since: v1.1.0


Here is an example bash command using the NIFCloud provider:

```bash
NIFCLOUD_ACCESS_KEY_ID=xxxx \
NIFCLOUD_SECRET_ACCESS_KEY=yyyy \
lego --email you@example.com --dns nifcloud --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NIFCLOUD_ACCESS_KEY_ID` | Access key |
| `NIFCLOUD_SECRET_ACCESS_KEY` | Secret access key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NIFCLOUD_HTTP_TIMEOUT` | API request timeout |
| `NIFCLOUD_POLLING_INTERVAL` | Time between DNS propagation check |
| `NIFCLOUD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `NIFCLOUD_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://mbaas.nifcloud.com/doc/current/rest/common/format.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nifcloud/nifcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
