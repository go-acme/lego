package egoscale

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"path"
	"strconv"
)

// RunstatusService is a runstatus service
type RunstatusService struct {
	ID      int    `json:"id"` // missing field
	Name    string `json:"name"`
	PageURL string `json:"page_url,omitempty"` // fake field
	State   string `json:"state,omitempty"`
	URL     string `json:"url,omitempty"`
}

// FakeID fills up the ID field as it's currently missing
func (service *RunstatusService) FakeID() error {
	if service.ID > 0 {
		return nil
	}

	if service.URL == "" {
		return fmt.Errorf("empty URL for %#v", service)
	}

	u, err := url.Parse(service.URL)
	if err != nil {
		return err
	}

	s := path.Base(u.Path)
	id, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	service.ID = id
	return nil
}

// Match returns true if the other service has got similarities with itself
func (service RunstatusService) Match(other RunstatusService) bool {
	if other.Name != "" && service.Name == other.Name {
		return true
	}

	if other.ID > 0 && service.ID == other.ID {
		return true
	}

	return false
}

// RunstatusServiceList service list
type RunstatusServiceList struct {
	Services []RunstatusService `json:"services"`
}

// DeleteRunstatusService delete runstatus service
func (client *Client) DeleteRunstatusService(ctx context.Context, service RunstatusService) error {
	if service.URL == "" {
		return fmt.Errorf("empty URL for %#v", service)
	}

	_, err := client.runstatusRequest(ctx, service.URL, nil, "DELETE")
	return err
}

// CreateRunstatusService create runstatus service
func (client *Client) CreateRunstatusService(ctx context.Context, service RunstatusService) (*RunstatusService, error) {
	if service.PageURL == "" {
		return nil, fmt.Errorf("empty Page URL for %#v", service)
	}

	page, err := client.GetRunstatusPage(ctx, RunstatusPage{URL: service.PageURL})
	if err != nil {
		return nil, err
	}

	resp, err := client.runstatusRequest(ctx, page.ServicesURL, service, "POST")
	if err != nil {
		return nil, err
	}

	s := &RunstatusService{}
	if err := json.Unmarshal(resp, s); err != nil {
		return nil, err
	}

	return s, nil
}

// GetRunstatusService displays service detail.
func (client *Client) GetRunstatusService(ctx context.Context, service RunstatusService) (*RunstatusService, error) {
	if service.URL != "" {
		return client.getRunstatusService(ctx, service.URL)
	}

	if service.PageURL == "" {
		return nil, fmt.Errorf("empty Page URL in %#v", service)
	}

	page, err := client.getRunstatusPage(ctx, service.PageURL)
	if err != nil {
		return nil, err
	}

	ss, err := client.ListRunstatusServices(ctx, *page)
	if err != nil {
		return nil, err
	}

	for i := range ss {
		if ss[i].Match(service) {
			return &ss[i], nil
		}
	}

	return nil, fmt.Errorf("%#v not found", service)
}

func (client *Client) getRunstatusService(ctx context.Context, serviceURL string) (*RunstatusService, error) {
	resp, err := client.runstatusRequest(ctx, serviceURL, nil, "GET")
	if err != nil {
		return nil, err
	}

	s := &RunstatusService{}
	if err := json.Unmarshal(resp, &s); err != nil {
		return nil, err
	}

	if err := s.FakeID(); err != nil {
		log.Printf("bad fake ID for %#v, %s", s, err)
	}

	return s, nil
}

// ListRunstatusServices displays the list of services.
func (client *Client) ListRunstatusServices(ctx context.Context, page RunstatusPage) ([]RunstatusService, error) {
	if page.ServicesURL == "" {
		return nil, fmt.Errorf("empty Services URL for %#v", page)
	}

	resp, err := client.runstatusRequest(ctx, page.ServicesURL, nil, "GET")
	if err != nil {
		return nil, err
	}

	var p *RunstatusServiceList
	if err := json.Unmarshal(resp, &p); err != nil {
		return nil, err
	}

	// NOTE: fix the missing IDs
	for i := range p.Services {
		if err := p.Services[i].FakeID(); err != nil {
			log.Printf("bad fake ID for %#v, %s", p.Services[i], err)
		}
	}

	// NOTE: no pagination
	return p.Services, nil
}
