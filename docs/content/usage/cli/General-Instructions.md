---
title: General Instructions
date: 2019-03-03T16:39:46+01:00
draft: false
summary: Read this first to clarify some assumptions made by the following guides.
weight: 1
---

These examples assume you have [lego installed]({{< ref "installation" >}}).
You can get a pre-built binary from the [releases](https://github.com/go-acme/lego/releases) page.

The web server examples require that the `lego` binary has permission to bind to ports 80 and 443.
If your environment does not allow you to bind to these ports, please read [Running without root privileges]({{< ref "usage/cli/Options#running-without-root-privileges" >}}) and [Port Usage]({{< ref "usage/cli/Options#port-usage" >}}).

Unless otherwise instructed with the `--path` command line flag, lego will look for a directory named `.lego` in the *current working directory*.
If you run `cd /dir/a && lego ... run`, lego will create a directory `/dir/a/.lego` where it will save account registration and certificate files into.
If you later try to renew a certificate with `cd /dir/b && lego ... renew`, lego will likely produce an error.
