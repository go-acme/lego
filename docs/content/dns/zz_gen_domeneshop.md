---
title: "Domeneshop"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: domeneshop
dnsprovider:
  since:    "v4.3.0"
  code:     "domeneshop"
  url:      "https://domene.shop"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/domeneshop/domeneshop.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Domeneshop](https://domene.shop).


<!--more-->

- Code: `domeneshop`
- Since: v4.3.0


Here is an example bash command using the Domeneshop provider:

```bash
DOMENESHOP_API_TOKEN=<token> \
DOMENESHOP_API_SECRET=<secret> \
lego --email example@example.com --dns domeneshop -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DOMENESHOP_API_SECRET` | API secret |
| `DOMENESHOP_API_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DOMENESHOP_HTTP_TIMEOUT` | API request timeout |
| `DOMENESHOP_POLLING_INTERVAL` | Time between DNS propagation check |
| `DOMENESHOP_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

### API credentials

Visit the following page for information on how to create API credentials with Domeneshop:

  https://api.domeneshop.no/docs/#section/Authentication



## More information

- [API documentation](https://api.domeneshop.no/docs)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/domeneshop/domeneshop.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
