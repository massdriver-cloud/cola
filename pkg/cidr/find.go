package cidr

import (
	"fmt"
	"net"

	"github.com/apparentlymart/go-cidr/cidr"
)

// FindAvailableCIDR will find a CIDR range of specified desiredMask size within the
// rootCIDR given a list of already existing usedCIDRs.
func FindAvailableCIDR(rootCIDR *net.IPNet, desiredMask *net.IPMask, usedCIDRs []*net.IPNet) (*net.IPNet, error) {
	// if somehow the rootCIDR is within a used CIDR, then this is impossible
	for _, used := range usedCIDRs {
		if ContainsCIDR(used, rootCIDR) {
			// If the masks are equal this just means the the used CIDR is identical to the root CIDR, but still means theres no more space
			if EqualMask(&rootCIDR.Mask, &used.Mask) {
				return nil, fmt.Errorf("%w: a used CIDR matches the root CIDR", ErrNoAvailableCidr)
			}
			return nil, fmt.Errorf("%w: root CIDR is within a used CIDR", ErrInvalidInputRanges)
		}
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
		return nil, fmt.Errorf("%w: CIDR range collides with an existing CIDR", ErrNoAvailableCidr)
	}

	if EqualMask(desiredMask, &current.Mask) {
		if ContainsExistingCIDR(current, usedCIDRs) {
			return nil, fmt.Errorf("%w: CIDR range contains an existing CIDR", ErrNoAvailableCidr)
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

// ContainsExistingCIDR returns true if any of the usedCIDRs are contained within the currentCIDR,
// and false otherwise.
func ContainsExistingCIDR(currentCIDR *net.IPNet, usedCIDRs []*net.IPNet) bool {
	for _, usedCIDR := range usedCIDRs {
		if ContainsCIDR(currentCIDR, usedCIDR) {
			return true
		}
	}
	return false
}

// ContainsCIDR returns true if the childCIDR is contained within parentCIDR, and false otherwise.
// Comparison checking is inclusive, so identical CIDRs will return true.
func ContainsCIDR(parentCIDR *net.IPNet, childCIDR *net.IPNet) bool {
	firstIP, lastIP := cidr.AddressRange(childCIDR)
	if parentCIDR.Contains(firstIP) && parentCIDR.Contains(lastIP) {
		return true
	}
	return false
}

// ChildCIDRs will return the two child CIDRs from extending the mask 1 bit
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
