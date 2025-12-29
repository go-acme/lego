---
title: "ManageEngine CloudDNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: manageengine
dnsprovider:
  since:    "v4.21.0"
  code:     "manageengine"
  url:      "https://clouddns.manageengine.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/manageengine/manageengine.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [ManageEngine CloudDNS](https://clouddns.manageengine.com).


<!--more-->

- Code: `manageengine`
- Since: v4.21.0


Here is an example bash command using the ManageEngine CloudDNS provider:

```bash
MANAGEENGINE_CLIENT_ID="xxx" \
MANAGEENGINE_CLIENT_SECRET="yyy" \
lego --dns manageengine -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `MANAGEENGINE_CLIENT_ID` | Client ID |
| `MANAGEENGINE_CLIENT_SECRET` | Client Secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `MANAGEENGINE_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `MANAGEENGINE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `MANAGEENGINE_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://pitstop.manageengine.com/portal/en/kb/articles/manageengine-clouddns-rest-api-documentation)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/manageengine/manageengine.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
