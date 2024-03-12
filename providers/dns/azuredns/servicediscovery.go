package azuredns

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
)

type ServiceDiscoveryZone struct {
	Name           string
	SubscriptionID string
	ResourceGroup  string
}

const ResourceGraphQuery = `
resources
| where type =~ "%s"
%s
| project subscriptionId, resourceGroup, name
`

const ResourceGraphQueryOptionsTop = 1000

func discoverDnsZones(config *Config, credentials azcore.TokenCredential) (map[string]ServiceDiscoveryZone, error) {
	ctx := context.Background()
	zones := map[string]ServiceDiscoveryZone{}

	resourceType := "microsoft.network/dnszones"
	if config.PrivateZone {
		resourceType = "microsoft.network/privatednszones"
	}

	resourceGraphConditions := []string{}
	if config.SubscriptionID != "" {
		resourceGraphConditions = append(
			resourceGraphConditions,
			fmt.Sprintf(`| where subscriptionId =~ "%s"`, config.SubscriptionID),
		)
	}
	if config.ResourceGroup != "" {
		resourceGraphConditions = append(
			resourceGraphConditions,
			fmt.Sprintf(`| where resourceGroup =~ "%s"`, config.ResourceGroup),
		)
	}

	resourceGraphQuery := fmt.Sprintf(
		ResourceGraphQuery,
		resourceType,
		strings.Join(resourceGraphConditions, "\n"),
	)

	options := arm.ClientOptions{
		ClientOptions: azcore.ClientOptions{
			Cloud: config.Environment,
		},
	}

	resourceGraphClient, err := armresourcegraph.NewClient(credentials, &options)
	if err != nil {
		return zones, err
	}

	requestQueryTop := int32(ResourceGraphQueryOptionsTop)
	requestQuerySkip := int32(0)

	// Set options
	resultFormat := armresourcegraph.ResultFormatObjectArray
	requestOptions := armresourcegraph.QueryRequestOptions{
		ResultFormat: &resultFormat,
		Top:          &requestQueryTop,
		Skip:         &requestQuerySkip,
	}

	for {
		// Create the query request
		request := armresourcegraph.QueryRequest{
			Query:   &resourceGraphQuery,
			Options: &requestOptions,
		}

		var result, queryErr = resourceGraphClient.Resources(ctx, request, nil)
		if queryErr != nil {
			return zones, queryErr
		}

		if resultList, ok := result.Data.([]interface{}); ok {
			for _, row := range resultList {
				if rowData, ok := row.(map[string]interface{}); ok {
					if zoneName, ok := rowData["name"].(string); ok {
						if _, exists := zones[zoneName]; exists {
							return zones, fmt.Errorf(`found duplicate dns zone "%s"`, zoneName)
						}

						zones[zoneName] = ServiceDiscoveryZone{
							Name:           zoneName,
							ResourceGroup:  rowData["resourceGroup"].(string),
							SubscriptionID: rowData["subscriptionId"].(string),
						}
					}
				}
			}
		} else {
			// got invalid or empty data, skipping
			break
		}

		*requestOptions.Skip += requestQueryTop
		if result.TotalRecords != nil {
			if int64(*requestOptions.Skip) >= *result.TotalRecords {
				break
			}
		}
	}

	return zones, nil
}
