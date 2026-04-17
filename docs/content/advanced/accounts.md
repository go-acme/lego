---
title: "Account Operations"
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 3
---

This section describes account operations.

<!--more-->

## List Accounts

You can list all the accounts registered in your local storage:

```bash
lego accounts list
```

Output:

```
Found the following accounts:
noemail@example.com
├── Email: 
├── Server: https://acme-v02.api.letsencrypt.org/directory
├── Key Type: EC256
└── Path: /path/to/.lego/accounts/acme-v02.api.letsencrypt.org/noemail@example.com/account.json

...

```

## Register

You can register a new account by using the following command:

{{< tabs >}}
{{% tab title="Simple Registration" %}}

```bash
lego accounts register --account-id='myaccount'
```

{{% /tab %}}
{{% tab title="With External Account Binding" %}}

```bash
lego accounts register --account-id='myaccount' \
    --server https://example.com/ca \
    --eab \
    --eab.kid xxx \
    --eab.hmac yyy
```

{{% /tab %}}
{{< /tabs >}}

## Key Rollover

You can change the account private key (a key rollover) by using the following command:

```bash
lego accounts keyrollover --account-id='myaccount'
```

## Account Recovery

You can recover/import an account, to do that, you need the private key of the account.

```bash
lego account recover --account-id='myaccount' --private-key /path/to/private-key.pem
```

The account will be imported and added to `.lego/accounts/`.
