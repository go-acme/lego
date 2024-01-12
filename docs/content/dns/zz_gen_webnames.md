---
title: "Webnames"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: webnames
dnsprovider:
  since:    "v4.15.0"
  code:     "webnames"
  url:      "https://www.webnames.ru/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/webnames/webnames.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Webnames](https://www.webnames.ru/).


<!--more-->

- Code: `webnames`
- Since: v4.15.0


Here is an example bash command using the Webnames provider:

```bash
WEBNAMES_API_KEY=xxxxxx \
lego --email you@example.com --dns webnames --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `WEBNAMES_API_KEY` | Domain API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `WEBNAMES_HTTP_TIMEOUT` | API request timeout |
| `WEBNAMES_POLLING_INTERVAL` | Time between DNS propagation check |
| `WEBNAMES_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `WEBNAMES_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).

## API Key

To obtain the key, you need to change the DNS server to `*.nameself.com`: Personal account / My domains and services / Select the required domain / DNS servers

The API key can be found: Personal account / My domains and services / Select the required domain / Zone management / acme.sh or certbot settings



## More information

- [API documentation](https://github.com/regtime-ltd/certbot-dns-webnames)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/webnames/webnames.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
