package internal

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
)

// Client VK client.
type Client struct {
	openstack     *gophercloud.ProviderClient
	authOpts      gophercloud.AuthOptions
	authenticated bool
	baseURL       *url.URL
}

// NewClient creates a Client.
func NewClient(endpoint string, authOpts gophercloud.AuthOptions) (*Client, error) {
	err := validateAuthOptions(authOpts)
	if err != nil {
		return nil, err
	}

	openstackClient, err := openstack.NewClient(authOpts.IdentityEndpoint)
	if err != nil {
		return nil, fmt.Errorf("new client: %w", err)
	}

	baseURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}

	return &Client{
		openstack: openstackClient,
		authOpts:  authOpts,
		baseURL:   baseURL,
	}, nil
}

func (c *Client) ListZones() ([]DNSZone, error) {
	endpoint := c.baseURL.JoinPath("/")

	var zones []DNSZone
	opts := &gophercloud.RequestOpts{JSONResponse: &zones}

	err := c.request(http.MethodGet, endpoint, opts)
	if err != nil {
		return nil, err
	}

	return zones, nil
}

func (c *Client) ListTXTRecords(zoneUUID string) ([]DNSTXTRecord, error) {
	endpoint := c.baseURL.JoinPath(zoneUUID, "txt", "/")

	var records []DNSTXTRecord
	opts := &gophercloud.RequestOpts{JSONResponse: &records}

	err := c.request(http.MethodGet, endpoint, opts)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (c *Client) CreateTXTRecord(zoneUUID string, record *DNSTXTRecord) error {
	endpoint := c.baseURL.JoinPath(zoneUUID, "txt", "/")

	opts := &gophercloud.RequestOpts{
		JSONBody:     record,
		JSONResponse: record,
	}

	return c.request(http.MethodPost, endpoint, opts)
}

func (c *Client) DeleteTXTRecord(zoneUUID, recordUUID string) error {
	endpoint := c.baseURL.JoinPath(zoneUUID, "txt", recordUUID)

	return c.request(http.MethodDelete, endpoint, &gophercloud.RequestOpts{})
}

func (c *Client) request(method string, endpoint *url.URL, options *gophercloud.RequestOpts) error {
	if err := c.lazyAuth(); err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	_, err := c.openstack.Request(method, endpoint.String(), options)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}

	return nil
}

func (c *Client) lazyAuth() error {
	if c.authenticated {
		return nil
	}

	err := openstack.Authenticate(c.openstack, c.authOpts)
	if err != nil {
		return err
	}

	c.authenticated = true

	return nil
}

func validateAuthOptions(opts gophercloud.AuthOptions) error {
	if opts.TenantID == "" {
		return errors.New("project id is missing in credentials information")
	}

	if opts.Username == "" {
		return errors.New("username is missing in credentials information")
	}

	if opts.Password == "" {
		return errors.New("password is missing in credentials information")
	}

	if opts.IdentityEndpoint == "" {
		return errors.New("identity endpoint is missing in config")
	}

	if opts.DomainName == "" {
		return errors.New("domain name is missing in config")
	}

	return nil
}
