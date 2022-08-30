package vkcloud

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
)

type Client struct {
	endpoint      string
	openstack     *gophercloud.ProviderClient
	authOpts      gophercloud.AuthOptions
	authenticated bool
}

func NewClient(endpoint string, authOpts gophercloud.AuthOptions) (*Client, error) {
	openstackClient, err := openstack.NewClient(authOpts.IdentityEndpoint)
	if err != nil {
		return nil, err
	}

	return &Client{
		endpoint:      endpoint,
		openstack:     openstackClient,
		authOpts:      authOpts,
		authenticated: false,
	}, nil
}

func (r *Client) LazyAuth() error {
	if r.authenticated {
		return nil
	}

	err := openstack.Authenticate(r.openstack, r.authOpts)
	if err != nil {
		return err
	}

	r.authenticated = true

	return nil
}

func (r *Client) ListZones() (DNSZones, error) {
	if err := r.LazyAuth(); err != nil {
		return nil, err
	}

	var zones DNSZones

	_, err := r.openstack.Request(
		"GET",
		fmt.Sprintf("%s/v2/dns/", r.endpoint),
		&gophercloud.RequestOpts{
			JSONResponse: &zones,
		},
	)
	if err != nil {
		return nil, err
	}

	return zones, nil
}

func (r *Client) ListTXTRecords(zoneUUID string) (DNSTXTRecords, error) {
	if err := r.LazyAuth(); err != nil {
		return nil, err
	}

	var records DNSTXTRecords

	_, err := r.openstack.Request(
		"GET",
		fmt.Sprintf("%s/v2/dns/%s/txt/", r.endpoint, zoneUUID),
		&gophercloud.RequestOpts{
			JSONResponse: &records,
		},
	)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (r *Client) CreateTXTRecord(zoneUUID string, record *DNSTXTRecord) error {
	if err := r.LazyAuth(); err != nil {
		return err
	}

	_, err := r.openstack.Request(
		"POST",
		fmt.Sprintf("%s/v2/dns/%s/txt/", r.endpoint, zoneUUID),
		&gophercloud.RequestOpts{
			JSONBody:     record,
			JSONResponse: record,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *Client) DeleteTXTRecord(zoneUUID, recordUUID string) error {
	if err := r.LazyAuth(); err != nil {
		return err
	}

	_, err := r.openstack.Request(
		"DELETE",
		fmt.Sprintf("%s/v2/dns/%s/txt/%s", r.endpoint, zoneUUID, recordUUID),
		&gophercloud.RequestOpts{},
	)
	if err != nil {
		return err
	}

	return nil
}
