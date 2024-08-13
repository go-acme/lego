---
title: "Yandex Cloud"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: yandexcloud
dnsprovider:
  since:    "v4.9.0"
  code:     "yandexcloud"
  url:      "https://cloud.yandex.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/yandexcloud/yandexcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Yandex Cloud](https://cloud.yandex.com).


<!--more-->

- Code: `yandexcloud`
- Since: v4.9.0


Here is an example bash command using the Yandex Cloud provider:

```bash
YANDEX_CLOUD_IAM_TOKEN=<base64_IAM_token> \
YANDEX_CLOUD_FOLDER_ID=<folder/project_id> \
lego --email you@example.com --dns yandexcloud --domains "example.org" --domains "*.example.org" run

# ---

YANDEX_CLOUD_IAM_TOKEN=$(echo '{ \
  "id": "<string id>", \
  "service_account_id": "<string id>", \
  "created_at": "<datetime>", \
  "key_algorithm": "RSA_2048", \
  "public_key": "-----BEGIN PUBLIC KEY-----<rsa public key>-----END PUBLIC KEY-----", \
  "private_key": "-----BEGIN PRIVATE KEY-----<rsa private key>-----END PRIVATE KEY-----" \
}' | base64) \
YANDEX_CLOUD_FOLDER_ID=<yandex cloud folder(project) id> \
lego --email you@example.com --dns yandexcloud --domains "example.org" --domains "*.example.org" run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `YANDEX_CLOUD_FOLDER_ID` | The string id of folder (aka project) in Yandex Cloud |
| `YANDEX_CLOUD_IAM_TOKEN` | The base64 encoded json which contains information about iam token of service account with `dns.admin` permissions |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `YANDEX_CLOUD_POLLING_INTERVAL` | Time between DNS propagation check |
| `YANDEX_CLOUD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `YANDEX_CLOUD_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

## IAM Token

The simplest way to retrieve IAM access token is usage of yc-cli,
follow [docs](https://cloud.yandex.ru/docs/iam/operations/iam-token/create-for-sa) to get it

```bash
yc iam key create --service-account-name my-robot --output key.json
cat key.json | base64
```



## More information

- [API documentation](https://cloud.yandex.com/en/docs/dns/quickstart)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/yandexcloud/yandexcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
