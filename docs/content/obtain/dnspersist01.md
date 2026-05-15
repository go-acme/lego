---
title: "DNS-PERSIST-01 Challenge"
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 4
---

This guide explains how to get and renew a certificate with the DNS-PERSIST-01 challenge.

<!--more-->

{{% notice note %}}
- The [RFC](https://datatracker.ietf.org/doc/draft-ietf-acme-dns-persist/) is still a draft.
- This is currently not available in most CA production.
{{% /notice %}}

{{% notice important %}}
This challenge could be less secure than [DNS-01]({{% ref "obtain/dns01" %}}) due to its requirements.

This is especially true if your DNS provider does not offer any way to limit the access controls to the specific persistent record required by the DNS-PERSIST-01 challenge.

- [7. Security Considerations](https://www.ietf.org/archive/id/draft-ietf-acme-dns-persist-01.html#section-7)
- [9.5. DNS Provider Considerations](https://www.ietf.org/archive/id/draft-ietf-acme-dns-persist-01.html#section-9.5)

The security of this challenge relies primarily on protecting your account's private key.

{{% /notice %}}

{{< tabs groupid="usage-examples" >}}
{{% tab title="Classic Way" %}}

Execute the following command:

```bash
lego run -d 'example.com' --dns-persist
```

To know the available options, read the [documentation]({{% ref "references/ref-flags/#llego-run" %}}).

{{% /tab %}}
{{% tab title="With a Configuration File" %}}

Create a `.lego.yml` file with the following content:

```yaml
certificates:
  foo:
    challenge: dns-persist-01
    domains:
      - example.com
```

And execute:

```bash
lego
```

To know the available options, read the [documentation]({{% ref "references/ref-file/#challenges" %}}).

{{% /tab %}}
{{< /tabs >}}
