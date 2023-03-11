package internal

import "fmt"

// ChangeResourceRecordSetsRequest is a complex type that contains change information for the resource record set.
type ChangeResourceRecordSetsRequest struct {
	XMLNs       string      `xml:"xmlns,attr"`
	ChangeBatch ChangeBatch `xml:"ChangeBatch"`
}

// ChangeResourceRecordSetsResponse is a complex type containing the response for the request.
type ChangeResourceRecordSetsResponse struct {
	ChangeInfo ChangeInfo `xml:"ChangeInfo"`
}

// GetChangeResponse is a complex type that contains the ChangeInfo element.
type GetChangeResponse struct {
	ChangeInfo ChangeInfo `xml:"ChangeInfo"`
}

type Error struct {
	Type    string `xml:"Type"`
	Message string `xml:"Message"`
	Code    string `xml:"Code"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%s(%s): %s", e.Type, e.Code, e.Message)
}

// ErrorResponse is the information for any errors.
type ErrorResponse struct {
	Error     Error  `xml:"Error"`
	RequestID string `xml:"RequestId"`
}

// ChangeBatch is the information for a change request.
type ChangeBatch struct {
	Changes Changes `xml:"Changes"`
	Comment string  `xml:"Comment"`
}

// Changes is array of Change.
type Changes struct {
	Change []Change `xml:"Change"`
}

// Change is the information for each resource record set that you want to change.
type Change struct {
	Action            string            `xml:"Action"`
	ResourceRecordSet ResourceRecordSet `xml:"ResourceRecordSet"`
}

// ResourceRecordSet is the information about the resource record set to create or delete.
type ResourceRecordSet struct {
	Name            string          `xml:"Name"`
	Type            string          `xml:"Type"`
	TTL             int             `xml:"TTL"`
	ResourceRecords ResourceRecords `xml:"ResourceRecords"`
}

// ResourceRecords is array of ResourceRecord.
type ResourceRecords struct {
	ResourceRecord []ResourceRecord `xml:"ResourceRecord"`
}

// ResourceRecord is the information specific to the resource record.
type ResourceRecord struct {
	Value string `xml:"Value"`
}

// ChangeInfo is A complex type that describes change information about changes made to your hosted zone.
type ChangeInfo struct {
	ID          string `xml:"Id"`
	Status      string `xml:"Status"`
	SubmittedAt string `xml:"SubmittedAt"`
}
