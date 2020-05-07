---
title: "Netlify"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: netlify
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/netlify/netlify.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.3.7

Configuration for [Netlify](https://www.netlify.com).


<!--more-->

- Code: `netlify`

Here is an example bash command using the Netlify provider:

```bash
NETLIFY_TOKEN=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
lego --dns netlify --domains my.domain.com --email my@email.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NETLIFY_TOKEN` | Token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NETLIFY_HTTP_TIMEOUT` | API request timeout |
| `NETLIFY_POLLING_INTERVAL` | Time between DNS propagation check |
| `NETLIFY_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `NETLIFY_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://open-api.netlify.com/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/netlify/netlify.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
