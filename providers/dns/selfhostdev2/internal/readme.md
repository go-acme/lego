# SelfHost.(de|eu)

SelfHost doesn't provide public API documentation.

selfHOST offers a dedicated DNS API for ACME DNS-01 challenges.

It uses an API key per domain.

Keys can be generated on their customer portal.

- **Endpoint:** `POST https://my.selfhost.de/cgi-bin/dns-api.pl`
- **Content-Type:** `application/json`
- **Request body:**
  ```json
  {
    "api_key": "<key_id>.<secret>",
    "action": "present",
    "record_id": 123456,
    "content": "<43-char DNS-01 TXT value>"
  }
    ```
- `action`: `present` or `cleanup`
- `record_id`: numeric ID of an existing TXT record (JSON number, not a string)
- `content`: the DNS-01 challange
- **API key format:** `<key_id>.<secret>` – `key_id` is numeric (no leading zero), `secret` is 64 lowercase hex characters
- **Success:** HTTP `202 Accepted` (the request is accepted; DNS propagation happens asynchronously)

https://www.selfhost.de/faq-support/wie-kann-ich-die-acme-dns-api-fuer-zertifikate-nutzen/
