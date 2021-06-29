---
title: "Internet.bs"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: internetbs
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/internetbs/internetbs.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v4.5.0

Configuration for [Internet.bs](https://internetbs.net).


<!--more-->

- Code: `internetbs`

Here is an example bash command using the Internet.bs provider:

```bash
INTERNET_BS_API_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxx \
INTERNET_BS_PASSWORD=yyyyyyyyyyyyyyyyyyyyyyyyyy \
lego --email myemail@example.com --dns internetbs --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `INTERNET_BS_API_KEY` | API key |
| `INTERNET_BS_PASSWORD` | API password |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `INTERNET_BS_HTTP_TIMEOUT` | API request timeout |
| `INTERNET_BS_POLLING_INTERVAL` | Time between DNS propagation check |
| `INTERNET_BS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `INTERNET_BS_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://internetbs.net/internet-bs-api.pdf)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/internetbs/internetbs.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
