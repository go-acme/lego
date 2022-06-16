---
title: "Joker"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: joker
dnsprovider:
  since:    "v2.6.0"
  code:     "joker"
  url:      "https://joker.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/joker/joker.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Joker](https://joker.com).


<!--more-->

- Code: `joker`
- Since: v2.6.0


Here is an example bash command using the Joker provider:

```bash
# SVC
JOKER_API_MODE=SVC \
JOKER_USERNAME=<your email> \
JOKER_PASSWORD=<your password> \
lego --email you@example.com --dns joker --domains my.example.org run

# DMAPI
JOKER_API_MODE=DMAPI \
JOKER_USERNAME=<your email> \
JOKER_PASSWORD=<your password> \
lego --email you@example.com --dns joker --domains my.example.org run
## or
JOKER_API_MODE=DMAPI \
JOKER_API_KEY=<your API key> \
lego --email you@example.com --dns joker --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `JOKER_API_KEY` | API key (only with DMAPI mode) |
| `JOKER_API_MODE` | 'DMAPI' or 'SVC'. DMAPI is for resellers accounts. (Default: DMAPI) |
| `JOKER_PASSWORD` | Joker.com password |
| `JOKER_USERNAME` | Joker.com username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `JOKER_HTTP_TIMEOUT` | API request timeout |
| `JOKER_POLLING_INTERVAL` | Time between DNS propagation check |
| `JOKER_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `JOKER_SEQUENCE_INTERVAL` | Time between sequential requests (only with 'SVC' mode) |
| `JOKER_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).

## SVC mode

In the SVC mode, username and passsword are not your email and account passwords, but those displayed in Joker.com domain dashboard when enabling Dynamic DNS.

As per [Joker.com documentation](https://joker.com/faq/content/6/496/en/let_s-encrypt-support.html):

> 1. please login at Joker.com, visit 'My Domains',
>    find the domain you want to add  Let's Encrypt certificate for, and chose "DNS" in the menu
>
> 2. on the top right, you will find the setting for 'Dynamic DNS'.
>    If not already active, please activate it.
>    It will not affect any other already existing DNS records of this domain.
>
> 3. please take a note of the credentials which are now shown as 'Dynamic DNS Authentication', consisting of a 'username' and a 'password'.
>
> 4. this is all you have to do here - and only once per domain.



## More information

- [API documentation](https://joker.com/faq/category/39/22-dmapi.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/joker/joker.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
