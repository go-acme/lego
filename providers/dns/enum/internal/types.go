package internal

import "fmt"

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%s: %s", a.Code, a.Message)
}

type AddRecordSetValueRequest struct {
	ProjectID string          `json:"projectId,omitempty"`
	ZoneID    string          `json:"zoneId,omitempty"`
	Name      string          `json:"name,omitempty"`
	Type      string          `json:"type,omitempty"`
	TTL       int             `json:"ttl,omitempty"`
	Value     *DNSRecordValue `json:"value,omitempty"`
}

type RemoveRecordSetValueRequest struct {
	ProjectID string `json:"projectId,omitempty"`
	ZoneID    string `json:"zoneId,omitempty"`
	Name      string `json:"name,omitempty"`
	Type      string `json:"type,omitempty"`
	Content   string `json:"content,omitempty"`
}

type GetZoneByNameRequest struct {
	ProjectID string `json:"projectId,omitempty"`
	Name      string `json:"name,omitempty"`
}

type RecordSetResponse struct {
	RecordSet *DNSRecordSet `json:"recordSet,omitempty"`
}

type GetZoneByNameResponse struct {
	Zone *DNSZone `json:"zone,omitempty"`
}

type ListProjectsResponse struct {
	Projects []*Project    `json:"projects,omitempty"`
	Page     *PageResponse `json:"page,omitempty"`
}

type PageResponse struct {
	NextPageToken string `json:"nextPageToken,omitempty"`
	TotalCount    int32  `json:"totalCount,omitempty"`
}

type DNSRecordSet struct {
	ID      string            `json:"id,omitempty"`
	ZoneID  string            `json:"zoneId,omitempty"`
	Name    string            `json:"name,omitempty"`
	Type    string            `json:"type,omitempty"`
	TTL     int32             `json:"ttl,omitempty"`
	Records []*DNSRecordValue `json:"records,omitempty"`
}

type DNSRecordValue struct {
	Content  string `json:"content,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
}

type Project struct {
	ID              string            `json:"id,omitempty"`
	OrgID           string            `json:"orgId,omitempty"`
	Name            string            `json:"name,omitempty"`
	Description     string            `json:"description,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	DefaultRegionID string            `json:"defaultRegionId,omitempty"`
	DefaultZoneID   string            `json:"defaultZoneId,omitempty"`
}

type DNSZone struct {
	ID                  string   `json:"id,omitempty"`
	ProjectID           string   `json:"projectId,omitempty"`
	Name                string   `json:"name,omitempty"`
	Nameservers         []string `json:"nameservers,omitempty"`
	Status              string   `json:"status,omitempty"`
	DeletionProtected   bool     `json:"deletionProtected,omitempty"`
	ObservedNameservers []string `json:"observedNameservers,omitempty"`
}
