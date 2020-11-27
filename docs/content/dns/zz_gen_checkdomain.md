---
title: "Checkdomain"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: checkdomain
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/checkdomain/checkdomain.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v3.3.0

Configuration for [Checkdomain](https://checkdomain.de/).


<!--more-->

- Code: `checkdomain`

Here is an example bash command using the Checkdomain provider:

```bash
CHECKDOMAIN_TOKEN=yoursecrettoken \
lego --email myemail@example.com --dns checkdomain --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `CHECKDOMAIN_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `CHECKDOMAIN_ENDPOINT` | API endpoint URL, defaults to https://api.checkdomain.de |
| `CHECKDOMAIN_HTTP_TIMEOUT` | API request timeout, defaults to 30 seconds |
| `CHECKDOMAIN_POLLING_INTERVAL` | Time between DNS propagation check |
| `CHECKDOMAIN_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `CHECKDOMAIN_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://developer.checkdomain.de/reference/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/checkdomain/checkdomain.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
