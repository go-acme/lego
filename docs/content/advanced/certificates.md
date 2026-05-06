---
title: "Certificate Operations"
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 2
---

This section describes certificate operations.

<!--more-->

## List Certificates

You can list all certificates managed by lego with:

```bash
lego certificates list
```

Output:

```
Found the following certificates:
_.example.com
├── Status: this certificate is expired.
├── Domains: *.example.com, example.com
├── Expiration Date: 2026-04-08 21:02:27 +0000 UTC
├── Issuer: CN=(STAGING) Puzzling Parsnip E7,O=(STAGING) Let's Encrypt,C=US
└── Certificate Path: /path/to/.lego/certificates/_.example.com.crt

...
```

## Revoke Certificates

You can revoke existing certificates.

{{< tabs groupid="usage-examples" >}}
{{% tab title="Classic Way" %}}

Execute the following command:

```bash
lego certificates revoke --cert-name 'example.com'
```

{{% /tab %}}
{{% tab title="With a Configuration File" %}}

If you have the following `.lego.yml` configuration file:

```yaml
certificates:
  foo:
    challenge: http-01
    domains:
      - example.com
```

And execute:

```bash
lego certificates revoke --cert-name foo
```

When using a configuration file, you can revoke all certificates at once:

```bash
lego certificates revoke
```

{{% /tab %}}
{{< /tabs >}}

