---
title: "Designate DNSaaS for Openstack"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: designate
dnsprovider:
  since:    "v2.2.0"
  code:     "designate"
  url:      "https://docs.openstack.org/designate/latest/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/designate/designate.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Designate DNSaaS for Openstack](https://docs.openstack.org/designate/latest/).


<!--more-->

- Code: `designate`
- Since: v2.2.0


Here is an example bash command using the Designate DNSaaS for Openstack provider:

```bash
# With a `clouds.yaml`
OS_CLOUD=my_openstack \
lego --email you@example.com --dns designate --domains my.example.org run

# or

OS_AUTH_URL=https://openstack.example.org \
OS_REGION_NAME=RegionOne \
OS_PROJECT_ID=23d4522a987d4ab529f722a007c27846
OS_USERNAME=myuser \
OS_PASSWORD=passw0rd \
lego --email you@example.com --dns designate --domains my.example.org run

# or

OS_AUTH_URL=https://openstack.example.org \
OS_REGION_NAME=RegionOne \
OS_AUTH_TYPE=v3applicationcredential \
OS_APPLICATION_CREDENTIAL_ID=imn74uq0or7dyzz20dwo1ytls4me8dry \
OS_APPLICATION_CREDENTIAL_SECRET=68FuSPSdQqkFQYH5X1OoriEIJOwyLtQ8QSqXZOc9XxFK1A9tzZT6He2PfPw0OMja \
lego --email you@example.com --dns designate --domains my.example.org run
```




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
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DESIGNATE_POLLING_INTERVAL` | Time between DNS propagation check |
| `DESIGNATE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `DESIGNATE_TTL` | The TTL of the TXT record used for the DNS challenge |
| `OS_PROJECT_ID` | Project ID |
| `OS_TENANT_NAME` | Tenant name (deprecated see OS_PROJECT_NAME and OS_PROJECT_ID) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).

## Description

There are three main ways of authenticating with Designate:

1. The first one is by using the `OS_CLOUD` environment variable and a `clouds.yaml` file.
2. The second one is using your username and password, via the `OS_USERNAME`, `OS_PASSWORD` and `OS_PROJECT_NAME` environment variables.
3. The third one is by using an application credential, via the `OS_APPLICATION_CREDENTIAL_*` and `OS_USER_ID` environment variables.

For the username/password and application methods, the `OS_AUTH_URL` and `OS_REGION_NAME` environment variables are required.

For more information, you can read about the different methods of authentication with OpenStack in the Keystone's documentation and the gophercloud documentation:

- [Keystone username/password](https://docs.openstack.org/keystone/latest/user/supported_clients.html)
- [Keystone application credentials](https://docs.openstack.org/keystone/latest/user/application_credentials.html)

Public cloud providers with support for Designate:

- [Fuga Cloud](https://fuga.cloud/)



## More information

- [API documentation](https://docs.openstack.org/designate/latest/)
- [Go client](https://pkg.go.dev/github.com/gophercloud/gophercloud/openstack/dns/v2)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/designate/designate.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
