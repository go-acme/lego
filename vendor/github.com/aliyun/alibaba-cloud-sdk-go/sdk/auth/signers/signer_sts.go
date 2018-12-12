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

type SignerSts struct {
	credential *credentials.StsTokenCredential
}

func NewSignerSts(credential *credentials.StsTokenCredential) (*SignerSts, error) {
	return &SignerSts{
		credential: credential,
	}, nil
}

func (*SignerSts) GetName() string {
	return "HMAC-SHA1"
}

func (*SignerSts) GetType() string {
	return ""
}

func (*SignerSts) GetVersion() string {
	return "1.0"
}

func (signer *SignerSts) GetAccessKeyId() string {
	return signer.credential.AccessKeyId
}

func (signer *SignerSts) GetExtraParam() map[string]string {
	return map[string]string{"SecurityToken": signer.credential.AccessKeyStsToken}
}

func (signer *SignerSts) Sign(stringToSign, secretSuffix string) string {
	secret := signer.credential.AccessKeySecret + secretSuffix
	return ShaHmac1(stringToSign, secret)
}

func (signer *SignerSts) Shutdown() {

}
