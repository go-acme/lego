---
title: "DNSPod (deprecated)"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dnspod
dnsprovider:
  since:    "v0.4.0"
  code:     "dnspod"
  url:      "https://www.dnspod.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnspod/dnspod.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Use the Tencent Cloud provider instead.



<!--more-->

- Code: `dnspod`
- Since: v0.4.0


Here is an example bash command using the DNSPod (deprecated) provider:

```bash
DNSPOD_API_KEY=xxxxxx \
lego --email you@example.com --dns dnspod --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DNSPOD_API_KEY` | The user token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DNSPOD_HTTP_TIMEOUT` | API request timeout |
| `DNSPOD_POLLING_INTERVAL` | Time between DNS propagation check |
| `DNSPOD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `DNSPOD_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://docs.dnspod.com/api/)
- [Go client](https://github.com/nrdcg/dnspod-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnspod/dnspod.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
