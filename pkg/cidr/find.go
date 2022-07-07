package cidr

import (
	"errors"
	"net"

	"github.com/apparentlymart/go-cidr/cidr"
)

func FindAvailableCIDR(rootCidr *net.IPNet, desiredMask *net.IPMask, usedCidrs []*net.IPNet) (*net.IPNet, error) {
	// ensure inputs are valid
	err := cidr.VerifyNoOverlap(usedCidrs, rootCidr)
	if err != nil {
		return nil, err
	}

	// If the rootCidr mask and desired mask are equal, the only possible cidr is the entire range
	if EqualMask(&rootCidr.Mask, desiredMask) {
		if len(usedCidrs) == 0 {
			return rootCidr, nil
		} else {
			return nil, errors.New("unable to find available cidr")
		}
	}

	// If the root cidr has a smaller mask than the desired cidr, then this is impossible
	if SmallerMask(&rootCidr.Mask, desiredMask) {
		return nil, errors.New("desired mask is larger than available cidr")
	}

	result, err := evaluateChildren(rootCidr, desiredMask, usedCidrs)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("unable to find available cidr")
	}
	return result, nil
}

//                                Core Algorithm
// We're going to walk down the CIDR, each iteration splitting into the 2 children. For each child we check:
//   1. If we match an existing CIDR, skip it
//   2. If our mask fits the desired mask size then just make sure...
//   3. We don't contain an already existing CIDR (a /18 block might look good, til you check to see there are existing /20 blocks within it)
// If all of this passes, then we have found our CIDR. Otherwise, recursively keep searching.
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
func evaluateChildren(currentCidr *net.IPNet, desiredMask *net.IPMask, usedCidrs []*net.IPNet) (*net.IPNet, error) {
	child1, child2, err := ChildCidrs(currentCidr)
	if err != nil {
		return nil, err
	}

	for _, child := range []*net.IPNet{child1, child2} {
		if !MatchesExistingCidr(child, usedCidrs) {
			if EqualMask(desiredMask, &child.Mask) {
				if !ContainsExistingCidr(child, usedCidrs) {
					return child, nil
				}
			} else {
				result, err := evaluateChildren(child, desiredMask, usedCidrs)
				if err != nil {
					return nil, err
				}
				if result != nil {
					return result, nil
				}
			}
		}
	}

	return nil, nil
}

func MatchesExistingCidr(currentCidr *net.IPNet, usedCidrs []*net.IPNet) bool {
	for _, usedCidr := range usedCidrs {
		if EqualCidrs(currentCidr, usedCidr) {
			return true
		}
	}
	return false
}

func ContainsExistingCidr(currentCidr *net.IPNet, usedCidrs []*net.IPNet) bool {
	for _, usedCidr := range usedCidrs {
		if ContainsCidr(currentCidr, usedCidr) {
			return true
		}
	}
	return false
}

func ContainsCidr(parentCidr *net.IPNet, childCidr *net.IPNet) bool {
	firstIP, lastIP := cidr.AddressRange(childCidr)
	if parentCidr.Contains(firstIP) && parentCidr.Contains(lastIP) {
		return true
	}
	return false
}

func ChildCidrs(parent *net.IPNet) (*net.IPNet, *net.IPNet, error) {
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
