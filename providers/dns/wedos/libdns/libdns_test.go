package libdns

import (
	"fmt"
	"testing"
)

func ExampleRelativeName() {
	fmt.Println(RelativeName("sub.example.com.", "example.com."))
	// Output: sub
}

func ExampleAbsoluteName() {
	fmt.Println(AbsoluteName("sub", "example.com."))
	// Output: sub.example.com.
}

func TestRelativeName(t *testing.T) {
	for i, test := range []struct {
		fqdn, zone string
		expect     string
	}{
		{
			fqdn:   "",
			zone:   "",
			expect: "",
		},
		{
			fqdn:   "",
			zone:   "example.com",
			expect: "",
		},
		{
			fqdn:   "example.com",
			zone:   "",
			expect: "example.com",
		},
		{
			fqdn:   "sub.example.com",
			zone:   "example.com",
			expect: "sub",
		},
		{
			fqdn:   "foo.bar.example.com",
			zone:   "bar.example.com",
			expect: "foo",
		},
		{
			fqdn:   "foo.bar.example.com",
			zone:   "example.com",
			expect: "foo.bar",
		},
		{
			fqdn:   "foo.bar.example.com.",
			zone:   "example.com.",
			expect: "foo.bar",
		},
		{
			fqdn:   "example.com",
			zone:   "example.net",
			expect: "example.com",
		},
	} {
		actual := RelativeName(test.fqdn, test.zone)
		if actual != test.expect {
			t.Errorf("Test %d: FQDN=%s ZONE=%s - expected '%s' but got '%s'",
				i, test.fqdn, test.zone, test.expect, actual)
		}
	}
}

func TestAbsoluteName(t *testing.T) {
	for i, test := range []struct {
		name, zone string
		expect     string
	}{
		{
			name:   "",
			zone:   "example.com",
			expect: "example.com",
		},
		{
			name:   "@",
			zone:   "example.com.",
			expect: "example.com.",
		},
		{
			name:   "www",
			zone:   "example.com.",
			expect: "www.example.com.",
		},
		{
			name:   "www",
			zone:   "example.com.",
			expect: "www.example.com.",
		},
		{
			name:   "www.",
			zone:   "example.com.",
			expect: "www.example.com.",
		},
		{
			name:   "foo.bar",
			zone:   "example.com.",
			expect: "foo.bar.example.com.",
		},
		{
			name:   "foo.bar.",
			zone:   "example.com.",
			expect: "foo.bar.example.com.",
		},
		{
			name:   "foo",
			zone:   "",
			expect: "foo",
		},
	} {
		actual := AbsoluteName(test.name, test.zone)
		if actual != test.expect {
			t.Errorf("Test %d: NAME=%s ZONE=%s - expected '%s' but got '%s'",
				i, test.name, test.zone, test.expect, actual)
		}
	}
}
