/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package signers

import (
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/jmespath/go-jmespath"
	"net/http"
	"strings"
	"time"
)

type SignerEcsInstance struct {
	*credentialUpdater
	sessionCredential *sessionCredential
	credential        *credentials.StsRoleNameOnEcsCredential
	commonApi         func(request *requests.CommonRequest, signer interface{}) (response *responses.CommonResponse, err error)
}

func NewSignereEcsInstance(credential *credentials.StsRoleNameOnEcsCredential, commonApi func(*requests.CommonRequest, interface{}) (response *responses.CommonResponse, err error)) (signer *SignerEcsInstance, err error) {
	signer = &SignerEcsInstance{
		credential: credential,
		commonApi:  commonApi,
	}

	signer.credentialUpdater = &credentialUpdater{
		credentialExpiration: defaultDurationSeconds / 60,
		buildRequestMethod:   signer.buildCommonRequest,
		responseCallBack:     signer.refreshCredential,
		refreshApi:           signer.refreshApi,
	}

	return
}

func (*SignerEcsInstance) GetName() string {
	return "HMAC-SHA1"
}

func (*SignerEcsInstance) GetType() string {
	return ""
}

func (*SignerEcsInstance) GetVersion() string {
	return "1.0"
}

func (signer *SignerEcsInstance) GetAccessKeyId() string {
	if signer.sessionCredential == nil || signer.needUpdateCredential() {
		signer.updateCredential()
	}
	if len(signer.sessionCredential.accessKeyId) <= 0 {
		return ""
	}
	return signer.sessionCredential.accessKeyId
}

func (signer *SignerEcsInstance) GetExtraParam() map[string]string {
	if signer.sessionCredential == nil {
		return make(map[string]string)
	}
	if len(signer.sessionCredential.securityToken) <= 0 {
		return make(map[string]string)
	}
	return map[string]string{"SecurityToken": signer.sessionCredential.securityToken}
}

func (signer *SignerEcsInstance) Sign(stringToSign, secretSuffix string) string {
	secret := signer.sessionCredential.accessKeySecret + secretSuffix
	return ShaHmac1(stringToSign, secret)
}

func (signer *SignerEcsInstance) buildCommonRequest() (request *requests.CommonRequest, err error) {
	request = requests.NewCommonRequest()
	return
}

func (signer *SignerEcsInstance) refreshApi(request *requests.CommonRequest) (response *responses.CommonResponse, err error) {
	requestUrl := "http://100.100.100.200/latest/meta-data/ram/security-credentials/" + signer.credential.RoleName
	httpRequest, err := http.NewRequest(requests.GET, requestUrl, strings.NewReader(""))
	if err != nil {
		fmt.Println("refresh Ecs sts token err", err)
		return
	}
	httpClient := &http.Client{}
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		fmt.Println("refresh Ecs sts token err", err)
		return
	}

	response = responses.NewCommonResponse()
	err = responses.Unmarshal(response, httpResponse, "")

	return
}

func (signer *SignerEcsInstance) refreshCredential(response *responses.CommonResponse) (err error) {
	if response.GetHttpStatus() != http.StatusOK {
		fmt.Println("refresh Ecs sts token err, httpStatus: " + string(response.GetHttpStatus()) + ", message = " + response.GetHttpContentString())
		return
	}
	var data interface{}
	err = json.Unmarshal(response.GetHttpContentBytes(), &data)
	if err != nil {
		fmt.Println("refresh Ecs sts token err, json.Unmarshal fail", err)
		return
	}
	code, err := jmespath.Search("Code", data)
	if err != nil {
		fmt.Println("refresh Ecs sts token err, fail to get Code", err)
		return
	}
	if code.(string) != "Success" {
		fmt.Println("refresh Ecs sts token err, Code is not Success", err)
		return
	}
	accessKeyId, err := jmespath.Search("AccessKeyId", data)
	if err != nil {
		fmt.Println("refresh Ecs sts token err, fail to get AccessKeyId", err)
		return
	}
	accessKeySecret, err := jmespath.Search("AccessKeySecret", data)
	if err != nil {
		fmt.Println("refresh Ecs sts token err, fail to get AccessKeySecret", err)
		return
	}
	securityToken, err := jmespath.Search("SecurityToken", data)
	if err != nil {
		fmt.Println("refresh Ecs sts token err, fail to get SecurityToken", err)
		return
	}
	expiration, err := jmespath.Search("Expiration", data)
	if err != nil {
		fmt.Println("refresh Ecs sts token err, fail to get Expiration", err)
		return
	}
	if accessKeyId == nil || accessKeySecret == nil || securityToken == nil {
		if signer.sessionCredential == nil {
			panic("refresh Ecs sts token failed, accessKeyId, accessKeySecret or securityToken is null")
		}
	}

	expirationTime, err := time.Parse("2006-01-02T15:04:05Z", expiration.(string))
	signer.credentialExpiration = int(expirationTime.Unix() - time.Now().Unix())
	signer.sessionCredential = &sessionCredential{
		accessKeyId:     accessKeyId.(string),
		accessKeySecret: accessKeySecret.(string),
		securityToken:   securityToken.(string),
	}

	return
}

func (signer *SignerEcsInstance) Shutdown() {

}
