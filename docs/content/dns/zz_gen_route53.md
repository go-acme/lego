---
title: "Amazon Route 53"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: route53
dnsprovider:
  since:    "v0.3.0"
  code:     "route53"
  url:      "https://aws.amazon.com/route53/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/route53/route53.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Amazon Route 53](https://aws.amazon.com/route53/).


<!--more-->

- Code: `route53`
- Since: v0.3.0


{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `AWS_ACCESS_KEY_ID` | Managed by the AWS client. Access key ID (`AWS_ACCESS_KEY_ID_FILE` is not supported, use `AWS_SHARED_CREDENTIALS_FILE` instead) |
| `AWS_ASSUME_ROLE_ARN` | Managed by the AWS Role ARN (`AWS_ASSUME_ROLE_ARN` is not supported) |
| `AWS_HOSTED_ZONE_ID` | Override the hosted zone ID. |
| `AWS_PROFILE` | Managed by the AWS client (`AWS_PROFILE_FILE` is not supported) |
| `AWS_REGION` | Managed by the AWS client (`AWS_REGION_FILE` is not supported) |
| `AWS_SDK_LOAD_CONFIG` | Managed by the AWS client. Retrieve the region from the CLI config file (`AWS_SDK_LOAD_CONFIG_FILE` is not supported) |
| `AWS_SECRET_ACCESS_KEY` | Managed by the AWS client. Secret access key (`AWS_SECRET_ACCESS_KEY_FILE` is not supported, use `AWS_SHARED_CREDENTIALS_FILE` instead) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `AWS_MAX_RETRIES` | The number of maximum returns the service will use to make an individual API request |
| `AWS_POLLING_INTERVAL` | Time between DNS propagation check |
| `AWS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `AWS_SHARED_CREDENTIALS_FILE` | Managed by the AWS client. Shared credentials file. |
| `AWS_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).

## Description

AWS Credentials are automatically detected in the following locations and prioritized in the following order:

1. Environment variables: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, [`AWS_SESSION_TOKEN`]
2. Shared credentials file (defaults to `~/.aws/credentials`, profiles can be specified using `AWS_PROFILE`)
3. Amazon EC2 IAM role

The AWS Region is automatically detected in the following locations and prioritized in the following order:

1. Environment variables: `AWS_REGION`
2. Shared configuration file if `AWS_SDK_LOAD_CONFIG` is set (defaults to `~/.aws/config`, profiles can be specified using `AWS_PROFILE`)

If `AWS_HOSTED_ZONE_ID` is not set, Lego tries to determine the correct public hosted zone via the FQDN.

See also:

- [sessions](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/sessions.html)
- [Setting AWS Credentials](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials)
- [Setting AWS Region](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-the-region)

## Policy

The following AWS IAM policy document describes the permissions required for lego to complete the DNS challenge.

```json
{
   "Version": "2012-10-17",
   "Statement": [
       {
           "Sid": "",
           "Effect": "Allow",
           "Action": [
               "route53:GetChange",
               "route53:ChangeResourceRecordSets",
               "route53:ListResourceRecordSets"
           ],
           "Resource": [
               "arn:aws:route53:::hostedzone/*",
               "arn:aws:route53:::change/*"
           ]
       },
       {
           "Sid": "",
           "Effect": "Allow",
           "Action": "route53:ListHostedZonesByName",
           "Resource": "*"
       }
   ]
}
```




## More information

- [API documentation](https://docs.aws.amazon.com/Route53/latest/APIReference/API_Operations_Amazon_Route_53.html)
- [Go client](https://github.com/aws/aws-sdk-go/aws)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/route53/route53.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
