package egoscale

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jinzhu/copier"
)

// Get loads the given Affinity Group
func (ag *AffinityGroup) Get(ctx context.Context, client *Client) error {
	if ag.ID == "" && ag.Name == "" {
		return fmt.Errorf("An Affinity Group may only be searched using ID or Name")
	}

	resp, err := client.RequestWithContext(ctx, &ListAffinityGroups{
		ID:   ag.ID,
		Name: ag.Name,
	})

	if err != nil {
		return err
	}

	ags := resp.(*ListAffinityGroupsResponse)
	count := len(ags.AffinityGroup)
	if count == 0 {
		return &ErrorResponse{
			ErrorCode: ParamError,
			ErrorText: fmt.Sprintf("AffinityGroup not found id: %s, name: %s", ag.ID, ag.Name),
		}
	} else if count > 1 {
		return fmt.Errorf("More than one Affinity Group was found. Query; id: %s, name: %s", ag.ID, ag.Name)
	}

	return copier.Copy(ag, ags.AffinityGroup[0])
}

// Delete removes the given Affinity Group
func (ag *AffinityGroup) Delete(ctx context.Context, client *Client) error {
	if ag.ID == "" && ag.Name == "" {
		return fmt.Errorf("An Affinity Group may only be deleted using ID or Name")
	}

	req := &DeleteAffinityGroup{
		Account:  ag.Account,
		DomainID: ag.DomainID,
	}

	if ag.ID != "" {
		req.ID = ag.ID
	} else {
		req.Name = ag.Name
	}

	return client.BooleanRequestWithContext(ctx, req)
}

// name returns the CloudStack API command name
func (*CreateAffinityGroup) name() string {
	return "createAffinityGroup"
}

func (*CreateAffinityGroup) asyncResponse() interface{} {
	return new(CreateAffinityGroupResponse)
}

// name returns the CloudStack API command name
func (*UpdateVMAffinityGroup) name() string {
	return "updateVMAffinityGroup"
}

func (*UpdateVMAffinityGroup) asyncResponse() interface{} {
	return new(UpdateVMAffinityGroupResponse)
}

func (req *UpdateVMAffinityGroup) onBeforeSend(params *url.Values) error {
	// Either AffinityGroupIDs or AffinityGroupNames must be set
	if len(req.AffinityGroupIDs) == 0 && len(req.AffinityGroupNames) == 0 {
		params.Set("affinitygroupids", "")
	}
	return nil
}

// name returns the CloudStack API command name
func (*DeleteAffinityGroup) name() string {
	return "deleteAffinityGroup"
}

func (*DeleteAffinityGroup) asyncResponse() interface{} {
	return new(booleanResponse)
}

// name returns the CloudStack API command name
func (*ListAffinityGroups) name() string {
	return "listAffinityGroups"
}

func (*ListAffinityGroups) response() interface{} {
	return new(ListAffinityGroupsResponse)
}

// name returns the CloudStack API command name
func (*ListAffinityGroupTypes) name() string {
	return "listAffinityGroupTypes"
}

func (*ListAffinityGroupTypes) response() interface{} {
	return new(ListAffinityGroupTypesResponse)
}
