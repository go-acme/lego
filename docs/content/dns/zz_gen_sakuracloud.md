---
title: "Sakura Cloud"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: sakuracloud
dnsprovider:
  since:    "v1.1.0"
  code:     "sakuracloud"
  url:      "https://cloud.sakura.ad.jp/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/sakuracloud/sakuracloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Sakura Cloud](https://cloud.sakura.ad.jp/).


<!--more-->

- Code: `sakuracloud`
- Since: v1.1.0


Here is an example bash command using the Sakura Cloud provider:

```bash
SAKURACLOUD_ACCESS_TOKEN=xxxxx \
SAKURACLOUD_ACCESS_TOKEN_SECRET=yyyyy \
lego --email you@example.com --dns sakuracloud --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `SAKURACLOUD_ACCESS_TOKEN` | Access token |
| `SAKURACLOUD_ACCESS_TOKEN_SECRET` | Access token secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `SAKURACLOUD_HTTP_TIMEOUT` | API request timeout |
| `SAKURACLOUD_POLLING_INTERVAL` | Time between DNS propagation check |
| `SAKURACLOUD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `SAKURACLOUD_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://developer.sakura.ad.jp/cloud/api/1.1/)
- [Go client](https://github.com/sacloud/iaas-api-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/sakuracloud/sakuracloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
