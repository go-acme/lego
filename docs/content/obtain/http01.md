---
title: "HTTP-01 Challenge"
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 1
---

This guide explains how to get and renew a certificate with the HTTP-01 challenge.

<!--more-->

{{% notice note %}}
The examples require that the `lego` binary has permission to bind to ports 80.  
If your environment does not allow you to bind to these ports, please read [Running without root privileges]({{% ref "advanced/tips#running-without-root-privileges" %}}) and [Port Usage]({{% ref "advanced/tips/#port-usage" %}}).
{{% /notice %}}

## Using the built-in web server

{{< tabs groupid="usage-examples" >}}
{{% tab title="Classic Way" %}}

Execute the following command:

```bash
lego run -d 'example.com' --http
```

{{% /tab %}}
{{% tab title="With a Configuration File" %}}

Create a `.lego.yml` file with the following content:

```yaml
certificates:
  foo:
    challenge: http-01
    domains:
      - example.com
```

And execute:

```bash
lego
```

{{% /tab %}}
{{< /tabs >}}

## Using an existing, running web server

If you have an existing server running on port 80, the `--http` option also requires the `--http.webroot` option.
This just writes the http-01 challenge token to the given directory in the folder `.well-known/acme-challenge` and does not start a server.

The given directory **should** be publicly served as `/` on the domain(s) for the validation to complete.

If the given directory is not publicly served, you will have to support rewriting the request to the directory;

You could also implement a rewrite to rewrite `.well-known/acme-challenge` to the given directory `.well-known/acme-challenge`.

You should be able to run an existing webserver on port 80 and have lego write the token file with the HTTP-01 challenge key authorization to `<webroot dir>/.well-known/acme-challenge/` by running something like:

{{< tabs groupid="usage-examples" >}}
{{% tab title="Classic Way" %}}

Execute the following command:

```bash
lego run --http --http.webroot /path/to/webroot --domains example.com
```

{{% /tab %}}
{{% tab title="With a Configuration File" %}}

Create a `.lego.yml` file with the following content:

```yaml
challenges:
  mychallenge:
    http:
      webroot: /tmp/webroot

certificates:
  foo:
    challenge: mychallenge
    domains:
      - example.com
```

And execute:

```bash
lego
```

{{% /tab %}}
{{< /tabs >}}
