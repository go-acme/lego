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
# Application Key authentication:

OVH_APPLICATION_KEY=1234567898765432 \
OVH_APPLICATION_SECRET=b9841238feb177a84330febba8a832089 \
OVH_CONSUMER_KEY=256vfsd347245sdfg \
OVH_ENDPOINT=ovh-eu \
lego --email you@example.com --dns ovh -d '*.example.com' -d example.com run

# Or Access Token:

OVH_ACCESS_TOKEN=xxx \
OVH_ENDPOINT=ovh-eu \
lego --email you@example.com --dns ovh -d '*.example.com' -d example.com run

# Or OAuth2:

OVH_CLIENT_ID=yyy \
OVH_CLIENT_SECRET=xxx \
OVH_ENDPOINT=ovh-eu \
lego --email you@example.com --dns ovh -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `OVH_ACCESS_TOKEN` | Access token |
| `OVH_APPLICATION_KEY` | Application key (Application Key authentication) |
| `OVH_APPLICATION_SECRET` | Application secret (Application Key authentication) |
| `OVH_CLIENT_ID` | Client ID (OAuth2) |
| `OVH_CLIENT_SECRET` | Client secret (OAuth2) |
| `OVH_CONSUMER_KEY` | Consumer key (Application Key authentication) |
| `OVH_ENDPOINT` | Endpoint URL (ovh-eu or ovh-ca) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `OVH_HTTP_TIMEOUT` | API request timeout in seconds (Default: 180) |
| `OVH_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `OVH_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `OVH_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

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

## OAuth2 Client Credentials

Another method for authentication is by using OAuth2 client credentials.

An IAM policy and service account can be created by following the [OVH guide](https://help.ovhcloud.com/csm/en-manage-service-account?id=kb_article_view&sysparm_article=KB0059343).

Following IAM policies need to be authorized for the affected domain:

* dnsZone:apiovh:record/create
* dnsZone:apiovh:record/delete
* dnsZone:apiovh:refresh

## Important Note

Both authentication methods cannot be used at the same time.



## More information

- [API documentation](https://eu.api.ovh.com/)
- [Go client](https://github.com/ovh/go-ovh)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ovh/ovh.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
