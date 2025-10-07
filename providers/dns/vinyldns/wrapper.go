package vinyldns

import (
	"context"
	"fmt"

	"github.com/cenkalti/backoff/v5"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/wait"
	"github.com/vinyldns/go-vinyldns/vinyldns"
)

func (d *DNSProvider) getRecordSet(fqdn string) (*vinyldns.RecordSet, error) {
	zoneName, hostName, err := splitDomain(fqdn)
	if err != nil {
		return nil, err
	}

	zone, err := d.client.ZoneByName(zoneName)
	if err != nil {
		return nil, err
	}

	allRecordSets, err := d.client.RecordSetsListAll(zone.ID, vinyldns.ListFilter{NameFilter: hostName})
	if err != nil {
		return nil, err
	}

	var recordSets []vinyldns.RecordSet
	for _, i := range allRecordSets {
		if i.Type == "TXT" {
			recordSets = append(recordSets, i)
		}
	}

	switch {
	case len(recordSets) > 1:
		return nil, fmt.Errorf("ambiguous recordset definition of %s", fqdn)
	case len(recordSets) == 1:
		return &recordSets[0], nil
	default:
		return nil, nil
	}
}

func (d *DNSProvider) createRecordSet(ctx context.Context, fqdn string, records []vinyldns.Record) error {
	zoneName, hostName, err := splitDomain(fqdn)
	if err != nil {
		return err
	}

	zone, err := d.client.ZoneByName(zoneName)
	if err != nil {
		return err
	}

	recordSet := vinyldns.RecordSet{
		Name:    hostName,
		ZoneID:  zone.ID,
		Type:    "TXT",
		TTL:     d.config.TTL,
		Records: records,
	}

	resp, err := d.client.RecordSetCreate(&recordSet)
	if err != nil {
		return err
	}

	return d.waitForChanges(ctx, "CreateRS", resp)
}

func (d *DNSProvider) updateRecordSet(ctx context.Context, recordSet *vinyldns.RecordSet, newRecords []vinyldns.Record) error {
	operation := "delete"
	if len(recordSet.Records) < len(newRecords) {
		operation = "add"
	}

	recordSet.Records = newRecords
	recordSet.TTL = d.config.TTL

	resp, err := d.client.RecordSetUpdate(recordSet)
	if err != nil {
		return err
	}

	return d.waitForChanges(ctx, "UpdateRS - "+operation, resp)
}

func (d *DNSProvider) deleteRecordSet(ctx context.Context, existingRecord *vinyldns.RecordSet) error {
	resp, err := d.client.RecordSetDelete(existingRecord.ZoneID, existingRecord.ID)
	if err != nil {
		return err
	}

	return d.waitForChanges(ctx, "DeleteRS", resp)
}

func (d *DNSProvider) waitForChanges(ctx context.Context, operation string, resp *vinyldns.RecordSetUpdateResponse) error {
	return wait.Retry(ctx,
		func() error {
			change, err := d.client.RecordSetChange(resp.Zone.ID, resp.RecordSet.ID, resp.ChangeID)
			if err != nil {
				return fmt.Errorf("failed to query change status: %w", err)
			}

			if change.Status != "Complete" {
				return fmt.Errorf("waiting operation: %s, zoneID: %s, recordsetID: %s, changeID: %s",
					operation, resp.Zone.ID, resp.RecordSet.ID, resp.ChangeID)
			}

			return nil
		},
		backoff.WithBackOff(backoff.NewConstantBackOff(d.config.PollingInterval)),
		backoff.WithMaxElapsedTime(d.config.PropagationTimeout),
	)
}

// splitDomain splits the hostname from the authoritative zone, and returns both parts.
func splitDomain(fqdn string) (string, string, error) {
	zone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", "", fmt.Errorf("could not find zone: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(fqdn, zone)
	if err != nil {
		return "", "", err
	}

	return zone, subDomain, nil
}
