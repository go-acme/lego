---
title: "Rain Cloud DNS"
date: 2024-11-18T16:22:46+01:00
draft: false
slug: raindns
dnsprovider:
  since:    "v1.0.0"
  code:     "raindns"
  url:      "https://www.apifox.cn/apidoc/shared-a4595cc8-44c5-4678-a2a3-eed7738dab03"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/raindns/raindns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Rain Cloud DNS](https://www.apifox.cn/apidoc/shared-a4595cc8-44c5-4678-a2a3-eed7738dab03).


<!--more-->

- Code: `raindns`
- Since: v1.0.0


Here is an example bash command using the Rain Cloud DNS provider:

```bash
# Setup using credentials
RAIN_API_KEY=abcdefghijklmnopqrstuvwx \
lego --email you@example.com --dns raindns - -d '*.example.com' -d example.com run
```

## Credentials

| Environment Variable Name | Description |
|---------------------------|-------------|
| `RAIN_API_KEY`            | Access key ID |


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `RAIN_HTTP_TIMEOUT` | API request timeout |
| `RAIN_POLLING_INTERVAL` | Time between DNS propagation check |
| `RAIN_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `RAIN_TTL` | The TTL of the TXT record used for the DNS challenge |
