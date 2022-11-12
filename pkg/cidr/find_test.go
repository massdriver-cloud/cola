package cidr_test

import (
	"errors"
	"net"
	"testing"

	"github.com/massdriver-cloud/cola/pkg/cidr"
)

func TestFindAvailableCIDRs(t *testing.T) {
	type testData struct {
		name        string
		baseCIDR    string
		usedCIDRs   []string
		desiredMask net.IPMask
		want        string
		wantError   error
	}
	tests := []testData{
		{
			name:     "Comment example",
			baseCIDR: "10.0.0.0/16",
			usedCIDRs: []string{
				"10.0.0.0/18",
				"10.0.64.0/20",
				"10.0.80.0/24",
			},
			desiredMask: net.CIDRMask(21, 32),
			want:        "10.0.88.0/21",
			wantError:   nil,
		},
		{
			name:        "Basic",
			baseCIDR:    "10.0.0.0/16",
			usedCIDRs:   []string{},
			desiredMask: net.CIDRMask(24, 32),
			want:        "10.0.0.0/24",
			wantError:   nil,
		},
		{
			name:     "Simple Collision",
			baseCIDR: "10.0.0.0/16",
			usedCIDRs: []string{
				"10.0.0.0/24",
			},
			desiredMask: net.CIDRMask(24, 32),
			want:        "10.0.1.0/24",
			wantError:   nil,
		},
		{
			name:     "Collision Used Larger",
			baseCIDR: "10.0.0.0/16",
			usedCIDRs: []string{
				"10.0.0.0/23",
			},
			desiredMask: net.CIDRMask(24, 32),
			want:        "10.0.2.0/24",
			wantError:   nil,
		},
		{
			name:     "Collision Used Smaller",
			baseCIDR: "10.0.0.0/16",
			usedCIDRs: []string{
				"10.0.0.0/24",
			},
			desiredMask: net.CIDRMask(23, 32),
			want:        "10.0.2.0/23",
			wantError:   nil,
		},
		{
			name:     "Error invalid subnets",
			baseCIDR: "10.0.0.0/16",
			usedCIDRs: []string{
				"10.0.0.0/24",
				"10.1.0.0/24",
			},
			desiredMask: net.CIDRMask(23, 32),
			want:        "",
			wantError:   cidr.ErrInvalidInputRanges,
		},
		{
			name:        "Successful entire subnet",
			baseCIDR:    "10.0.0.0/16",
			usedCIDRs:   []string{},
			desiredMask: net.CIDRMask(16, 32),
			want:        "10.0.0.0/16",
			wantError:   nil,
		},
		{
			name:     "Error entire subnet",
			baseCIDR: "10.0.0.0/16",
			usedCIDRs: []string{
				"10.0.0.0/24",
			},
			desiredMask: net.CIDRMask(16, 32),
			want:        "",
			wantError:   cidr.ErrNoAvailableCIDR,
		},
		{
			name:     "Error full",
			baseCIDR: "10.0.0.0/16",
			usedCIDRs: []string{
				"10.0.0.0/18",
				"10.0.64.0/18",
				"10.0.128.0/18",
				"10.0.196.0/18",
			},
			desiredMask: net.CIDRMask(24, 32),
			want:        "",
			wantError:   cidr.ErrNoAvailableCIDR,
		},
		{
			name:        "Error Mask too large",
			baseCIDR:    "10.0.0.0/16",
			usedCIDRs:   []string{},
			desiredMask: net.CIDRMask(15, 32),
			want:        "",
			wantError:   cidr.ErrNoAvailableCIDR,
		},
		{
			name:        "baseCIDR is usedCIDR",
			baseCIDR:    "10.0.0.0/16",
			usedCIDRs:   []string{"10.0.0.0/16"},
			desiredMask: net.CIDRMask(24, 32),
			want:        "",
			wantError:   cidr.ErrNoAvailableCIDR,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, baseCIDR, _ := net.ParseCIDR(tc.baseCIDR)
			usedCIDRs := make([]*net.IPNet, len(tc.usedCIDRs))
			for i, usedCIDR := range tc.usedCIDRs {
				_, usedCIDR, _ := net.ParseCIDR(usedCIDR)
				usedCIDRs[i] = usedCIDR
			}
			got, err := cidr.FindAvailableCIDR(baseCIDR, &tc.desiredMask, usedCIDRs)
			if tc.wantError != nil {
				if err == nil {
					t.Fatalf("Expected error: %s, got nil", tc.wantError.Error())
				}
				if !errors.Is(err, tc.wantError) {
					t.Fatalf("Invalid error, want: %s, got %s,", tc.wantError.Error(), err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %s,", err.Error())
				}
				if got.String() != tc.want {
					t.Fatalf("want: %v, got: %v", tc.want, got.String())
				}
			}
		})
	}
}

func TestMatchesExistingCIDR(t *testing.T) {
	type testData struct {
		name        string
		currentCIDR string
		usedCIDRs   []string
		want        bool
	}
	tests := []testData{
		{
			name:        "Basic True",
			currentCIDR: "10.0.0.0/24",
			usedCIDRs: []string{
				"10.0.0.0/24",
			},
			want: true,
		},
		{
			name:        "Basic False",
			currentCIDR: "10.0.0.0/24",
			usedCIDRs:   []string{},
			want:        false,
		},
		{
			name:        "Multiple True",
			currentCIDR: "10.0.1.0/24",
			usedCIDRs: []string{
				"10.0.0.0/24",
				"10.0.1.0/24",
				"10.0.2.0/24",
			},
			want: true,
		},
		{
			name:        "Multiple False",
			currentCIDR: "10.0.4.0/24",
			usedCIDRs: []string{

				"10.0.0.0/24",
				"10.0.1.0/24",
				"10.0.2.0/24",
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, currentCIDR, _ := net.ParseCIDR(tc.currentCIDR)
			usedCIDRs := make([]*net.IPNet, len(tc.usedCIDRs))
			for i, usedCIDR := range tc.usedCIDRs {
				_, usedCIDR, _ := net.ParseCIDR(usedCIDR)
				usedCIDRs[i] = usedCIDR
			}
			got := cidr.MatchesExistingCIDR(currentCIDR, usedCIDRs)

			if got != tc.want {
				t.Fatalf("want: %v, got: %v", tc.want, got)
			}
		})
	}
}

func TestContainsExistingCIDR(t *testing.T) {
	type testData struct {
		name        string
		currentCIDR string
		usedCIDRs   []string
		want        bool
	}
	tests := []testData{
		{
			name:        "Basic True",
			currentCIDR: "10.0.0.0/20",
			usedCIDRs: []string{
				"10.0.0.0/24",
			},
			want: true,
		},
		{
			name:        "Basic False",
			currentCIDR: "10.0.0.0/20",
			usedCIDRs:   []string{},
			want:        false,
		},
		{
			name:        "Multiple True",
			currentCIDR: "10.0.0.0/20",
			usedCIDRs: []string{
				"10.0.15.0/24",
				"10.0.16.0/24",
				"10.0.17.0/24",
			},
			want: true,
		},
		{
			name:        "Multiple False",
			currentCIDR: "10.0.0.0/20",
			usedCIDRs: []string{
				"10.0.16.0/24",
				"10.0.17.0/24",
				"10.0.18.0/24",
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, currentCIDR, _ := net.ParseCIDR(tc.currentCIDR)
			usedCIDRs := make([]*net.IPNet, len(tc.usedCIDRs))
			for i, usedCIDR := range tc.usedCIDRs {
				_, usedCIDR, _ := net.ParseCIDR(usedCIDR)
				usedCIDRs[i] = usedCIDR
			}
			got := cidr.ContainsExistingCIDR(currentCIDR, usedCIDRs)

			if got != tc.want {
				t.Fatalf("want: %v, got: %v", tc.want, got)
			}
		})
	}
}

func TestContainsCIDR(t *testing.T) {
	type testData struct {
		name        string
		currentCIDR string
		testCIDR    string
		want        bool
	}
	tests := []testData{
		{
			name:        "Basic True",
			currentCIDR: "10.0.16.0/20",
			testCIDR:    "10.0.17.0/24",
			want:        true,
		},
		{
			name:        "Fully contains True",
			currentCIDR: "10.0.16.0/20",
			testCIDR:    "10.0.16.0/24",
			want:        true,
		},
		{
			name:        "Basic False",
			currentCIDR: "10.0.16.0/20",
			testCIDR:    "10.0.15.0/24",
			want:        false,
		},
		{
			name:        "Lower True",
			currentCIDR: "10.0.16.0/20",
			testCIDR:    "10.0.16.0/24",
			want:        true,
		},
		{
			name:        "Upper true",
			currentCIDR: "10.0.16.0/20",
			testCIDR:    "10.0.31.0/24",
			want:        true,
		},
		{
			name:        "Inverse False",
			currentCIDR: "10.0.16.0/20",
			testCIDR:    "10.0.0.0/18",
			want:        false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, currentCIDR, _ := net.ParseCIDR(tc.currentCIDR)
			_, testCIDR, _ := net.ParseCIDR(tc.testCIDR)

			got := cidr.ContainsCIDR(currentCIDR, testCIDR)

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

	got1, got2, err := cidr.ChildCIDRs(parent)
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
