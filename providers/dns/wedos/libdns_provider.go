package wedos

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/wedos/libdns"
)

// Provider implements the libdns interfaces for Wedos.
// things not implemented since, not required yet || for lego :
// * altering DNS records without know ID
// * returning IDs of newly created records.
type Provider struct {
	Username     string
	WapiPassword string
	HTTPClient   *http.Client
	lock         sync.Mutex
}

func (p *Provider) Ping(ctx context.Context) error {
	_, e := askWedos(ctx, p.HTTPClient, p.Username, p.WapiPassword, "ping", nil)
	return e
}

// Commit not really required, all changes will be auto-committed after 5 mintues.
func (p *Provider) Commit(ctx context.Context, zone string) error {
	_, e := askWedos(ctx, p.HTTPClient, p.Username, p.WapiPassword, "dns-domain-commit", map[string]interface{}{
		"domain": strings.TrimRight(zone, "."),
	})
	return e
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	resp, err := askWedos(ctx, p.HTTPClient, p.Username, p.WapiPassword, "dns-rows-list",
		map[string]interface{}{
			"domain": strings.TrimRight(zone, "."),
		})
	if err != nil {
		return nil, err
	}

	recs := make([]libdns.Record, 0, len(resp.DNSRowsList))
	for _, row := range resp.DNSRowsList {
		record := libdns.Record{
			ID:    row.ID,
			Type:  row.RDType,
			Name:  row.Name,
			Value: row.RData,
			TTL:   0,
		}
		ttl, te := row.TTL.Int64()
		if te != nil {
			return recs, te
		}
		record.TTL = time.Duration(ttl) * time.Second
		recs = append(recs, record)
	}

	return recs, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if records == nil {
		return nil, nil
	}

	var created []libdns.Record
	for _, record := range records {
		if record.TTL.Seconds() == 0 {
			record.TTL = 1800 * time.Second
		}

		_, err := askWedos(ctx, p.HTTPClient, p.Username, p.WapiPassword, "dns-row-add",
			map[string]interface{}{
				"domain": strings.TrimRight(zone, "."),
				"name":   record.Name,
				"ttl":    int(record.TTL.Seconds()),
				"type":   record.Type,
				"rdata":  record.Value,
			})
		if err != nil {
			return created, err
		}
		created = append(created, record)
	}
	return created, nil
}

// DeleteRecords deletes the records from the zone. If a record does not have an ID,
// it will be looked up. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if records == nil {
		return nil, nil
	}

	var removed []libdns.Record
	for _, record := range records {
		if record.ID == "" {
			return removed, errors.New("removing record without ID is not implemented")
		}

		_, err := askWedos(ctx, p.HTTPClient, p.Username, p.WapiPassword, "dns-row-delete",
			map[string]interface{}{
				"domain": strings.TrimRight(zone, "."),
				"row_id": record.ID,
			})
		if err != nil {
			return removed, err
		}
		removed = append(removed, record)
	}
	return removed, nil
}

// SetRecords sets the records in the zone, either by updating existing records
// or creating new ones. It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if records == nil {
		return nil, nil
	}

	var done []libdns.Record
	for _, record := range records {
		if record.TTL.Seconds() == 0 {
			record.TTL = 1800 * time.Second
		}

		if record.ID == "" {
			_, err := askWedos(ctx, p.HTTPClient, p.Username, p.WapiPassword, "dns-row-add",
				map[string]interface{}{
					"domain": strings.TrimRight(zone, "."),
					"name":   record.Name,
					"ttl":    int(record.TTL.Seconds()),
					"type":   record.Type,
					"rdata":  record.Value,
				})
			if err != nil {
				return done, err
			}
		} else {
			_, err := askWedos(ctx, p.HTTPClient, p.Username, p.WapiPassword, "dns-row-update",
				map[string]interface{}{
					"domain": strings.TrimRight(zone, "."),
					"row_id": record.ID,
					// "name":   record.Name, not editable
					"ttl": int(record.TTL.Seconds()),
					// "type":   record.Type, not editable
					"rdata": record.Value,
				})
			if err != nil {
				return done, err
			}
		}
		done = append(done, record)
	}
	return done, nil
}

func (p *Provider) FillRecordID(ctx context.Context, zone string, record libdns.Record) (libdns.Record, error) {
	existing, err := p.GetRecords(ctx, zone)
	if err != nil {
		return record, err
	}
	if existing == nil {
		return record, nil
	}
	for _, candidate := range existing {
		if candidate.Type == record.Type && candidate.Name == record.Name {
			record.ID = candidate.ID
			return record, nil
		}
	}
	return record, nil
}

// Interface guards.
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
