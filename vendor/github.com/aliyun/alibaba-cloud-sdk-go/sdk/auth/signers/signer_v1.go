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
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
)

type SignerV1 struct {
	credential *credentials.BaseCredential
}

func (signer *SignerV1) GetExtraParam() map[string]string {
	return nil
}

func NewSignerV1(credential *credentials.BaseCredential) (*SignerV1, error) {
	return &SignerV1{
		credential: credential,
	}, nil
}

func (*SignerV1) GetName() string {
	return "HMAC-SHA1"
}

func (*SignerV1) GetType() string {
	return ""
}

func (*SignerV1) GetVersion() string {
	return "1.0"
}

func (signer *SignerV1) GetAccessKeyId() string {
	return signer.credential.AccessKeyId
}

func (signer *SignerV1) Sign(stringToSign, secretSuffix string) string {
	secret := signer.credential.AccessKeySecret + secretSuffix
	return ShaHmac1(stringToSign, secret)
}

func (signer *SignerV1) Shutdown() {

}
