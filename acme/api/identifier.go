package api

import (
	"cmp"
	"maps"
	"net"
	"slices"

	"github.com/go-acme/lego/v4/acme"
)

func createIdentifiers(domains []string) []acme.Identifier {
	uniqIdentifiers := make(map[string]acme.Identifier)

	for _, domain := range domains {
		if _, ok := uniqIdentifiers[domain]; ok {
			continue
		}

		ident := acme.Identifier{Value: domain, Type: "dns"}

		if net.ParseIP(domain) != nil {
			ident.Type = "ip"
		}

		uniqIdentifiers[domain] = ident
	}

	return slices.AppendSeq(make([]acme.Identifier, 0, len(uniqIdentifiers)), maps.Values(uniqIdentifiers))
}

// compareIdentifiers compares 2 slices of [acme.Identifier].
func compareIdentifiers(a, b []acme.Identifier) int {
	// Clones slices to avoid modifying original slices.
	right := slices.Clone(a)
	left := slices.Clone(b)

	slices.SortStableFunc(right, compareIdentifier)
	slices.SortStableFunc(left, compareIdentifier)

	return slices.CompareFunc(right, left, compareIdentifier)
}

func compareIdentifier(right, left acme.Identifier) int {
	return cmp.Or(
		cmp.Compare(right.Type, left.Type),
		cmp.Compare(right.Value, left.Value),
	)
}
