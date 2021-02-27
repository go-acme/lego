---
title: "Designate DNSaaS for Openstack"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: designate
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/designate/designate.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v2.2.0

Configuration for [Designate DNSaaS for Openstack](https://docs.openstack.org/designate/latest/).


<!--more-->

- Code: `designate`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `OS_APPLICATION_CREDENTIAL_ID` | Application credential ID |
| `OS_APPLICATION_CREDENTIAL_NAME` | Application credential name |
| `OS_APPLICATION_CREDENTIAL_SECRET` | Application credential secret |
| `OS_AUTH_URL` | Identity endpoint URL |
| `OS_PASSWORD` | Password |
| `OS_PROJECT_NAME` | Project name |
| `OS_REGION_NAME` | Region name |
| `OS_USERNAME` | Username |
| `OS_USER_ID` | User ID |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DESIGNATE_POLLING_INTERVAL` | Time between DNS propagation check |
| `DESIGNATE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `DESIGNATE_TTL` | The TTL of the TXT record used for the DNS challenge |
| `OS_PROJECT_ID` | Project ID |
| `OS_TENANT_NAME` | Tenant name (deprecated see OS_PROJECT_NAME and OS_PROJECT_ID) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).

## Description

There are three main ways of authenticating with Designate:

The first one is by using the `OS_CLOUD` environment variable and a
`clouds.yaml` file.

The second one is using your username and password, via the `OS_USERNAME`,
`OS_PASSWORD` and `OS_PROJECT_NAME` environment variables.

The third one is by using an application credential, via the
`OS_APPLICATION_CREDENTIAL_*` and `OS_USER_ID` environment variables.

For the username/password and application methods, the `OS_AUTH_URL` and
`OS_REGION_NAME` environment variables are required.

For more information, you can read about the different method of authentication
with OpenStack in then Keystone's documentation and then gophercloud
documentation (links below).

[Keystone username/password](https://docs.openstack.org/keystone/latest/user/supported_clients.html)
[Keystone application credentials](https://docs.openstack.org/keystone/latest/user/application_credentials.html)



## More information

- [API documentation](https://docs.openstack.org/designate/latest/)
- [Go client](https://godoc.org/github.com/gophercloud/gophercloud/openstack/dns/v2)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/designate/designate.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
