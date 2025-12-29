---
title: "VK Cloud"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: vkcloud
dnsprovider:
  since:    "v4.9.0"
  code:     "vkcloud"
  url:      "https://mcs.mail.ru/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/vkcloud/vkcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [VK Cloud](https://mcs.mail.ru/).


<!--more-->

- Code: `vkcloud`
- Since: v4.9.0


Here is an example bash command using the VK Cloud provider:

```bash
VK_CLOUD_PROJECT_ID="<your_project_id>" \
VK_CLOUD_USERNAME="<your_email>" \
VK_CLOUD_PASSWORD="<your_password>" \
lego --dns vkcloud -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `VK_CLOUD_PASSWORD` | Password for VK Cloud account |
| `VK_CLOUD_PROJECT_ID` | String ID of project in VK Cloud |
| `VK_CLOUD_USERNAME` | Email of VK Cloud account |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `VK_CLOUD_DNS_ENDPOINT` | URL of DNS API. Defaults to https://mcs.mail.ru/public-dns but can be changed for usage with private clouds |
| `VK_CLOUD_DOMAIN_NAME` | Openstack users domain name. Defaults to `users` but can be changed for usage with private clouds |
| `VK_CLOUD_IDENTITY_ENDPOINT` | URL of OpenStack Auth API, Defaults to https://infra.mail.ru:35357/v3/ but can be changed for usage with private clouds |
| `VK_CLOUD_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `VK_CLOUD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `VK_CLOUD_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

## Credential information

You can find all required and additional information on ["Project/Keys" page](https://mcs.mail.ru/app/en/project/keys) of your cloud.

| ENV Variable               | Parameter from page |
|----------------------------|---------------------|
| VK_CLOUD_PROJECT_ID        | Project ID          |
| VK_CLOUD_USERNAME          | Username            |
| VK_CLOUD_DOMAIN_NAME       | User Domain Name    |
| VK_CLOUD_IDENTITY_ENDPOINT | Identity endpoint   |



## More information

- [API documentation](https://mcs.mail.ru/docs/networks/vnet/networks/publicdns/api)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/vkcloud/vkcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
