package azuredns

import (
	"bytes"
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/go-acme/lego/v4/providers/dns/internal/ptr"
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
		ResultFormat: to.Ptr(armresourcegraph.ResultFormatObjectArray),
		Top:          to.Ptr(ResourceGraphQueryOptionsTop),
		Skip:         to.Ptr[int32](0),
	}

	zones := map[string]ServiceDiscoveryZone{}

	for {
		// create the query request
		request := armresourcegraph.QueryRequest{
			Query:   to.Ptr(createGraphQuery(config)),
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
			if int64(ptr.Deref(requestOptions.Skip)) >= ptr.Deref(result.TotalRecords) {
				break
			}
		}
	}

	return zones, nil
}

func createGraphQuery(config *Config) string {
	buf := new(bytes.Buffer)
	buf.WriteString("\nresources\n")

	resourceType := ResourceGraphTypePublicDNSZone
	if config.PrivateZone {
		resourceType = ResourceGraphTypePrivateDNSZone
	}

	_, _ = fmt.Fprintf(buf, "| where type =~ %q\n", resourceType)

	if config.SubscriptionID != "" {
		_, _ = fmt.Fprintf(buf, "| where subscriptionId =~ %q\n", config.SubscriptionID)
	}

	if config.ResourceGroup != "" {
		_, _ = fmt.Fprintf(buf, "| where resourceGroup =~ %q\n", config.ResourceGroup)
	}

	if config.ServiceDiscoveryFilter != "" {
		_, _ = fmt.Fprintf(buf, "| %s\n", config.ServiceDiscoveryFilter)
	}

	buf.WriteString("| project subscriptionId, resourceGroup, name")

	return buf.String()
}
