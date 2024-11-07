---
title: "Vercel"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: vercel
dnsprovider:
  since:    "v4.7.0"
  code:     "vercel"
  url:      "https://vercel.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/vercel/vercel.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Vercel](https://vercel.com).


<!--more-->

- Code: `vercel`
- Since: v4.7.0


Here is an example bash command using the Vercel provider:

```bash
VERCEL_API_TOKEN=xxxxxx \
lego --email you@example.com --dns vercel -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `VERCEL_API_TOKEN` | Authentication token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `VERCEL_HTTP_TIMEOUT` | API request timeout |
| `VERCEL_POLLING_INTERVAL` | Time between DNS propagation check |
| `VERCEL_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `VERCEL_TEAM_ID` | Team ID (ex: team_xxxxxxxxxxxxxxxxxxxxxxxx) |
| `VERCEL_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://vercel.com/docs/rest-api#endpoints/dns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/vercel/vercel.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
