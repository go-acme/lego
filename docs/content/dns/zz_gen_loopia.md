---
title: "Loopia"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: loopia
dnsprovider:
  since:    "v4.2.0"
  code:     "loopia"
  url:      "https://loopia.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/loopia/loopia.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Loopia](https://loopia.com).


<!--more-->

- Code: `loopia`
- Since: v4.2.0


Here is an example bash command using the Loopia provider:

```bash
LOOPIA_API_USER=xxxxxxxx \
LOOPIA_API_PASSWORD=yyyyyyyy \
lego --dns loopia -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `LOOPIA_API_PASSWORD` | API password |
| `LOOPIA_API_USER` | API username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `LOOPIA_API_URL` | API endpoint. Ex: https://api.loopia.se/RPCSERV or https://api.loopia.rs/RPCSERV |
| `LOOPIA_HTTP_TIMEOUT` | API request timeout in seconds (Default: 60) |
| `LOOPIA_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2400) |
| `LOOPIA_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `LOOPIA_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 300) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

### API user

You can [generate a new API user](https://customerzone.loopia.com/api/) from your account page.

It needs to have the following permissions:

* addZoneRecord
* getZoneRecords
* removeZoneRecord
* removeSubdomain



## More information

- [API documentation](https://www.loopia.com/api)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/loopia/loopia.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
