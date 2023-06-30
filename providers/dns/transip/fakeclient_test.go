package transip

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/transip/gotransip/v6/domain"
	"github.com/transip/gotransip/v6/rest"
)

type dnsEntryWrapper struct {
	DNSEntry domain.DNSEntry `json:"dnsEntry"`
}

type dnsEntriesWrapper struct {
	DNSEntries []domain.DNSEntry `json:"dnsEntries"`
}

type fakeClient struct {
	dnsEntries           []domain.DNSEntry
	setDNSEntriesLatency time.Duration
	getInfoLatency       time.Duration
	domainName           string
}

func (f *fakeClient) PutWithResponse(_ rest.Request) (rest.Response, error) {
	panic("not implemented")
}

func (f *fakeClient) PostWithResponse(_ rest.Request) (rest.Response, error) {
	panic("not implemented")
}

func (f *fakeClient) PatchWithResponse(_ rest.Request) (rest.Response, error) {
	panic("not implemented")
}

func (f *fakeClient) Get(request rest.Request, dest interface{}) error {
	if f.getInfoLatency != 0 {
		time.Sleep(f.getInfoLatency)
	}

	if request.Endpoint != fmt.Sprintf("/domains/%s/dns", f.domainName) {
		return fmt.Errorf("function GET for endpoint %s not implemented", request.Endpoint)
	}

	entries := dnsEntriesWrapper{DNSEntries: f.dnsEntries}

	body, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("can't encode json: %w", err)
	}

	err = json.Unmarshal(body, dest)
	if err != nil {
		return fmt.Errorf("can't decode json: %w", err)
	}

	return nil
}

func (f *fakeClient) Put(request rest.Request) error {
	if f.getInfoLatency != 0 {
		time.Sleep(f.getInfoLatency)
	}

	return fmt.Errorf("function PUT for endpoint %s not implemented", request.Endpoint)
}

func (f *fakeClient) Post(request rest.Request) error {
	if f.getInfoLatency != 0 {
		time.Sleep(f.getInfoLatency)
	}

	if request.Endpoint != fmt.Sprintf("/domains/%s/dns", f.domainName) {
		return fmt.Errorf("function POST for endpoint %s not implemented", request.Endpoint)
	}

	body, err := request.GetJSONBody()
	if err != nil {
		return fmt.Errorf("unable get request body")
	}

	var entry dnsEntryWrapper
	if err := json.Unmarshal(body, &entry); err != nil {
		return fmt.Errorf("unable to decode request body")
	}

	f.dnsEntries = append(f.dnsEntries, entry.DNSEntry)

	return nil
}

func (f *fakeClient) Delete(request rest.Request) error {
	if f.getInfoLatency != 0 {
		time.Sleep(f.getInfoLatency)
	}

	if request.Endpoint != fmt.Sprintf("/domains/%s/dns", f.domainName) {
		return fmt.Errorf("function DELETE for endpoint %s not implemented", request.Endpoint)
	}

	body, err := request.GetJSONBody()
	if err != nil {
		return fmt.Errorf("unable get request body")
	}

	var entry dnsEntryWrapper
	if err := json.Unmarshal(body, &entry); err != nil {
		return fmt.Errorf("unable to decode request body")
	}

	cp := make([]domain.DNSEntry, 0)

	for _, e := range f.dnsEntries {
		if e.Name == entry.DNSEntry.Name {
			continue
		}

		cp = append(cp, e)
	}

	f.dnsEntries = cp

	return nil
}

func (f *fakeClient) Patch(request rest.Request) error {
	if f.getInfoLatency != 0 {
		time.Sleep(f.getInfoLatency)
	}

	return fmt.Errorf("function PATCH for endpoint %s not implemented", request.Endpoint)
}
