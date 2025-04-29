package api

import (
	"cmp"
	"slices"

	"github.com/go-acme/lego/v4/acme"
)

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
