---
title: "Archive Operations"
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 7
---

This section describes operations on the archives.

<!--more-->

## List

You can list all the backuped accounts and certificates with the following command:

```bash
lego archives list
```

To know the available options, run:

```bash
lego archives list --help
```

Or read the [documentation]({{% ref "references/ref-flags/#lego-archives-list" %}}).

## Restore

You can restore a backup of an account or a certificate with the following command:

```bash
lego archives restore
```

The command will ask you for the backup file to restore.

To know the available options, run:

```bash
lego archives restore --help
```

Or read the [documentation]({{% ref "references/ref-flags/#lego-archives-restore" %}}).
