---
title: "Gyro"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: gyro
dnsprovider:
  since:    "v5.7.0"
  code:     "gyro"
  url:      "https://www.gyrohosting.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gyro/gyro.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Gyro](https://www.gyrohosting.com/).


<!--more-->

- Code: `gyro`
- Since: v5.7.0


Here is an example bash command using the Gyro provider:

```bash
GYRO_TOKEN="xxxxxxxxxxxxxxxxxxxxx" \
lego run --dns gyro -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `GYRO_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `GYRO_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `GYRO_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `GYRO_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `GYRO_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.gyrohosting.com/developers/en/#tag-managed-dns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gyro/gyro.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
