package cidr

import (
	"fmt"
	"net"

	"github.com/apparentlymart/go-cidr/cidr"
)

func FindAvailableCIDR(rootCIDR *net.IPNet, desiredMask *net.IPMask, usedCIDRs []*net.IPNet) (*net.IPNet, error) {
	// ensure inputs are valid
	err := cidr.VerifyNoOverlap(usedCIDRs, rootCIDR)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInputRanges, err.Error())
	}

	// If the rootCIDR mask and desired mask are equal, the only possible cidr is the entire range
	if EqualMask(&rootCIDR.Mask, desiredMask) {
		if len(usedCIDRs) == 0 {
			return rootCIDR, nil
		}
		return nil, fmt.Errorf("%w: desired mask would consume the entire root CIDR range", ErrNoAvailableCidr)
	}

	// If the root cidr has a smaller mask than the desired cidr, then this is impossible
	if SmallerMask(&rootCIDR.Mask, desiredMask) {
		return nil, fmt.Errorf("%w: desired mask is larger than the root CIDR range", ErrNoAvailableCidr)
	}

	return evaluateCidr(rootCIDR, desiredMask, usedCIDRs)
}

//                                Core Algorithm
// We're going to walk down the CIDR, each iteration checking the current CIDR to see:
//   1. If we match an existing CIDR, skip it
//   2. If our mask fits the desired mask size then just make sure...
//   3. We don't contain an already existing CIDR (a /18 block might look good, til you check to see there are existing /20 blocks within it)
// If all of this passes, then we have found our CIDR. Otherwise, continue walking tree by check each child (always 2)
//
//                                    Example
// Root CIDR: 10.0.0.0/16
// Existing CIDRs:
//   10.0.0.0/18
//   10.0.64.0/20
//   10.0.80.0/24
// Desired Mask: /21
//                     This graph shows the walk. We always check child1 first
//                     Doubles lines (// or \\) show paths that are visited
//                     Single lines (/ or \) are never visited in this example
//
//                                   10.0.0.0/16
//                                  //         \
//                          (1)  child1       child2
//                                //             \
//                            10.0.0.0/17     10.0.128.0/17
//                           //        \\
//                   (2)  child1      child2   (3)
//                         //            \\
//                    10.0.0.0/18    10.0.64.0/18
//                     (conflict)    //        \
//                          (4)   child1      child2
//                                 //            \
//                           10.0.64.0/19      10.0.96.0/19
//                            //        \\
//                    (5)  child1      child2  (6)
//                          //            \\
//                    10.0.64.0/20      10.0.80.0/20
//                     (conflict)       //        \\
//                                   child1      child2
//                                    //            \\
//                        (7)    10.0.80.0/21    10.0.88.0/21  (8)
//                     (contains another subnet)   FOUND MATCH!
//
//                                 RESULT: 10.0.88.0/21
func evaluateCidr(current *net.IPNet, desiredMask *net.IPMask, usedCIDRs []*net.IPNet) (*net.IPNet, error) {
	if MatchesExistingCIDR(current, usedCIDRs) {
		return nil, fmt.Errorf("%w: CIDR range collides with an existing CIDR", ErrCidrAlreadyInUse)
	}

	if EqualMask(desiredMask, &current.Mask) {
		if ContainsExistingCIDR(current, usedCIDRs) {
			return nil, fmt.Errorf("%w: CIDR range contains an existing CIDR", ErrCidrAlreadyInUse)
		} else {
			// We found it!
			return current, nil
		}
	} else {
		child1, child2, err := ChildCIDRs(current)
		if err != nil {
			return nil, err
		}

		for _, child := range []*net.IPNet{child1, child2} {
			result, err := evaluateCidr(child, desiredMask, usedCIDRs)
			// if the result is set with no errors it means we found a CIDR, and should return it
			// all the way up the stack. Otherwise we no-op, which will either check the other child,
			// or return the catch-all error that no CIDRs exist in this current branch of the tree
			if result != nil && err == nil {
				return result, nil
			}
		}
	}

	return nil, fmt.Errorf("%w: searched all available ranges could not find space for requested mask", ErrNoAvailableCidr)
}

func MatchesExistingCIDR(currentCIDR *net.IPNet, usedCIDRs []*net.IPNet) bool {
	for _, usedCIDR := range usedCIDRs {
		if EqualCIDRs(currentCIDR, usedCIDR) {
			return true
		}
	}
	return false
}

func ContainsExistingCIDR(currentCIDR *net.IPNet, usedCIDRs []*net.IPNet) bool {
	for _, usedCIDR := range usedCIDRs {
		if ContainsCIDR(currentCIDR, usedCIDR) {
			return true
		}
	}
	return false
}

func ContainsCIDR(parentCIDR *net.IPNet, childCIDR *net.IPNet) bool {
	firstIP, lastIP := cidr.AddressRange(childCIDR)
	if parentCIDR.Contains(firstIP) && parentCIDR.Contains(lastIP) {
		return true
	}
	return false
}

func ChildCIDRs(parent *net.IPNet) (*net.IPNet, *net.IPNet, error) {
	child1, err := cidr.Subnet(parent, 1, 0)
	if err != nil {
		return nil, nil, err
	}
	child2, err := cidr.Subnet(parent, 1, 1)
	if err != nil {
		return nil, nil, err
	}
	return child1, child2, nil
}
