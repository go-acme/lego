# Notes

## Forum

- https://derak.cloud/faq/programming/%da%86%da%af%d9%88%d9%86%d9%87-%d9%85%db%8c%d8%aa%d9%88%d8%a7%d9%86-%d8%a8%d9%87-api%d9%87%d8%a7-%d8%af%d8%b3%d8%aa%d8%b1%d8%b3%db%8c-%d8%af%d8%a7%d8%b4%d8%aa%d8%9f/
- https://derak.cloud/faq/programming/%d8%af%d8%b1%db%8c%d8%a7%d9%81%d8%aa-%da%a9%d9%84%db%8c%d8%af-api-api-key/

---

## DNS records (API)

### GET: Get a list of all DNS records

ex: `https://api.derak.cloud/v1.0/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords`

#### Query

| The name of the parameter | Description                      |
|---------------------------|----------------------------------|
| dnsType                   | dnsType query                    |
| content                   | The Host value of the DNS record |

#### Errors

| type error        | Error code |
|-------------------|------------|
| ForbiddenError    | 1003       |
| RateLimitExceeded | 1013       |


#### Example

```bash
curl -X GET --user "api:YOUR_API_KEY" \
https://api.derak.cloud/v1.0/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords
```
```bash
curl -X GET --user "api:api-MbmnxdpIBvk14nk5LFFdG1CV9PdMDfqi3tZAixBZLXYzM3qc187d7ede2de" \
https://api.derak.cloud/v1.0/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords \
-F dnsType="TXT" 
```


### PUT: Creating a new DNS record on the desired website

ex: `https://api.derak.cloud/v1.0/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords`

#### parameters

| The name of the parameter | Description                                                                                                                                                                  |
|---------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| *type                     | DNS record type [of types Aand AAAAand CNAMEand MXand NSand CAAand TXTand SPFand PTRand SRV]                                                                                 |
| *host                     | The Host value of the DNS record                                                                                                                                             |
| *content                  | The Host value of the DNS record                                                                                                                                             |
| ttl                       | TTL of DNS record [default: 0]                                                                                                                                               |
| cloud                     | This parameter specifies whether the traffic of this record passes through the cloud or not [Default: false]                                                                 |
| priority                  | Priority of MX and SRV records [Default: 0]                                                                                                                                  |
| service                   | SRV record service                                                                                                                                                           |
| protocol                  | SRV record protocol [default: _tcp]                                                                                                                                          |
| weight                    | SRV Record Weight [Default: 0]                                                                                                                                               |
| port                      | Priority of MX and SRV records [Default: 0]                                                                                                                                  |
| advanced                  | This parameter specifies whether this record has advanced settings or not [default: false]                                                                                   |
| upstreamPort              | Upstream Port of DNS record [Default: 80]                                                                                                                                    |
| upstreamProtocol          | Upstream protocol related to DNS records. Note that if you change these settings for another record of the same subdomain, the settings will be overwritten. [Default: http] |
| customSSLType             | Custom SSL related DNS record. Note that if you change these settings for another record of the same subdomain, the settings will be overwritten.                            |

#### Errors

| type error         | Error code |
|--------------------|------------|
| ForbiddenError     | 1003       |
| RateLimitExceeded  | 1013       |
| DNSValidationError | 1008       |

#### Example

```bash
curl -X PUT --user "api:YOUR_API_KEY" \
https://api.derak.cloud/v1.0/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords  \
-F type="A"  \
-F host="app"  \
-F content="1.2.3.4"
```

### GET: Get the information of a single DNS record

ex: `https://api.derak.cloud/v1.0/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords/:recordId`

#### Errors

| type error          | Error code |
|---------------------|------------|
| ForbiddenError      | 1003       |
| RateLimitExceeded   | 1013       |
| RecordNotFoundError | 1021       |

#### Example

```bash
curl -X GET --user "api:YOUR_API_KEY" \
https://api.derak.cloud/v1.0/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords/:recordId
```

### PATCH: Edit the parameters of a DNS record

`https://api.derak.cloud/v1.0/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords/:recordId`

#### parameters

| The name of the parameter | Description                                                                                                                                                                  |
|---------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| type                      | DNS record type [of types Aand AAAAand CNAMEand MXand NSand CAAand TXTand SPFand PTRand SRV]                                                                                 |
| host                      | The Host value of the DNS record                                                                                                                                             |
| content                   | The Host value of the DNS record                                                                                                                                             |
| ttl                       | TTL of DNS record [default: 0]                                                                                                                                               |
| cloud                     | This parameter specifies whether the traffic of this record passes through the cloud or not [Default: false]                                                                 |
| priority                  | Priority of MX and SRV records [Default: 0]                                                                                                                                  |
| service                   | SRV record service                                                                                                                                                           |
| protocol                  | SRV record protocol [default: _tcp]                                                                                                                                          |
| weight                    | SRV Record Weight [Default: 0]                                                                                                                                               |
| port                      | Priority of MX and SRV records [Default: 0]                                                                                                                                  |
| advanced                  | This parameter specifies whether this record has advanced settings or not [default: false]                                                                                   |
| upstreamPort              | Upstream Port of DNS record [Default: 80]                                                                                                                                    |
| upstreamProtocol          | Upstream protocol related to DNS records. Note that if you change these settings for another record of the same subdomain, the settings will be overwritten. [Default: http] |
| customSSLType             | Custom SSL related DNS record. Note that if you change these settings for another record of the same subdomain, the settings will be overwritten.                            |

#### Errors

| type error          | Error code |
|---------------------|------------|
| ForbiddenError      | 1003       |
| RateLimitExceeded   | 1013       |
| RecordNotFoundError | 1021       |
| DNSValidationError  | 1008       |

#### Example

```bash
curl -X PATCH --user "api:YOUR_API_KEY" \
https://api.derak.cloud/v1.0/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords/:recordId  \
-F cloud="true"
```

### DELETE: Delete a DNS record

ex: `https://api.derak.cloud/v1.0/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords/:recordId`

#### Errors

| type error          | Error code |
|---------------------|------------|
| ForbiddenError      | 1003       |
| RateLimitExceeded   | 1013       |
| RecordNotFoundError | 1021       |

#### Example

```bash
curl -X DELETE --user "api:YOUR_API_KEY" \
https://api.derak.cloud/v1.0/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords/:recordId
```

---

## Cache clearing (API)

### POST: Clearing (Purge Cache) specified parameters, if no parameter is specified, the entire cache is deleted.

ex: `https://api.derak.cloud/v1.0/zones/47c0ecf6c91243308c649ad1d2d618dd/cache/purge`

#### parameters

| The name of the parameter | Description                         |
|---------------------------|-------------------------------------|
| hostname                  | The hostname to be deleted          |
| hostnames                 | An array of hostnames to be cleared |
| url                       | The URL to be deleted               |
| urls                      | An array of URLs to be purged       |

#### Errors

| type error        | Error code |
|-------------------|------------|
| ForbiddenError    | 1003       |
| RateLimitExceeded | 1013       |

#### Examples

Purge URLS:

```bash
curl -X POST --user "api:YOUR_API_KEY" \
https://api.derak.cloud/v1.0/zones/47c0ecf6c91243308c649ad1d2d618dd/cache/purge  \
-F urls[]="https://www.derak.cloud/post/1"  \
-F urls[]="https://www.derak.cloud/post/2"
```

Purge HOSTNAMES:

```bash
curl -X POST --user "api:YOUR_API_KEY" \
https://api.derak.cloud/v1.0/zones/47c0ecf6c91243308c649ad1d2d618dd/cache/purge  \
-F hostnames[]="www.derak.cloud"  \
-F hostnames[]="app.derak.cloud"
```

Purge EVERYTHING:

```bash
curl -X POST --user "api:YOUR_API_KEY" \
https://api.derak.cloud/v1.0/zones/47c0ecf6c91243308c649ad1d2d618dd/cache/purge
```

---

## API for SSL certificates

### PUT: Enable SSL for a domain

ex: `https://api.derak.cloud/v1.0/zones/47c0ecf6c91243308c649ad1d2d618dd/ssl/`

#### Errors

| type error     | Error code |
|----------------|------------|
| ForbiddenError | 1003       |

#### Example

```bash
curl -X PUT --user "api:YOUR_API_KEY" \
https://api.derak.cloud/v1.0/zones/47c0ecf6c91243308c649ad1d2d618dd/ssl/
```

### DELETE: Disable SSL for a domain

ex: `https://api.derak.cloud/v1.0/zones/47c0ecf6c91243308c649ad1d2d618dd/ssl/`

#### Errors

| type error     | Error code |
|----------------|------------|
| ForbiddenError | 1003       |

#### Example

```bash
curl -X DELETE --user "api:YOUR_API_KEY" \
https://api.derak.cloud/v1.0/zones/47c0ecf6c91243308c649ad1d2d618dd/ssl/
```
