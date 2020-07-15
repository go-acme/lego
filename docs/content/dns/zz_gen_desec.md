---
title: "deSEC.io"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: desec
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/desec/desec.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v3.7.0

Configuration for [deSEC.io](https://desec.io).


<!--more-->

- Code: `desec`

Here is an example bash command using the deSEC.io provider:

```bash
DESEC_TOKEN=x-xxxxxxxxxxxxxxxxxxxxxxxxxx \
lego --dns desec --domains my.domain.com --email my@email.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DESEC_TOKEN` | Domain token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DESEC_HTTP_TIMEOUT` | API request timeout |
| `DESEC_POLLING_INTERVAL` | Time between DNS propagation check |
| `DESEC_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `DESEC_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://desec.readthedocs.io/en/latest/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/desec/desec.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
