---
title: "Manual"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: manual
---

Solving the DNS-01 challenge using CLI prompt.

<!--more-->

## Example

```txt
Do you accept the TOS? Y/n

[INFO] acme: Registering account for test@test.com
!!!! HEADS UP !!!!

    Your account credentials have been saved in your Let's Encrypt
    configuration directory at "~/.lego/accounts".
    You should make a secure backup of this folder now. This
    configuration directory will also contain certificates and
    private keys obtained from Let's Encrypt so making regular
    backups of this folder is ideal.
    
[INFO] [test.com] acme: Obtaining bundled SAN certificate
[INFO] [test.com] AuthURL: https://acme-v02.api.letsencrypt.org/acme/authz/lornkZmVYjsh5wLHpxdQcZDPekGf_TYUM-MTJk3-yrA
[INFO] [test.com] acme: Could not find solver for: tls-alpn-01
[INFO] [test.com] acme: Could not find solver for: http-01
[INFO] [test.com] acme: use dns-01 solver
[INFO] [test.com] acme: Preparing to solve DNS-01
lego: Please create the following TXT record in your test.com. zone:
_acme-challenge.test.com. 120 IN TXT "VP-dby1RBuUOnDZg1n9sF-cwicLsognMzJb0Vx8ttAI"
lego: Press 'Enter' when you are done

Do you accept the TOS? Y/n

[INFO] acme: Registering account for test@test.com
!!!! HEADS UP !!!!

    Your account credentials have been saved in your Let's Encrypt
    configuration directory at "~/.lego/accounts".
    You should make a secure backup of this folder now. This
    configuration directory will also contain certificates and
    private keys obtained from Let's Encrypt so making regular
    backups of this folder is ideal.

[INFO] [test.com] acme: Obtaining bundled SAN certificate
[INFO] [test.com] AuthURL: https://acme-v02.api.letsencrypt.org/acme/authz/lornkZmVYjsh5wLHpxdQcZDPekGf_TYUM-MTJk3-yrA
[INFO] [test.com] acme: Could not find solver for: tls-alpn-01
[INFO] [test.com] acme: Could not find solver for: http-01
[INFO] [test.com] acme: use dns-01 solver
[INFO] [test.com] acme: Preparing to solve DNS-01
lego: Please create the following TXT record in your test.com. zone:
_acme-challenge.test.com. 120 IN TXT "VP-dby1RBuUOnDZg1n9sF-cwicLsognMzJb0Vx8ttAI"
lego: Press 'Enter' when you are done

```