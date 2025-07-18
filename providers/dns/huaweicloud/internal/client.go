/*
Copyright (c) Huawei Technologies Co., Ltd. 2020-present. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package internal is a partial copy of https://github.com/huaweicloud/huaweicloud-sdk-go-v3/blob/v0.1.159/services/dns/v2/dns_client.go
package internal

import (
	httpclient "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	hwdns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
)

type DnsClient struct {
	HcClient *httpclient.HcHttpClient
}

func NewDnsClient(hcClient *httpclient.HcHttpClient) *DnsClient {
	return &DnsClient{HcClient: hcClient}
}

func (c *DnsClient) ShowRecordSet(request *model.ShowRecordSetRequest) (*model.ShowRecordSetResponse, error) {
	requestDef := hwdns.GenReqDefForShowRecordSet()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowRecordSetResponse), nil
	}
}

func (c *DnsClient) CreateRecordSet(request *model.CreateRecordSetRequest) (*model.CreateRecordSetResponse, error) {
	requestDef := hwdns.GenReqDefForCreateRecordSet()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateRecordSetResponse), nil
	}
}

func (c *DnsClient) UpdateRecordSet(request *model.UpdateRecordSetRequest) (*model.UpdateRecordSetResponse, error) {
	requestDef := hwdns.GenReqDefForUpdateRecordSet()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateRecordSetResponse), nil
	}
}

func (c *DnsClient) DeleteRecordSet(request *model.DeleteRecordSetRequest) (*model.DeleteRecordSetResponse, error) {
	requestDef := hwdns.GenReqDefForDeleteRecordSet()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteRecordSetResponse), nil
	}
}

func (c *DnsClient) ListRecordSetsByZone(request *model.ListRecordSetsByZoneRequest) (*model.ListRecordSetsByZoneResponse, error) {
	requestDef := hwdns.GenReqDefForListRecordSetsByZone()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListRecordSetsByZoneResponse), nil
	}
}

func (c *DnsClient) ListPublicZones(request *model.ListPublicZonesRequest) (*model.ListPublicZonesResponse, error) {
	requestDef := hwdns.GenReqDefForListPublicZones()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListPublicZonesResponse), nil
	}
}
