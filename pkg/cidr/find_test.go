package cidr_test

import (
	"errors"
	"net"
	"testing"

	"github.com/massdriver-cloud/cola/pkg/cidr"
)

func TestFindAvailableCidrs(t *testing.T) {
	type testData struct {
		name        string
		baseCidr    string
		usedCidrs   []string
		desiredMask net.IPMask
		want        string
		wantError   error
	}
	tests := []testData{
		{
			name:        "Basic",
			baseCidr:    "10.0.0.0/16",
			usedCidrs:   []string{},
			desiredMask: net.CIDRMask(24, 32),
			want:        "10.0.0.0/24",
			wantError:   nil,
		},
		{
			name:     "Simple Collision",
			baseCidr: "10.0.0.0/16",
			usedCidrs: []string{
				"10.0.0.0/24",
			},
			desiredMask: net.CIDRMask(24, 32),
			want:        "10.0.1.0/24",
			wantError:   nil,
		},
		{
			name:     "Collision Used Larger",
			baseCidr: "10.0.0.0/16",
			usedCidrs: []string{
				"10.0.0.0/23",
			},
			desiredMask: net.CIDRMask(24, 32),
			want:        "10.0.2.0/24",
			wantError:   nil,
		},
		{
			name:     "Collision Used Smaller",
			baseCidr: "10.0.0.0/16",
			usedCidrs: []string{
				"10.0.0.0/24",
			},
			desiredMask: net.CIDRMask(23, 32),
			want:        "10.0.2.0/23",
			wantError:   nil,
		},
		{
			name:     "Error invalid subnets",
			baseCidr: "10.0.0.0/16",
			usedCidrs: []string{
				"10.0.0.0/24",
				"10.1.0.0/24",
			},
			desiredMask: net.CIDRMask(23, 32),
			want:        "",
			wantError:   errors.New("10.0.0.0/16 does not fully contain 10.1.0.0/24"),
		},
		{
			name:        "Successful entire subnet",
			baseCidr:    "10.0.0.0/16",
			usedCidrs:   []string{},
			desiredMask: net.CIDRMask(16, 32),
			want:        "10.0.0.0/16",
			wantError:   nil,
		},
		{
			name:     "Error entire subnet",
			baseCidr: "10.0.0.0/16",
			usedCidrs: []string{
				"10.0.0.0/24",
			},
			desiredMask: net.CIDRMask(16, 32),
			want:        "",
			wantError:   errors.New("unable to find available cidr"),
		},
		{
			name:     "Error full",
			baseCidr: "10.0.0.0/16",
			usedCidrs: []string{
				"10.0.0.0/18",
				"10.0.64.0/18",
				"10.0.128.0/18",
				"10.0.196.0/18",
			},
			desiredMask: net.CIDRMask(24, 32),
			want:        "",
			wantError:   errors.New("unable to find available cidr"),
		},
		{
			name:        "Error Mask too large",
			baseCidr:    "10.0.0.0/16",
			usedCidrs:   []string{},
			desiredMask: net.CIDRMask(15, 32),
			want:        "",
			wantError:   errors.New("desired mask is larger than available cidr"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, baseCidr, _ := net.ParseCIDR(tc.baseCidr)
			usedCidrs := make([]*net.IPNet, len(tc.usedCidrs))
			for i, usedCidr := range tc.usedCidrs {
				_, usedCidr, _ := net.ParseCIDR(usedCidr)
				usedCidrs[i] = usedCidr
			}
			got, err := cidr.FindAvailableCIDR(baseCidr, &tc.desiredMask, usedCidrs)
			if tc.wantError != nil {
				if err == nil {
					t.Fatalf("Expected error: %s, got nil", tc.wantError.Error())
				}
				if err.Error() != tc.wantError.Error() {
					t.Fatalf("Invalid error, want: %s, got %s,", tc.wantError.Error(), err.Error())
				}
			} else {
				if err != nil {
					if err.Error() != tc.wantError.Error() {
						t.Fatalf("Unexpected error: %s,", err.Error())
					}
				}
				if got.String() != tc.want {
					t.Fatalf("want: %v, got: %v", tc.want, got.String())
				}
			}
		})
	}
}

func TestMatchesExistingCidr(t *testing.T) {
	type testData struct {
		name        string
		currentCidr string
		usedCidrs   []string
		want        bool
	}
	tests := []testData{
		{
			name:        "Basic True",
			currentCidr: "10.0.0.0/24",
			usedCidrs: []string{
				"10.0.0.0/24",
			},
			want: true,
		},
		{
			name:        "Basic False",
			currentCidr: "10.0.0.0/24",
			usedCidrs:   []string{},
			want:        false,
		},
		{
			name:        "Multiple True",
			currentCidr: "10.0.1.0/24",
			usedCidrs: []string{
				"10.0.0.0/24",
				"10.0.1.0/24",
				"10.0.2.0/24",
			},
			want: true,
		},
		{
			name:        "Multiple False",
			currentCidr: "10.0.4.0/24",
			usedCidrs: []string{

				"10.0.0.0/24",
				"10.0.1.0/24",
				"10.0.2.0/24",
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, currentCidr, _ := net.ParseCIDR(tc.currentCidr)
			usedCidrs := make([]*net.IPNet, len(tc.usedCidrs))
			for i, usedCidr := range tc.usedCidrs {
				_, usedCidr, _ := net.ParseCIDR(usedCidr)
				usedCidrs[i] = usedCidr
			}
			got := cidr.MatchesExistingCidr(currentCidr, usedCidrs)

			if got != tc.want {
				t.Fatalf("want: %v, got: %v", tc.want, got)
			}
		})
	}
}

func TestContainsExistingCidr(t *testing.T) {
	type testData struct {
		name        string
		currentCidr string
		usedCidrs   []string
		want        bool
	}
	tests := []testData{
		{
			name:        "Basic True",
			currentCidr: "10.0.0.0/20",
			usedCidrs: []string{
				"10.0.0.0/24",
			},
			want: true,
		},
		{
			name:        "Basic False",
			currentCidr: "10.0.0.0/20",
			usedCidrs:   []string{},
			want:        false,
		},
		{
			name:        "Multiple True",
			currentCidr: "10.0.0.0/20",
			usedCidrs: []string{
				"10.0.15.0/24",
				"10.0.16.0/24",
				"10.0.17.0/24",
			},
			want: true,
		},
		{
			name:        "Multiple False",
			currentCidr: "10.0.0.0/20",
			usedCidrs: []string{
				"10.0.16.0/24",
				"10.0.17.0/24",
				"10.0.18.0/24",
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, currentCidr, _ := net.ParseCIDR(tc.currentCidr)
			usedCidrs := make([]*net.IPNet, len(tc.usedCidrs))
			for i, usedCidr := range tc.usedCidrs {
				_, usedCidr, _ := net.ParseCIDR(usedCidr)
				usedCidrs[i] = usedCidr
			}
			got := cidr.ContainsExistingCidr(currentCidr, usedCidrs)

			if got != tc.want {
				t.Fatalf("want: %v, got: %v", tc.want, got)
			}
		})
	}
}

func TestContainsCidr(t *testing.T) {
	type testData struct {
		name        string
		currentCidr string
		testCidr    string
		want        bool
	}
	tests := []testData{
		{
			name:        "Fully contains True",
			currentCidr: "10.0.16.0/20",
			testCidr:    "10.0.17.0/24",
			want:        true,
		},
		{
			name:        "Basic False",
			currentCidr: "10.0.16.0/20",
			testCidr:    "10.0.15.0/24",
			want:        false,
		},
		{
			name:        "Lower True",
			currentCidr: "10.0.16.0/20",
			testCidr:    "10.0.16.0/24",
			want:        true,
		},
		{
			name:        "Upper true",
			currentCidr: "10.0.16.0/20",
			testCidr:    "10.0.31.0/24",
			want:        true,
		},
		{
			name:        "Inverse False",
			currentCidr: "10.0.16.0/20",
			testCidr:    "10.0.0.0/18",
			want:        false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, currentCidr, _ := net.ParseCIDR(tc.currentCidr)
			_, testCidr, _ := net.ParseCIDR(tc.testCidr)

			got := cidr.ContainsCidr(currentCidr, testCidr)

			if got != tc.want {
				t.Fatalf("want: %v, got: %v", tc.want, got)
			}
		})
	}
}

func TestChildCIDRs(t *testing.T) {
	_, parent, _ := net.ParseCIDR("10.0.0.0/16")
	_, want1, _ := net.ParseCIDR("10.0.0.0/17")
	_, want2, _ := net.ParseCIDR("10.0.128.0/17")

	got1, got2, err := cidr.ChildCidrs(parent)
	if err != nil {
		t.Fatalf("%d, unexpected error", err)
	}

	if got1.String() != want1.String() {
		t.Fatalf("want: %v, got: %v", got1.String(), want1.String())
	}
	if got2.String() != want2.String() {
		t.Fatalf("want: %v, got: %v", got2.String(), want2.String())
	}
}
