package egoscale

import (
	"context"
	"fmt"

	"github.com/jinzhu/copier"
)

// Get fetches the resource
func (so *ServiceOffering) Get(ctx context.Context, client *Client) error {

	resp, err := client.List(so)
	if err != nil {
		return err
	}

	count := len(resp)
	if count == 0 {
		return &ErrorResponse{
			ErrorCode: ParamError,
			ErrorText: fmt.Sprintf("ServiceOffering not found. %#v", so),
		}
	} else if count > 1 {
		return fmt.Errorf("More than one ServiceOffering was found")
	}

	return copier.Copy(so, resp[0].(*ServiceOffering))
}

// ListRequest builds the ListSecurityGroups request
func (so *ServiceOffering) ListRequest() (ListCommand, error) {
	req := &ListServiceOfferings{
		ID:           so.ID,
		DomainID:     so.DomainID,
		IsSystem:     &so.IsSystem,
		Name:         so.Name,
		Restricted:   &so.Restricted,
		SystemVMType: so.SystemVMType,
	}

	return req, nil
}

// name returns the CloudStack API command name
func (*ListServiceOfferings) name() string {
	return "listServiceOfferings"
}

func (*ListServiceOfferings) response() interface{} {
	return new(ListServiceOfferingsResponse)
}

// SetPage sets the current page
func (lso *ListServiceOfferings) SetPage(page int) {
	lso.Page = page
}

// SetPageSize sets the page size
func (lso *ListServiceOfferings) SetPageSize(pageSize int) {
	lso.PageSize = pageSize
}

func (*ListServiceOfferings) each(resp interface{}, callback IterateItemFunc) {
	sos, ok := resp.(*ListServiceOfferingsResponse)
	if !ok {
		callback(nil, fmt.Errorf("ListServiceOfferingsResponse expected, got %t", resp))
		return
	}

	for i := range sos.ServiceOffering {
		if !callback(&sos.ServiceOffering[i], nil) {
			break
		}
	}
}
