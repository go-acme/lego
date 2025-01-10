---
title: "Timeweb Cloud"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: timewebcloud
dnsprovider:
  since:    "v4.20.0"
  code:     "timewebcloud"
  url:      "https://timeweb.cloud/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/timewebcloud/timewebcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Timeweb Cloud](https://timeweb.cloud/).


<!--more-->

- Code: `timewebcloud`
- Since: v4.20.0


Here is an example bash command using the Timeweb Cloud provider:

```bash
TIMEWEBCLOUD_AUTH_TOKEN=xxxxxx \
lego --email you@example.com --dns timewebcloud -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `TIMEWEBCLOUD_AUTH_TOKEN` | Authentication token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `TIMEWEBCLOUD_HTTP_TIMEOUT` | API request timeout in seconds (Default: 10) |
| `TIMEWEBCLOUD_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `TIMEWEBCLOUD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://timeweb.cloud/api-docs)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/timewebcloud/timewebcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
