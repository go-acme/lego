---
title: "Gigahost"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: gigahostno
dnsprovider:
  since:    "v4.21.0"
  code:     "gigahostno"
  url:      "https://www.gigahost.no/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gigahostno/gigahostno.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Gigahost](https://www.gigahost.no/).


<!--more-->

- Code: `gigahostno`
- Since: v4.21.0


Here is an example bash command using the Gigahost provider:

```bash
# Using username/password (recommended for automatic token refresh)
GIGAHOSTNO_USERNAME=user@example.com \
GIGAHOSTNO_PASSWORD=yourpassword \
lego --email you@example.com --dns gigahostno -d '*.example.com' -d example.com run

# Using pre-generated token (alternative method)
GIGAHOSTNO_TOKEN=your-api-token \
lego --email you@example.com --dns gigahostno -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `GIGAHOSTNO_PASSWORD` | Password - required for username/password auth |
| `GIGAHOSTNO_TOKEN` | Pre-generated API token - alternative to username/password |
| `GIGAHOSTNO_USERNAME` | Username (email) - required for username/password auth |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `GIGAHOSTNO_2FA_CODE` | Two-factor authentication code (optional, not recommended for automated use) |
| `GIGAHOSTNO_HTTP_TIMEOUT` | API request timeout |
| `GIGAHOSTNO_POLLING_INTERVAL` | Time between DNS propagation check |
| `GIGAHOSTNO_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `GIGAHOSTNO_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

## Authentication Methods

Gigahost supports two authentication methods:

### 1. Username/Password Authentication (Recommended)

This is the **recommended method** for automated certificate management. Lego will:
- Automatically obtain API tokens using your credentials
- Cache tokens for their ~90 day lifetime
- Automatically refresh tokens when they expire

```bash
GIGAHOSTNO_USERNAME=user@example.com
GIGAHOSTNO_PASSWORD=yourpassword
```

### 2. Pre-generated Token Authentication

You can provide a pre-generated API token directly. This method:
- Skips the authentication step
- Does **not** cache or refresh the token automatically
- Requires manual token renewal when the token expires (~90 days)

```bash
GIGAHOSTNO_TOKEN=your-api-token
```

Use this method only if you have a specific need to manage tokens externally. For most use cases, username/password authentication is simpler and more reliable.

## Two-Factor Authentication (2FA)

**WARNING:** It is strongly recommended to **disable 2FA** for accounts used with lego.

While 2FA is supported via the `GIGAHOSTNO_2FA_CODE` environment variable, it is **not suitable for automated certificate renewal** because:

- TOTP codes expire after 30 seconds
- API tokens have a ~90 day lifetime
- When the cached token expires, a fresh 2FA code is required for re-authentication
- This means manual intervention is needed every ~90 days

For automated certificate management, create a dedicated Gigahost account without 2FA enabled, or disable 2FA on the account you plan to use with lego.



## More information

- [API documentation](https://gigahost.no/api-dokumentasjon)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gigahostno/gigahostno.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
