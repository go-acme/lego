---
title: "Namesilo"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: namesilo
dnsprovider:
  since:    "v2.7.0"
  code:     "namesilo"
  url:      "https://www.namesilo.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/namesilo/namesilo.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Namesilo](https://www.namesilo.com/).


<!--more-->

- Code: `namesilo`
- Since: v2.7.0


Here is an example bash command using the Namesilo provider:

```bash
NAMESILO_API_KEY=b9841238feb177a84330febba8a83208921177bffe733 \
lego --email you@example.com --dns namesilo --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NAMESILO_API_KEY` | Client ID |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NAMESILO_POLLING_INTERVAL` | Time between DNS propagation check |
| `NAMESILO_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation, it is better to set larger than 15m |
| `NAMESILO_TTL` | The TTL of the TXT record used for the DNS challenge, should be in [3600, 2592000] |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://www.namesilo.com/api_reference.php)
- [Go client](https://github.com/nrdcg/namesilo)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/namesilo/namesilo.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
