package cidr_test

import (
	"net"
	"testing"

	"github.com/massdriver-cloud/cola/pkg/cidr"
)

func TestEqualCIDRs(t *testing.T) {
	type testData struct {
		name string
		one  net.IPNet
		two  net.IPNet
		want bool
	}
	tests := []testData{
		{
			name: "Basic Equality",
			one:  net.IPNet{IP: net.IPv4(1, 2, 3, 4), Mask: net.CIDRMask(16, 32)},
			two:  net.IPNet{IP: net.IPv4(1, 2, 3, 4), Mask: net.CIDRMask(16, 32)},
			want: true,
		},
		{
			name: "Inequality IP",
			one:  net.IPNet{IP: net.IPv4(1, 2, 3, 4), Mask: net.CIDRMask(16, 32)},
			two:  net.IPNet{IP: net.IPv4(1, 2, 3, 5), Mask: net.CIDRMask(16, 32)},
			want: false,
		},
		{
			name: "Inequality Mask",
			one:  net.IPNet{IP: net.IPv4(1, 2, 3, 4), Mask: net.CIDRMask(16, 32)},
			two:  net.IPNet{IP: net.IPv4(1, 2, 3, 4), Mask: net.CIDRMask(15, 32)},
			want: false,
		},
		{
			name: "Equality v4 to v6",
			one:  net.IPNet{IP: net.IPv4(1, 2, 3, 4).To16(), Mask: net.CIDRMask(16, 32)},
			two:  net.IPNet{IP: net.IPv4(1, 2, 3, 4), Mask: net.CIDRMask(16, 32)},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := cidr.EqualCIDRs(&tc.one, &tc.two)

			if got != tc.want {
				t.Fatalf("want: %v, got: %v, cidr1: %v, cidr2: %v ", tc.want, got, tc.one.String(), tc.two.String())
			}
		})
	}
}
