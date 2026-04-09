package internal

import (
	"io"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type Client struct {
	*ucloud.Client
}

func NewClient(config *ucloud.Config, credential *auth.Credential) *Client {
	client := ucloud.NewClientWithMeta(config, credential, ucloud.ClientMeta{Product: "UDNR"})
	client.GetLogger().SetOutput(io.Discard)

	return &Client{
		Client: client,
	}
}

func (c *Client) NewDomainDNSAddRequest() *DomainDNSAddRequest {
	req := &DomainDNSAddRequest{}

	// setup request with client config
	c.SetupRequest(req)

	// setup retryable with default retry policy (retry for non-create action and common error)
	req.SetRetryable(false)

	return req
}

func (c *Client) NewDeleteDNSRecordRequest() *DeleteDNSRecordRequest {
	req := &DeleteDNSRecordRequest{}

	// setup request with client config
	c.SetupRequest(req)

	// setup retryable with default retry policy (retry for non-create action and common error)
	req.SetRetryable(false)

	return req
}

func (c *Client) NewDomainDNSQueryRequest() *DomainDNSQueryRequest {
	req := &DomainDNSQueryRequest{}

	// setup request with client config
	c.SetupRequest(req)

	// setup retryable with default retry policy (retry for non-create action and common error)
	req.SetRetryable(false)

	return req
}

// DomainDNSAdd adds a DNS record to a domain.
// https://docs.ucloud.cn/api/udnr-api/udnr_domain_dns_add
func (c *Client) DomainDNSAdd(req *DomainDNSAddRequest) (*DomainDNSAddResponse, error) {
	var res DomainDNSAddResponse

	reqCopier := *req

	err := c.InvokeAction("UdnrDomainDNSAdd", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}

// DeleteDNSRecord delete a DNS record.
// https://docs.ucloud.cn/api/udnr-api/udnr_delete_dns_record
func (c *Client) DeleteDNSRecord(req *DeleteDNSRecordRequest) (*DeleteDNSRecordResponse, error) {
	var res DeleteDNSRecordResponse

	reqCopier := *req

	err := c.InvokeAction("UdnrDeleteDnsRecord", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}

// DomainDNSQuery gets DNS records of a domain.
// https://docs.ucloud.cn/api/udnr-api/udnr_domain_dns_query
func (c *Client) DomainDNSQuery(req *DomainDNSQueryRequest) (*DomainDNSQueryResponse, error) {
	var res DomainDNSQueryResponse

	reqCopier := *req

	err := c.InvokeAction("UdnrDomainDNSQuery", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
