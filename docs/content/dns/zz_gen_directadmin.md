---
title: "DirectAdmin"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: directadmin
dnsprovider:
  since:    "v4.18.0"
  code:     "directadmin"
  url:      "https://www.directadmin.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/directadmin/directadmin.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [DirectAdmin](https://www.directadmin.com).


<!--more-->

- Code: `directadmin`
- Since: v4.18.0


Here is an example bash command using the DirectAdmin provider:

```bash
DIRECTADMIN_API_URL="http://example.com:2222" \
DIRECTADMIN_USERNAME=xxxx \
DIRECTADMIN_PASSWORD=yyy \
lego --dns directadmin -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DIRECTADMIN_API_URL` | URL of the API |
| `DIRECTADMIN_PASSWORD` | API password |
| `DIRECTADMIN_USERNAME` | API username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DIRECTADMIN_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `DIRECTADMIN_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 5) |
| `DIRECTADMIN_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `DIRECTADMIN_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 30) |
| `DIRECTADMIN_ZONE_NAME` | Zone name used to add the TXT record |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.directadmin.com/api.php)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/directadmin/directadmin.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
