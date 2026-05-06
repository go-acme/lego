---
title: 'Hooks'
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 1
---

This section describes how to use hooks.

<!--more-->

There are three hooks available:
- `pre-hook`
- `deploy-hook`
- `post-hook`

## Pre-Hook

The hook is executed only when the certificates are effectively renewed or created.

{{< tabs groupid="usage-examples" >}}
{{% tab title="Classic Way" %}}

Execute the following command:

```bash
lego run -d 'example.com' --pre-hook='./my-pre-hook.sh'
```

{{% /tab %}}
{{% tab title="With a Configuration File" %}}

Define the following section in your `.lego.yaml` file:

```yaml
hooks:
  pre:
    command: './my-pre-hook.sh'
```

{{% /tab %}}
{{< /tabs >}}

## Deploy-Hook

This hook is executed, before the creation or the renewal, in cases where a certificate will be effectively created/renewed.

{{< tabs groupid="usage-examples" >}}
{{% tab title="Classic Way" %}}

Execute the following command:

```bash
lego run -d 'example.com' --deploy-hook='./my-deploy-hook.sh'
```

{{% /tab %}}
{{% tab title="With a Configuration File" %}}

Define the following section in your `.lego.yaml` file:

```yaml
hooks:
  deploy:
    command: './my-deploy-hook.sh'
```

{{% /tab %}}
{{< /tabs >}}

## Post-Hook

This hook is executed, after the creation or the renewal, in cases where a certificate is created/renewed, regardless of whether any errors occurred.

{{< tabs groupid="usage-examples" >}}
{{% tab title="Classic Way" %}}

Execute the following command:

```bash
lego run -d 'example.com' --post-hook='./my-post-hook.sh'
```

{{% /tab %}}
{{% tab title="With a Configuration File" %}}

Define the following section in your `.lego.yaml` file:

```yaml
hooks:
  post:
    command: './my-post-hook.sh'
```

{{% /tab %}}
{{< /tabs >}}

## Environment Variables

Some details are passed through environment variables to help you with your hooks:

| Environment Variable            | Description                                          |
|---------------------------------|------------------------------------------------------|
| `LEGO_HOOK_ACCOUNT_ID`          | The account ID.                                      |
| `LEGO_HOOK_ACCOUNT_EMAIL`       | The account email (if available).                    |
| `LEGO_HOOK_ACCOUNT_SERVER`      | The server related to the account.                   |
| `LEGO_HOOK_CERT_NAME`           | The name/ID of the certificate.                      |
| `LEGO_HOOK_CERT_NAME_SANITIZED` | The sanitized name/ID of the certificate.            |
| `LEGO_HOOK_CERT_KEY_TYPE`       | The type of the certificate key.                     |
| `LEGO_HOOK_CERT_DOMAINS`        | The domains of the certificate.                      |
| `LEGO_HOOK_CERT_PATH`           | The path of the certificate.                         |
| `LEGO_HOOK_CERT_KEY_PATH`       | The path of the certificate key.                     |
| `LEGO_HOOK_ISSUER_CERT_PATH`    | The path of the issuer certificate.                  |
| `LEGO_HOOK_CERT_PEM_PATH`       | (only with `--pem`) The path to the PEM certificate. |
| `LEGO_HOOK_CERT_PFX_PATH`       | (only with `--pfx`) The path to the PFX certificate. |

## Use Case

A typical use case is distributing the certificate for other services and reload them if necessary.
Since many programs understand PEM-formatted TLS certificates, it is relatively simple to use certificates for more than a web server.

This example script installs the new certificate for a mail server and reloads it.
Beware: this is just a starting point, error checking is omitted for brevity.

```bash
#!/bin/bash

# copy certificates to a directory controlled by Postfix
postfix_cert_dir="/etc/postfix/certificates"

# our Postfix server only handles mail for @example.com domain
if [ "$LEGO_HOOK_CERT_NAME" = "example.com" ]; then
  install -u postfix -g postfix -m 0644 "$LEGO_HOOK_CERT_PATH" "$postfix_cert_dir"
  install -u postfix -g postfix -m 0640 "$LEGO_HOOK_CERT_KEY_PATH"  "$postfix_cert_dir"

  systemctl reload postfix@-service
fi
```
