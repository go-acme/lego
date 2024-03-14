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

const (
	ResourceGraphTypePublicDNSZone  = "microsoft.network/dnszones"
	ResourceGraphTypePrivateDNSZone = "microsoft.network/privatednszones"
)

const ResourceGraphQuery = `
resources
| where type =~ "%s"
%s
| project subscriptionId, resourceGroup, name`

const ResourceGraphQueryOptionsTop int32 = 1000

// discoverDNSZones finds all visible Azure DNS zones based on optional subscriptionID, resourceGroup and serviceDiscovery filter using Kusto query.
func discoverDNSZones(ctx context.Context, config *Config, credentials azcore.TokenCredential) (map[string]ServiceDiscoveryZone, error) {
	options := &arm.ClientOptions{
		ClientOptions: azcore.ClientOptions{
			Cloud: config.Environment,
		},
	}

	client, err := armresourcegraph.NewClient(credentials, options)
	if err != nil {
		return nil, err
	}

	// Set options
	requestOptions := &armresourcegraph.QueryRequestOptions{
		ResultFormat: pointer(armresourcegraph.ResultFormatObjectArray),
		Top:          pointer(ResourceGraphQueryOptionsTop),
		Skip:         pointer[int32](0),
	}

	zones := map[string]ServiceDiscoveryZone{}
	for {
		// create the query request
		request := armresourcegraph.QueryRequest{
			Query:   pointer(createGraphQuery(config)),
			Options: requestOptions,
		}

		result, err := client.Resources(ctx, request, nil)
		if err != nil {
			return zones, err
		}

		resultList, ok := result.Data.([]any)
		if !ok {
			// got invalid or empty data, skipping
			break
		}

		for _, row := range resultList {
			rowData, ok := row.(map[string]any)
			if !ok {
				continue
			}

			zoneName, ok := rowData["name"].(string)
			if !ok {
				continue
			}

			if _, exists := zones[zoneName]; exists {
				return zones, fmt.Errorf(`found duplicate dns zone "%s"`, zoneName)
			}

			zones[zoneName] = ServiceDiscoveryZone{
				Name:           zoneName,
				ResourceGroup:  rowData["resourceGroup"].(string),
				SubscriptionID: rowData["subscriptionId"].(string),
			}
		}

		*requestOptions.Skip += ResourceGraphQueryOptionsTop

		if result.TotalRecords != nil {
			if int64(deref(requestOptions.Skip)) >= deref(result.TotalRecords) {
				break
			}
		}
	}

	return zones, nil
}

func createGraphQuery(config *Config) string {
	var resourceGraphConditions []string

	// subscriptionID filter
	if config.SubscriptionID != "" {
		resourceGraphConditions = append(
			resourceGraphConditions,
			fmt.Sprintf(`| where subscriptionId =~ %q`, config.SubscriptionID),
		)
	}
	// resourceGroup filter
	if config.ResourceGroup != "" {
		resourceGraphConditions = append(
			resourceGraphConditions,
			fmt.Sprintf(`| where resourceGroup =~ %q`, config.ResourceGroup),
		)
	}
	// custom filter
	if config.ServiceDiscoveryFilter != "" {
		resourceGraphConditions = append(
			resourceGraphConditions,
			fmt.Sprintf(`| %s`, config.ServiceDiscoveryFilter),
		)
	}

	resourceType := ResourceGraphTypePublicDNSZone
	if config.PrivateZone {
		resourceType = ResourceGraphTypePrivateDNSZone
	}

	return fmt.Sprintf(
		ResourceGraphQuery,
		resourceType,
		strings.Join(resourceGraphConditions, "\n"),
	)
}
