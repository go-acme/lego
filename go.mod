module github.com/go-acme/lego/v5

go 1.25.0

ignore (
	./.github
	./docs
)

require (
	cloud.google.com/go/compute/metadata v0.9.0
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.22.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.14.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns v1.2.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns v1.3.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph v0.10.0
	github.com/BurntSushi/toml v1.6.0
	github.com/akamai/AkamaiOPEN-edgegrid-golang/v13 v13.3.0
	github.com/alibabacloud-go/darabonba-openapi/v2 v2.2.3
	github.com/alibabacloud-go/tea v1.5.2
	github.com/aliyun/credentials-go v1.4.7
	github.com/aws/aws-sdk-go-v2 v1.42.1
	github.com/aws/aws-sdk-go-v2/config v1.32.30
	github.com/aws/aws-sdk-go-v2/credentials v1.19.29
	github.com/aws/aws-sdk-go-v2/service/lightsail v1.57.1
	github.com/aws/aws-sdk-go-v2/service/route53 v1.64.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.105.1
	github.com/aws/aws-sdk-go-v2/service/sts v1.44.1
	github.com/aziontech/azionapi-go-sdk v0.147.0
	github.com/baidubce/bce-sdk-go v0.9.270
	github.com/bodgit/tsig v1.3.1
	github.com/bradfitz/gomemcache v0.0.0-20260422231931-4d751bb6e37c
	github.com/cenkalti/backoff/v5 v5.0.3
	github.com/dnsimple/dnsimple-go/v9 v9.1.0
	github.com/exoscale/egoscale/v3 v3.1.41
	github.com/go-acme/alidns-20150109/v5 v5.5.0
	github.com/go-acme/esa-20240910/v3 v3.4.0
	github.com/go-acme/jdcloud-sdk-go v1.64.0
	github.com/go-acme/tencentclouddnspod v1.3.24
	github.com/go-acme/tencentedgdeone v1.3.38
	github.com/go-jose/go-jose/v4 v4.1.4
	github.com/go-viper/mapstructure/v2 v2.5.0
	github.com/google/go-cmp v0.7.0
	github.com/google/go-querystring v1.2.0
	github.com/google/uuid v1.6.0
	github.com/gophercloud/gophercloud v1.14.1
	github.com/gophercloud/utils v0.0.0-20231010081019-80377eca5d56
	github.com/hashicorp/go-retryablehttp v0.7.8
	github.com/hashicorp/go-version v1.9.0
	github.com/huaweicloud/huaweicloud-sdk-go-v3 v0.1.205
	github.com/infobloxopen/infoblox-go-client/v2 v2.11.0
	github.com/joho/godotenv v1.5.1
	github.com/labbsr0x/bindman-dns-webhook v1.0.2
	github.com/ldez/grignotin v0.10.1
	github.com/linode/linodego v1.69.1
	github.com/liquidweb/liquidweb-go v1.6.4
	github.com/mattn/go-isatty v0.0.22
	github.com/mattn/go-zglob v0.0.6
	github.com/miekg/dns v1.1.72
	github.com/mimuret/golang-iij-dpf v0.9.1
	github.com/namedotcom/go/v4 v4.0.2
	github.com/nrdcg/auroradns v1.2.0
	github.com/nrdcg/bunny-go v0.1.0
	github.com/nrdcg/desec v0.11.1
	github.com/nrdcg/freemyip v0.3.0
	github.com/nrdcg/goacmedns v0.2.0
	github.com/nrdcg/goinwx v0.12.0
	github.com/nrdcg/mailinabox v0.3.0
	github.com/nrdcg/namesilo v0.5.0
	github.com/nrdcg/nodion v0.1.0
	github.com/nrdcg/oci-go-sdk/common/v1065 v1065.120.0
	github.com/nrdcg/oci-go-sdk/dns/v1065 v1065.120.0
	github.com/nrdcg/porkbun v0.4.0
	github.com/nrdcg/vegadns v0.3.0
	github.com/nzdjb/go-metaname v1.0.0
	github.com/ovh/go-ovh v1.9.0
	github.com/pquerna/otp v1.5.0
	github.com/regfish/regfish-dnsapi-go v0.1.1
	github.com/sacloud/api-client-go v0.3.5
	github.com/sacloud/iaas-api-go v1.29.2
	github.com/scaleway/scaleway-sdk-go v1.0.0-beta.36
	github.com/selectel/domains-go v1.1.0
	github.com/selectel/go-selvpcclient/v4 v4.2.0
	github.com/softlayer/softlayer-go v1.2.1
	github.com/stretchr/testify v1.11.1
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.3.133
	github.com/transip/gotransip/v6 v6.27.2
	github.com/ucloud/ucloud-sdk-go v0.22.90
	github.com/ultradns/ultradns-go-sdk v1.8.2-20260507133303-3f324c7
	github.com/urfave/cli/v3 v3.10.1
	github.com/vinyldns/go-vinyldns v0.9.18
	github.com/volcengine/volc-sdk-golang v1.0.251
	github.com/vultr/govultr/v3 v3.31.2
	github.com/yandex-cloud/go-genproto v0.95.0
	github.com/yandex-cloud/go-sdk/services/dns v0.0.65
	github.com/yandex-cloud/go-sdk/v2 v2.136.0
	gitlab.com/greyxor/slogor v1.6.10
	golang.org/x/crypto v0.54.0
	golang.org/x/net v0.57.0
	golang.org/x/oauth2 v0.36.0
	golang.org/x/text v0.40.0
	golang.org/x/time v0.15.0
	google.golang.org/api v0.288.0
	gopkg.in/ns1/ns1-go.v2 v2.18.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	software.sslmate.com/src/go-pkcs12 v0.7.3
)

require (
	cloud.google.com/go/auth v0.20.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	github.com/AdamSLevy/jsonrpc2/v14 v14.1.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.12.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.7.2 // indirect
	github.com/alexbrainman/sspi v0.0.0-20250919150558-7d374ff0d59e // indirect
	github.com/alibabacloud-go/alibabacloud-gateway-spi v0.0.5 // indirect
	github.com/alibabacloud-go/debug v1.0.1 // indirect
	github.com/alibabacloud-go/openapi-util v0.1.2 // indirect
	github.com/alibabacloud-go/tea-utils/v2 v2.0.9 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.14 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.31 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.30 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.31 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.4.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.32.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.37.1 // indirect
	github.com/aws/smithy-go v1.27.3 // indirect
	github.com/benbjohnson/clock v1.3.5 // indirect
	github.com/bodgit/gssapi v0.0.4 // indirect
	github.com/boombuler/barcode v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/clbanning/mxj/v2 v2.7.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.13 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-errors/errors v1.0.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.30.3 // indirect
	github.com/go-resty/resty/v2 v2.17.2 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/gofrs/flock v0.13.0 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.17 // indirect
	github.com/googleapis/gax-go/v2 v2.22.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/goidentity/v6 v6.0.1 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/json-iterator/go v1.1.13-0.20220915233716-71ac16282d12 // indirect
	github.com/k0kubun/go-ansi v0.0.0-20180517002512-3bf9e2903213 // indirect
	github.com/kolo/xmlrpc v0.0.0-20220921171641-a4b6fa1dd06b // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/labbsr0x/goh v1.0.1 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/liquidweb/liquidweb-cli v0.7.0 // indirect
	github.com/magiconair/properties v1.8.10 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/openshift/gssapi v0.0.0-20161010215902-5fb4217df13b // indirect
	github.com/pelletier/go-toml/v2 v2.1.0 // indirect
	github.com/peterhellberg/link v1.2.0 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/sacloud/go-http v0.1.9 // indirect
	github.com/sacloud/packages-go v0.1.0 // indirect
	github.com/sacloud/saclient-go v0.4.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/sirupsen/logrus v1.9.4 // indirect
	github.com/softlayer/xmlrpc v0.0.0-20200409220501-5f089df7cb7e // indirect
	github.com/sony/gobreaker/v2 v2.4.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.18.2 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tjfoc/gmsm v1.4.1 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.mongodb.org/mongo-driver v1.17.9 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.67.0 // indirect
	go.opentelemetry.io/otel v1.43.0 // indirect
	go.opentelemetry.io/otel/metric v1.43.0 // indirect
	go.opentelemetry.io/otel/trace v1.43.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/ratelimit v0.3.1 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/exp v0.0.0-20260410095643-746e56fc9e2f // indirect
	golang.org/x/mod v0.37.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/tools v0.47.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260414002931-afd174a4e478 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260630182238-925bb5da69e7 // indirect
	google.golang.org/grpc v1.82.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/ini.v1 v1.67.3 // indirect
)
