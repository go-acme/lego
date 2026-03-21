package internal

const (
	// ChangeTypeReplace Replace all records in this RRset with the provided ones.
	ChangeTypeReplace = `REPLACE`

	// ChangeTypeDelete Remove all records in this RRset.
	ChangeTypeDelete = `DELETE`

	// ChangeTypeExtend Add new records to the end of this RRset if not already present.
	ChangeTypeExtend = `EXTEND`

	// ChangeTypePrune Remove the specified record from the RRset if present.
	ChangeTypePrune = `PRUNE`
)

type Zone struct {
	ID                string   `json:"id,omitempty"`
	Name              string   `json:"name,omitempty"`
	NotifiedSerial    int      `json:"notified_serial,omitempty"`
	AlsoNotify        []string `json:"also_notify,omitempty"`
	AllowTransferKeys []string `json:"allow_transfer_keys,omitempty"`
	EndCustomer       string   `json:"endcustomer,omitempty"`
	RRSets            []RRSet  `json:"rrsets,omitempty"`
}

type RRSet struct {
	Name       string   `json:"name,omitempty"`
	Type       string   `json:"type,omitempty"`
	TTL        int      `json:"ttl"`
	ChangeType string   `json:"changetype,omitempty"`
	Records    []Record `json:"records,omitempty"`
}

type Record struct {
	Content  string `json:"content,omitempty"`
	Disabled bool   `json:"disabled"`
}

type PagedResponse[T any] struct {
	Data   T   `json:"data"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Total  int `json:"total"`
}

type Pager struct {
	Offset int `url:"offset,omitempty"`
	Limit  int `url:"limit,omitempty"`
}
