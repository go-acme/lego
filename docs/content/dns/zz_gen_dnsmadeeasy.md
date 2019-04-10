---
title: "DNS Made Easy"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dnsmadeeasy
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnsmadeeasy/dnsmadeeasy.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.4.0

Configuration for [DNS Made Easy](https://dnsmadeeasy.com/).


<!--more-->

- Code: `dnsmadeeasy`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DNSMADEEASY_API_KEY` | The API key |
| `DNSMADEEASY_API_SECRET` | The API Secret key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DNSMADEEASY_HTTP_TIMEOUT` | API request timeout |
| `DNSMADEEASY_POLLING_INTERVAL` | Time between DNS propagation check |
| `DNSMADEEASY_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `DNSMADEEASY_SANDBOX` | Activate the sandbox (boolean) |
| `DNSMADEEASY_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://api-docs.dnsmadeeasy.com/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnsmadeeasy/dnsmadeeasy.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
