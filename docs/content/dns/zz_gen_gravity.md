---
title: "Gravity"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: gravity
dnsprovider:
  since:    "v4.30.0"
  code:     "gravity"
  url:      "https://gravity.beryju.io/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gravity/gravity.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Gravity](https://gravity.beryju.io/).


<!--more-->

- Code: `gravity`
- Since: v4.30.0


Here is an example bash command using the Gravity provider:

```bash
GRAVITY_SERVER_URL="https://example.org:1234" \
GRAVITY_USERNAME="xxxxxxxxxxxxxxxxxxxxx" \
GRAVITY_PASSWORD="yyyyyyyyyyyyyyyyyyyyy" \
lego --email you@example.com --dns gravity -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `GRAVITY_PASSWORD` | Password |
| `GRAVITY_SERVER_URL` | URL of the server |
| `GRAVITY_USERNAME` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `GRAVITY_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `GRAVITY_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `GRAVITY_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `GRAVITY_SEQUENCE_INTERVAL` | Time between sequential requests in seconds (Default: 1) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://gravity.beryju.io/docs/api/reference/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gravity/gravity.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
