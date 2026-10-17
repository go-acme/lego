//go:build lego_dnslight

package dns

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProviderCodes(t *testing.T) {
	tags := getBuildTags()

	var expected []string

	for _, tag := range tags {
		if tag == buildTagDNSLight {
			continue
		}

		if !strings.HasPrefix(tag, buildTagDNSPrefix) {
			continue
		}

		expected = append(expected, strings.TrimPrefix(tag, buildTagDNSPrefix))
	}

	slices.Sort(expected)

	t.Attr("tags", strings.Join(tags, ","))

	codes := GetProviderCodes()
	slices.Sort(codes)

	assert.Equal(t, expected, codes)
}
