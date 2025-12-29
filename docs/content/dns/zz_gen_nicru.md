---
title: "RU CENTER"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: nicru
dnsprovider:
  since:    "v4.24.0"
  code:     "nicru"
  url:      "https://nic.ru/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nicru/nicru.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [RU CENTER](https://nic.ru/).


<!--more-->

- Code: `nicru`
- Since: v4.24.0


Here is an example bash command using the RU CENTER provider:

```bash
NICRU_USER="<your_user>" \
NICRU_PASSWORD="<your_password>" \
NICRU_SERVICE_ID="<service_id>" \
NICRU_SECRET="<service_secret>" \
lego --dns nicru -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NICRU_PASSWORD` | Password for an account in RU CENTER |
| `NICRU_SECRET` | Secret for application in DNS-hosting RU CENTER |
| `NICRU_SERVICE_ID` | Service ID for application in DNS-hosting RU CENTER |
| `NICRU_SERVICE_NAME` | Service Name for DNS-hosting RU CENTER |
| `NICRU_USER` | Agreement for an account in RU CENTER |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NICRU_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 60) |
| `NICRU_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 600) |
| `NICRU_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 30) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

## Credential information

You can find information about service ID and secret https://www.nic.ru/manager/oauth.cgi?step=oauth.app_list

| ENV Variable        | Parameter from page            | Example           |
|---------------------|--------------------------------|-------------------|
| NICRU_USER          | Username (Number of agreement) | NNNNNNN/NIC-D     |
| NICRU_PASSWORD      | Password account               |                   |
| NICRU_SERVICE_ID    | Application ID                 | hex-based, len 32 |
| NICRU_SECRET        | Identity endpoint              | string len 91     |



## More information

- [API documentation](https://www.nic.ru/help/api-dns-hostinga_3643.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nicru/nicru.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
