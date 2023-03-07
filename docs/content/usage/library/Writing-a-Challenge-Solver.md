---
title: "Writing a Challenge Solver"
date: 2019-03-03T16:39:46+01:00
draft: false
---

Lego can solve multiple ACME challenge types out of the box, but sometimes you have custom requirements.

<!--more-->

For example, you may want to write a solver for the DNS-01 challenge that works with a different DNS provider (lego already supports CloudFlare, AWS, DigitalOcean, and others).

The DNS-01 challenge is advantageous when other challenge types are impossible.
For example, the HTTP-01 challenge doesn't work well behind a load balancer or CDN and the TLS-ALPN-01 challenge breaks behind TLS termination.

But even if using HTTP-01 or TLS-ALPN-01 challenges, you may have specific needs that lego does not consider by default.

You can write something called a `challenge.Provider` that implements [this interface](https://pkg.go.dev/github.com/go-acme/lego/v4/challenge#Provider):

```go
type Provider interface {
	Present(domain, token, keyAuth string) error
	CleanUp(domain, token, keyAuth string) error
}
```

This provides the means to solve a challenge.
First you present a token to the ACME server in a way defined by the challenge type you're solving for, then you "clean up" after the challenge finishes.

## Writing a challenge.Provider

Pretend we want to write our own DNS-01 challenge provider (other challenge types have different requirements but the same principles apply).

This will let us prove ownership of domain names parked at a new, imaginary DNS service called BestDNS without having to start our own HTTP server.
BestDNS has an API that, given an authentication token, allows us to manipulate DNS records.

This simplistic example has only one field to store the auth token, but in reality you may need to keep more state.

```go
type DNSProviderBestDNS struct {
	apiAuthToken string
}
```

We should provide a constructor that returns a *pointer* to the `struct`.
This is important in case we need to maintain state in the `struct`.

```go
func NewDNSProviderBestDNS(apiAuthToken string) (*DNSProviderBestDNS, error) {
	return &DNSProviderBestDNS{apiAuthToken: apiAuthToken}, nil
}
```

Now we need to implement the interface.
We'll start with the `Present` method.
You'll be passed the `domain` name for which you're proving ownership, a `token`, and a `keyAuth` string.
How your provider uses `token` and `keyAuth`, or if you even use them at all, depends on the challenge type.
For DNS-01, we'll just use `domain` and `keyAuth`.

```go
func (d *DNSProviderBestDNS) Present(domain, token, keyAuth string) error {
    info := dns01.GetChallengeInfo(domain, keyAuth)
    // make API request to set a TXT record on fqdn with value and TTL
    return nil
}
```

After calling `dns01.GetChallengeInfo(domain, keyAuth)`, we now have the information we need to make our API request and set the TXT record:
- `FQDN` is the fully qualified domain name on which to set the TXT record.
- `EffectiveFQDN` is the fully qualified domain name after the CNAMEs resolutions on which to set the TXT record.
- `Value` is the record's value to set on the record.

So then you make an API request to the DNS service according to their docs.
Once the TXT record is set on the domain, you may return and the challenge will proceed.

The ACME server will then verify that you did what it required you to do, and once it is finished, lego will call your `CleanUp` method.
In our case, we want to remove the TXT record we just created.

```go
func (d *DNSProviderBestDNS) CleanUp(domain, token, keyAuth string) error {
    // clean up any state you created in Present, like removing the TXT record
}
```

In our case, we'd just make another API request to have the DNS record deleted; no need to keep it and clutter the zone file.

## Using your new challenge.Provider

To use your new challenge provider, call [`client.Challenge.SetDNS01Provider`](https://pkg.go.dev/github.com/go-acme/lego/v4/challenge/resolver#SolverManager.SetDNS01Provider) to tell lego, "For this challenge, use this provider".
In our case:

```go
bestDNS, err := NewDNSProviderBestDNS("my-auth-token")
if err != nil {
    return err
}

client.Challenge.SetDNS01Provider(bestDNS)
```

Then, when this client tries to solve the DNS-01 challenge, it will use our new provider, which sets TXT records on a domain name hosted by BestDNS.

That's really all there is to it.
Go make awesome things!
