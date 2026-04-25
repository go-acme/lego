---
title: "Installation"
date: 2019-03-03T16:39:46+01:00
weight: 1
draft: false
aliases:
  - installation
---

## Binaries

To get the binary, download the latest release for your OS/Arch from {{% button href=https://github.com/go-acme/lego/releases" icon="download" %}}the release page{{% /button %}} and put the binary somewhere convenient.

lego does not assume anything about the location you run it from.

{{% button href=https://github.com/go-acme/lego/releases" icon="download" style="primary" %}}Download{{% /button %}}

## From Docker

```bash
docker run goacme/lego -h
```

{{% button href="https://hub.docker.com/r/goacme/lego" icon="arrow-up-right-from-square" %}}Link to the Docker Hub{{% /button %}}

## From package managers

{{< tabs >}}
{{% tab title="Arch" %}}

```bash
pacman -S lego
```

{{% button href="https://archlinux.org/packages/extra/x86_64/lego/" icon="arrow-up-right-from-square" %}}Link to the package{{% /button %}}

{{% /tab %}}
{{% tab title="AUR" %}}

```bash
yay -S lego-bin
```

{{% button href="https://aur.archlinux.org/packages/lego-bin" icon="arrow-up-right-from-square" %}}Link to the package{{% /button %}}

{{% /tab %}}
{{% tab title="Snap" %}}

```bash
sudo snap install lego
```
Note: The snap can only write to the `/var/snap/lego/common/.lego` directory.

{{% button href="https://snapcraft.io/lego" icon="arrow-up-right-from-square" %}}Link to the package{{% /button %}}

{{% /tab %}}
{{% tab title="FreeBSD (Ports)" %}}

```bash
pkg install lego
```

{{% button href="https://www.freshports.org/security/lego" icon="arrow-up-right-from-square" %}}Link to the package{{% /button %}}

{{% /tab %}}
{{% tab title="Gentoo" %}}

You can [enable GURU](https://wiki.gentoo.org/wiki/Project:GURU/Information_for_End_Users) repository and then:

```bash
emerge app-crypt/lego
```

{{% button href="https://gitweb.gentoo.org/repo/proj/guru.git/tree/app-crypt/lego" icon="arrow-up-right-from-square" %}}Link to the package{{% /button %}}

{{% /tab %}}
{{% tab title="Homebrew" %}}

```bash
brew install lego
```

or

```bash
pkg install lego
```

{{% button href="https://formulae.brew.sh/formula/lego" icon="arrow-up-right-from-square" %}}Link to the package{{% /button %}}

{{% /tab %}}
{{% tab title="OpenBSD (Ports)" %}}

```bash
pkg_add lego
```

{{% button href="https://openports.pl/path/security/lego" icon="arrow-up-right-from-square" %}}Link to the package{{% /button %}}

{{% /tab %}}
{{< /tabs >}}

The lego maintainers are only maintaining the AUR and Snap packages, the other packages are community maintained.

## From sources

Requirements:

- go1.25+.

To install the latest version from sources, just run:

```bash
go install github.com/go-acme/lego/v5@latest
```

or

```bash
git clone git@github.com:go-acme/lego.git
cd lego
make        # tests + doc + build
make build  # only build
```
