---
title: "File Configuration"
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 5
---

The configuration file is a way to simplify the management of multiple certificates.

<!--more-->

## Commands

The configuration file is used by the following commands:

- `lego`
- `lego certificates revoke`
- `lego certificates list`
- `lego accounts list`
- `lego archives list`
- `lego archives restore`

## File Location and Format

The configuration file is a YAML file named `.lego.yml` (or `.lego.yaml`) placed in the current working directory.
lego will automatically find and use it when present.

You can also pass a custom path with the `--config` flag.

## Configuration File Structure

The configuration file is organized in a way that makes it easy to understand and modify.

The four main sections (`servers`, `accounts`, `challenges`, and `certificates`) are named maps:
each entry has a key (a name you choose) and a value (its configuration).

Certificates reference accounts and challenges by their name, and accounts reference servers by their name.

More information about the configuration file structure can be found in the [configuration file reference]({{% ref "references/ref-file" %}}).

## Smart Defaults

The configuration file applies a number of defaults to reduce verbosity:

| Setting                           | Description                                                                                                 |
|-----------------------------------|-------------------------------------------------------------------------------------------------------------|
| Storage                           | Defaults to `.lego` in the current directory.                                                               |
| Account server                    | Defaults to the Let's Encrypt production if not specified.                                                  |
| Certificate key type              | Inherits from its account if not specified.                                                                 |
| Certificate account               | If there is only one account defined, it is used automatically.                                             |
| Certificate challenge             | If there is only one challenge defined, it is used automatically.                                           |
| Certificate predefined challenges | `http-01`, `tls-alpn-01`, `dns-persist-01` can be used as a challenge name without dedicated configuration. |

This means the minimal configuration to obtain a certificate is just a certificate entry (and a challenge for DNS-01):

{{< tabs >}}
{{% tab title="HTTP-01" %}}

Minimal example for a certificate (Let's Encrypt).

```yaml
# .lego.yml
certificates:
  my-cert:
    challenge: http-01
    domains:
      - example.com
```

{{% /tab %}}
{{% tab title="TLS-ALPN-01" %}}

Minimal example for a certificate (Let's Encrypt).

```yaml
# .lego.yml
certificates:
  my-cert:
    challenge: tls-alpn-01
    domains:
      - example.com
```

{{% /tab %}}
{{% tab title="DNS-01" %}}

Minimal example for a wildcard certificate (Let's Encrypt and DNS-01 via Cloudflare).

```yaml
# .lego.yml
challenges:
  my-dns:
    dns:
      provider: cloudflare

certificates:
  my-cert:
    domains:
      - example.com
      - '*.example.com'
```

{{% /tab %}}
{{% tab title="DNS-PERSIST-01" %}}

Please read the dedicated [documentation]({{% ref "obtain/dnspersist01" %}}).

{{% /tab %}}
{{< /tabs >}}

## Archive Behavior

The configuration file drives lifecycle management:

- When a certificate entry is removed, its files are archived.
- When an account entry is removed, its files are archived.
- When a server entry is removed, the server and its related accounts are archived.

More information about commands related to archives can be found in the [archives section]({{% ref "advanced/archives" %}}).
