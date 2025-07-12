# dnshome.de API

## Add TXT record

```
https://<subdomain>:<subdomain_password>@www.dnshome.de/dyndns.php?acme=add&txt=<txtvalue>
```

- `<subdomain>`: the subdomain (ex: `lego.dnshome.de`).
- `<subdomain_password>`: the subdomain password.
- `<txtvalue>`: the value of the TXT record (12 characters minimum)

Only one TXT record can be used for a subdomain.

Always returns StatusOK (200)

If the API call works the first word of the response body is `successfully`.

If an error occurs the response body is `error - <ERRMSG>`.

Can be a POST or a GET.

## Remove TXT record

```
https://<subdomain>:<subdomain_password>@www.dnshome.de/dyndns.php?acme=rm
```

- `<subdomain>`: the subdomain (ex: `lego.dnshome.de`).
- `<subdomain_password>`: the subdomain password.

Only one TXT record can be used for a subdomain.

Always returns StatusOK (200)

If the API call works the first word of the response body is `successfully`.

If an error occurs the response body is `error - <ERRMSG>`.

Can be a POST or a GET.
