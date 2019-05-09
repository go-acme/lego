---
title: "HTTP request"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: httpreq
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/httpreq/httpreq.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v2.0.0

Configuration for [HTTP request](/dns/httpreq/).


<!--more-->

- Code: `httpreq`

Here is an example bash command using the HTTP request provider:

```bash
HTTPREQ_ENDPOINT=http://my.server.com:9090 \
lego --dns httpreq --domains my.domain.com --email my@email.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `HTTPREQ_ENDPOINT` | The URL of the server |
| `HTTPREQ_MODE` | `RAW`, none |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `HTTPREQ_HTTP_TIMEOUT` | API request timeout |
| `HTTPREQ_PASSWORD` | Basic authentication password |
| `HTTPREQ_POLLING_INTERVAL` | Time between DNS propagation check |
| `HTTPREQ_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `HTTPREQ_USERNAME` | Basic authentication username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).

## Description

The server must provide:

- `POST` `/present`
- `POST` `/cleanup`

The URL of the server must be define by `HTTPREQ_ENDPOINT`.

### Mode

There are 2 modes (`HTTPREQ_MODE`):

- default mode:
```json
{
  "fqdn": "_acme-challenge.domain.",
  "value": "LHDhK3oGRvkiefQnx7OOczTY5Tic_xZ6HcMOc_gmtoM"
}
```

- `RAW`
```json
{
  "domain": "domain",
  "token": "token",
  "keyAuth": "key"
}
```

### Authentication

Basic authentication (optional) can be set with some environment variables:

- `HTTPREQ_USERNAME` and `HTTPREQ_PASSWORD`
- both values must be set, otherwise basic authentication is not defined.





<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/httpreq/httpreq.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
