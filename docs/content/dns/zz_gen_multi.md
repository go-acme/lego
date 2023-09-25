---
title: "Multi"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: multi
dnsprovider:
  since:    "v4.14.1"
  code:     "multi"
  url:      ""
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/multi/multi.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Put the DNS challenge to multiple providers


<!--more-->

- Code: `multi`
- Since: v4.14.1


Here is an example bash command using the Multi provider:

```bash
AWS_ACCESS_KEY_ID=your_key_id \
AWS_SECRET_ACCESS_KEY=your_secret_access_key \
AWS_REGION=aws-region \
AWS_HOSTED_ZONE_ID=your_hosted_zone_id \
GOOGLE_DOMAINS_ACCESS_TOKEN=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
lego --email you@example.com --dns route53 --dns googledomains --domains my.example.org run
```





# Motivation

This is so that companies that use DNS split authority for
high-availability disaster-recover reasons, such as GitHub with OctoDNS
or Stack Overflow with dnscontrol can continue to request new
certificates with this library without risking the ACME server
querying the wrong nameserver.

## Usage

The simplest way to use this is to use the new `NewDNSProviderByNames` constructor

    import "github.com/go-acme/lego/v4/providers/dns/multi"
    provider, err = multi.NewDNSProviderByNames("route53", "googledomains")

If you need to use custom configurations, you can merge two created
`DNSProviders` with `NewDNSProviderFromOthers()`

    import (
        "github.com/go-acme/lego/v4/challenge"
        "github.com/go-acme/lego/v4/providers/dns"
        "github.com/go-acme/lego/v4/providers/dns/googledomains"
        "github.com/go-acme/lego/v4/providers/dns/multi"
        "github.com/go-acme/lego/v4/providers/dns/route53"
    )

    var r53Provider challenge.Provider
    if myRoleARN != "" {
        r53config := route53.NewDefaultConfig()
        r53config.AssumeRoleArn = myRoleARN

        r53Provider, _ = route53.NewDNSProviderConfig(r53config)
    } else {
        r53Provider, _ = dns.NewDNSChallengeProviderByName("route53")
    }
    googleProvider, _ := dns.NewDNSChallengeProviderByName("googledomains")
    multiProvider := multi.NewDNSProviderFromOthers(r53Provider, googleProvider)



## More information



<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/multi/multi.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
