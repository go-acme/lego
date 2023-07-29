---
title: "OVH"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: ovh
dnsprovider:
  since:    "v0.4.0"
  code:     "ovh"
  url:      "https://www.ovh.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ovh/ovh.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [OVH](https://www.ovh.com/).


<!--more-->

- Code: `ovh`
- Since: v0.4.0


Here is an example bash command using the OVH provider:

```bash
OVH_APPLICATION_KEY=1234567898765432 \
OVH_APPLICATION_SECRET=b9841238feb177a84330febba8a832089 \
OVH_CONSUMER_KEY=256vfsd347245sdfg \
OVH_ENDPOINT=ovh-eu \
lego --email you@example.com --dns ovh --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `OVH_APPLICATION_KEY` | Application key |
| `OVH_APPLICATION_SECRET` | Application secret |
| `OVH_CONSUMER_KEY` | Consumer key |
| `OVH_ENDPOINT` | Endpoint URL (ovh-eu or ovh-ca) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `OVH_HTTP_TIMEOUT` | API request timeout |
| `OVH_POLLING_INTERVAL` | Time between DNS propagation check |
| `OVH_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `OVH_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).

## Application Key and Secret

Application key and secret can be created by following the [OVH guide](https://docs.ovh.com/gb/en/customer/first-steps-with-ovh-api/).

When requesting the consumer key, the following configuration can be used to define access rights:

```json
{
  "accessRules": [
    {
      "method": "POST",
      "path": "/domain/zone/*"
    },
    {
      "method": "DELETE",
      "path": "/domain/zone/*"
    }
  ]
}
```



## More information

- [API documentation](https://eu.api.ovh.com/)
- [Go client](https://github.com/ovh/go-ovh)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ovh/ovh.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
