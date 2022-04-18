---
title: "Bluecat"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: bluecat
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/bluecat/bluecat.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.5.0

Configuration for [Bluecat](https://www.bluecatnetworks.com).


<!--more-->

- Code: `bluecat`

Here is an example bash command using the Bluecat provider:

```bash
BLUECAT_PASSWORD=mypassword \
BLUECAT_DNS_VIEW=myview \
BLUECAT_USER_NAME=myusername \
BLUECAT_CONFIG_NAME=myconfig \
BLUECAT_SERVER_URL=https://bam.example.com \
BLUECAT_TTL=30 \
lego --email myemail@example.com --dns bluecat --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `BLUECAT_CONFIG_NAME` | Configuration name |
| `BLUECAT_DNS_VIEW` | External DNS View Name |
| `BLUECAT_PASSWORD` | API password |
| `BLUECAT_SERVER_URL` | The server URL, should have scheme, hostname, and port (if required) of the authoritative Bluecat BAM serve |
| `BLUECAT_USER_NAME` | API username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `BLUECAT_HTTP_TIMEOUT` | API request timeout |
| `BLUECAT_POLLING_INTERVAL` | Time between DNS propagation check |
| `BLUECAT_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `BLUECAT_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://docs.bluecatnetworks.com/r/Address-Manager-API-Guide/REST-API/9.1.0)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/bluecat/bluecat.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
