package online

import (
	"fmt"
	"net/http"
)

type RecordPatch struct {
	Name       string        `json:"name"`
	Type       string        `json:"type"`
	Records    []PatchRecord `json:"records"`
	ChangeType string        `json:"changeType"`
}

type PatchRecord struct {
	TTL  string `json:"ttl"`
	Data string `json:"data"`
}

type Record struct {
	Type string
	Name string
	TTL  string
	Data string
}

func (d *OnlinedomainAPI) AddRecords(domainName string, rec []Record) error {

	var recsToAdd []RecordPatch

	for _, aRec := range rec {
		recsToAdd = append(recsToAdd, RecordPatch{
			Name:       aRec.Name,
			ChangeType: "ADD",
			Type:       aRec.Type,
			Records: []PatchRecord{
				{
					Data: aRec.Data,
					TTL:  aRec.TTL,
				},
			},
		})
	}

	resp, err := d.PatchResponse(d.API, fmt.Sprintf("api/v1/domain/%s/version/active", domainName), recsToAdd)
	if err != nil {
		return err
	}

	_, err = d.handleHTTPError([]int{http.StatusNoContent}, resp)
	if err != nil {
		return err
	}

	return nil
}

func (d *OnlinedomainAPI) DeleteRecords(domainName string, rec []Record) error {

	var recsToAdd []RecordPatch

	for _, aRec := range rec {
		recsToAdd = append(recsToAdd, RecordPatch{
			Name:       aRec.Name,
			ChangeType: "DELETE",
			Type:       aRec.Type,
			Records: []PatchRecord{
				{
					Data: aRec.Data,
					TTL:  aRec.TTL,
				},
			},
		})
	}

	resp, err := d.PatchResponse(d.API, fmt.Sprintf("api/v1/domain/%s/version/active", domainName), recsToAdd)
	if err != nil {
		return err
	}

	_, err = d.handleHTTPError([]int{http.StatusNoContent}, resp)
	if err != nil {
		return err
	}

	return nil
}
