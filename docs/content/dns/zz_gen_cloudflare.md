---
title: "Cloudflare"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: cloudflare
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/cloudflare/cloudflare.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.3.0

Configuration for [Cloudflare](https://www.cloudflare.com/dns/).


<!--more-->

- Code: `cloudflare`

Here is an example bash command using the Cloudflare provider:

```bash
CLOUDFLARE_EMAIL=foo@bar.com \
CLOUDFLARE_API_KEY=b9841238feb177a84330febba8a83208921177bffe733 \
lego --dns cloudflare --domains my.domain.com --email my@email.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `CF_API_EMAIL` | Account email |
| `CF_API_KEY` | API key |
| `CLOUDFLARE_API_KEY` | Alias to CLOUDFLARE_API_KEY |
| `CLOUDFLARE_EMAIL` | Alias to CF_API_EMAIL |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `CLOUDFLARE_HTTP_TIMEOUT` | API request timeout |
| `CLOUDFLARE_POLLING_INTERVAL` | Time between DNS propagation check |
| `CLOUDFLARE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `CLOUDFLARE_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).

The Global API Key needs to be used, not the Origin CA Key.



## More information

- [API documentation](https://api.cloudflare.com/)
- [Go client](https://github.com/cloudflare/cloudflare-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/cloudflare/cloudflare.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
