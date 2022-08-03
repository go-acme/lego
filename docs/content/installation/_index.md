---
title: "Installation"
date: 2019-03-03T16:39:46+01:00
weight: 1
draft: false
---

## Binaries

To get the binary just download the latest release for your OS/Arch from [the release page](https://github.com/go-acme/lego/releases) and put the binary somewhere convenient.
lego does not assume anything about the location you run it from.

## From Docker

```bash
docker run goacme/lego -h
```

## From package managers

- [ArchLinux (AUR)](https://aur.archlinux.org/packages/lego) (official):

  ```bash
  yay -S lego
  ```

- [FreeBSD (Ports)](https://www.freshports.org/security/lego) (unofficial):

  ```bash
  cd /usr/ports/security/lego && make install clean
  ```

  or

  ```bash
  pkg install lego
  ```

## From sources

Requirements:

- go1.17+
- environment variable: `GO111MODULE=on`

To install the latest version from sources, just run:

```bash
go install github.com/go-acme/lego/v4/cmd/lego@latest
```

or

```bash
git clone git@github.com:go-acme/lego.git
cd lego
make        # tests + doc + build
make build  # only build
```
