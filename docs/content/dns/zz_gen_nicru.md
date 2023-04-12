---
title: "RU CENTER"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: nicru
dnsprovider:
  since:    "v4.11.0"
  code:     "nicru"
  url:      "https://nic.ru/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nicru/nicru.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [RU CENTER](https://nic.ru/).


<!--more-->

- Code: `nicru`
- Since: v4.11.0


Here is an example bash command using the RU CENTER provider:

```bash
NIC_RU_USER="<your_user>" \
NIC_RU_PASSWORD="<your_password>" \
NIC_RU_SERVICE_ID="<service_id>" \
NIC_RU_SECRET="<service_secret>" \
NIC_RU_SERVICE_NAME="<service_name>" \
./lego --dns nicru --domains "*.example.com" --email you@example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NIC_RU_PASSWORD` | Password for account in RU CENTER |
| `NIC_RU_SECRET` | Secret for application in DNS-hosting RU CENTER |
| `NIC_RU_SERVICE_ID` | Service ID for application in DNS-hosting RU CENTER |
| `NIC_RU_SERVICE_NAME` | Service Name for DNS-hosting RU CENTER |
| `NIC_RU_USER` | Agreement for account in RU CENTER |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NIC_RU_HTTP_TIMEOUT` | API request timeout |
| `NIC_RU_POLLING_INTERVAL` | Time between DNS propagation check |
| `NIC_RU_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `NIC_RU_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

## Credential inforamtion

You can find information about service ID and secret https://www.nic.ru/manager/oauth.cgi?step=oauth.app_list

| ENV Variable         | Parameter from page            | Example           |
|----------------------|--------------------------------|-------------------|
| NIC_RU_USER          | Username (Number of agreement) | NNNNNNN/NIC-D     |
| NIC_RU_PASSWORD      | Password account               |                   |
| NIC_RU_SERVICE_ID    | Application ID                 | hex-based, len 32 |
| NIC_RU_SECRET        | Identity endpoint              | string len 91     |
| NIC_RU_SERVICE_NAME  | Service name in DNS-hosting    | DPNNNNNNNNNN      |



## More information

- [API documentation](https://www.nic.ru/help/api-dns-hostinga_3643.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nicru/nicru.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
