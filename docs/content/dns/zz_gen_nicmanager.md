---
title: "Nicmanager"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: nicmanager
dnsprovider:
  since:    "v4.5.0"
  code:     "nicmanager"
  url:      "https://www.nicmanager.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nicmanager/nicmanager.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Nicmanager](https://www.nicmanager.com/).


<!--more-->

- Code: `nicmanager`
- Since: v4.5.0


Here is an example bash command using the Nicmanager provider:

```bash
## Login using email

NICMANAGER_API_EMAIL = "you@example.com" \
NICMANAGER_API_PASSWORD = "password" \

# Optionally, if your account has TOTP enabled, set the secret here
NICMANAGER_API_OTP = "long-secret" \

lego --email you@example.com --dns nicmanager --domains my.example.org run

## Login using account name + username

NICMANAGER_API_LOGIN = "myaccount" \
NICMANAGER_API_USERNAME = "myuser" \
NICMANAGER_API_PASSWORD = "password" \

# Optionally, if your account has TOTP enabled, set the secret here
NICMANAGER_API_OTP = "long-secret" \

lego --email you@example.com --dns nicmanager --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NICMANAGER_API_EMAIL` | Email-based login |
| `NICMANAGER_API_LOGIN` | Login, used for Username-based login |
| `NICMANAGER_API_PASSWORD` | Password, always required |
| `NICMANAGER_API_USERNAME` | Username, used for Username-based login |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NICMANAGER_API_MODE` | mode: 'anycast' or 'zone' (default: 'anycast') |
| `NICMANAGER_API_OTP` | TOTP Secret (optional) |
| `NICMANAGER_HTTP_TIMEOUT` | API request timeout |
| `NICMANAGER_POLLING_INTERVAL` | Time between DNS propagation check |
| `NICMANAGER_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `NICMANAGER_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).

## Description

You can log in using your account name + username or using your email address.
Optionally if TOTP is configured for your account, set `NICMANAGER_API_OTP`.



## More information

- [API documentation](https://api.nicmanager.com/docs/v1/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nicmanager/nicmanager.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
