---
title: "File Configuration"
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 5
---

The configuration file is a way to simplify the management of multiple certificates.

<!--more-->

## File Location and Format

The configuration file is a YAML file named `.lego.yml` (or `.lego.yaml`) placed in the current working directory.
lego will automatically find and use it when present.

You can also pass a custom path with the `--config` flag.

## Configuration File Structure

The configuration file is organized in a way that makes it easy to understand and modify.

The four main sections (`servers`, `accounts`, `challenges`, and `certificates`) are named maps:
each entry has a key (a name you choose) and a value (its configuration).

Certificates reference accounts and challenges by their name, and accounts reference servers by their name.

More information about the configuration file structure can be found in the [configuration file structure]({{% ref "references/ref-file" %}}).

## Smart Defaults

The configuration file applies a number of defaults to reduce verbosity:

| Setting                | Description                                                       |
|------------------------|-------------------------------------------------------------------|
| Storage                | Defaults to `.lego` in the current directory.                     |
| Account server         | Defaults to the Let's Encrypt production if not specified.        |
| Certificate key type   | Inherits from its account if not specified.                       |
| Certificate account    | If there is only one account defined, it is used automatically.   |
| Certificate challenge  | If there is only one challenge defined, it is used automatically. |

This means the minimal configuration to obtain a certificate is just a challenge and a certificate entry:

```yaml
# .lego.yml
# Minimal example for a wildcard certificate (Let's Encrypt and DNS-01 via Cloudflare).
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

## Archive Behavior

The configuration file drives lifecycle management:

- When a certificate entry is removed, its files are archived.
- When an account entry is removed, its files are archived.
- When a server entry is removed, the server and its related accounts are archived.

More information about commands related to archives can be found in the [archives section]({{% ref "advanced/archives" %}}).
