package iijdpf

import (
	"context"
	"errors"
	"fmt"

	dpfzones "github.com/mimuret/golang-iij-dpf/pkg/apis/dpf/v1/zones"
	dpfapiutils "github.com/mimuret/golang-iij-dpf/pkg/apiutils"
	dpftypes "github.com/mimuret/golang-iij-dpf/pkg/types"
)

func (d *DNSProvider) addTxtRecord(zoneID, fqdn, rdata string) error {
	ctx := context.Background()

	r, err := dpfapiutils.GetRecordFromZoneID(ctx, d.client, zoneID, fqdn, dpfzones.TypeTXT)
	if err != nil && !errors.Is(err, dpfapiutils.ErrRecordNotFound) {
		return err
	}

	if r != nil {
		r.RData = append(r.RData, dpfzones.RecordRDATA{Value: rdata})

		_, _, err = dpfapiutils.SyncUpdate(ctx, d.client, r, nil)
		if err != nil {
			return fmt.Errorf("failed to update record: %w", err)
		}

		return nil
	}

	record := &dpfzones.Record{
		AttributeMeta: dpfzones.AttributeMeta{ZoneID: zoneID},
		Name:          fqdn,
		TTL:           dpftypes.NullablePositiveInt32(d.config.TTL),
		RRType:        dpfzones.TypeTXT,
		RData:         dpfzones.RecordRDATASlice{dpfzones.RecordRDATA{Value: rdata}},
		Description:   "ACME",
	}

	_, _, err = dpfapiutils.SyncCreate(ctx, d.client, record, nil)
	if err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}

	return nil
}

func (d *DNSProvider) deleteTxtRecord(zoneID, fqdn, rdata string) error {
	r, err := dpfapiutils.GetRecordFromZoneID(context.Background(), d.client, zoneID, fqdn, dpfzones.TypeTXT)
	if err != nil {
		if errors.Is(err, dpfapiutils.ErrRecordNotFound) {
			// empty target rrset
			return nil
		}
		return err
	}

	if len(r.RData) == 1 {
		// delete rrset
		if _, _, err := dpfapiutils.SyncDelete(context.Background(), d.client, r); err != nil {
			return fmt.Errorf("failed to delete record: %w", err)
		}

		return nil
	}

	// delete rdata
	rdataSlice := dpfzones.RecordRDATASlice{}
	for _, v := range r.RData {
		if v.Value != rdata {
			rdataSlice = append(rdataSlice, v)
		}
	}
	r.RData = rdataSlice

	if _, _, err := dpfapiutils.SyncUpdate(context.Background(), d.client, r, nil); err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	return nil
}

func (d *DNSProvider) commit(zoneID string) error {
	apply := &dpfzones.ZoneApply{
		AttributeMeta: dpfzones.AttributeMeta{ZoneID: zoneID},
		Description:   "ACME Processing",
	}

	_, _, err := dpfapiutils.SyncApply(context.Background(), d.client, apply, nil)
	if err != nil {
		return fmt.Errorf("failed to apply zone: %w", err)
	}

	return nil
}
