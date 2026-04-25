---
title: "DNS-01 Challenge"
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 1
---

This guide explains how to get and renew a certificate with the DNS-01 challenge.

<!--more-->

lego comes with support for [many providers]({{% ref "dns#dns-providers" %}}),
and you need to pick the one where your domain's DNS settings are set up.
Typically, this is the registrar where you bought the domain, but in some cases this can be another third-party provider.

## Using a DNS provider

For this example, let's assume you have set up Cloudflare for your domain.

{{< tabs groupid="usage-examples" >}}
{{% tab title="Classic Way" %}}
Execute the following command:

```bash
CLOUDFLARE_EMAIL='you@example.com' \
CLOUDFLARE_API_KEY='yourprivatecloudflareapikey' \
lego run --dns cloudflare --domains 'example.org' --domains '*.example.org'
```

{{% /tab %}}
{{% tab title="With a Configuration File" %}}

Create a `.lego.yml` file with the following content:

```yaml
challenges:
  cf:
    dns:
      provider: cloudflare

certificates:
  foo:
    domains:
      - example.com
      - '*.example.com'
```

And execute:

```bash
CLOUDFLARE_EMAIL='you@example.com' \
CLOUDFLARE_API_KEY='yourprivatecloudflareapikey' \
lego
```

{{% /tab %}}
{{< /tabs >}}

## Configuration and Credentials

More information [here]({{% ref "dns#configuration-and-credentials" %}})

## Tips

{{% notice title="For a zone that has multiple SOAs" icon="info-circle" %}}

This can often be found where your DNS provider has a zone entry for an internal network (i.e., a corporate network, or home LAN) as well as the public internet.
In this case, point lego at an external authoritative server for the zone using the additional parameter `--dns.resolvers`.

{{< tabs groupid="usage-examples" >}}
{{% tab title="Classic Way" %}}
Execute the following command:

```bash
CLOUDFLARE_EMAIL='you@example.com' \
CLOUDFLARE_API_KEY='yourprivatecloudflareapikey' \
lego run --dns cloudflare --dns.resolvers 9.9.9.9:53 -d 'example.org' -d '*.example.org'
```

{{% /tab %}}
{{% tab title="With a Configuration File" %}}

Create a `.lego.yml` file with the following content:

```yaml
challenges:
  cf:
    dns:
      provider: cloudflare
      resolvers:
        - 9.9.9.9:53
certificates:
  foo:
    domains:
      - example.org
      - '*.example.org'
```

And execute:

```bash
CLOUDFLARE_EMAIL='you@example.com' \
CLOUDFLARE_API_KEY='yourprivatecloudflareapikey' \
lego
```

{{% /tab %}}
{{< /tabs >}}



[More information about resolvers.]({{% ref "advanced/tips#dns-resolvers-and-challenge-verification" %}})

{{% /notice %}}

