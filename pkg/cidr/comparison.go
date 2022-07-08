package cidr

import "net"

func EqualCIDRs(x *net.IPNet, y *net.IPNet) bool {
	return x.IP.Equal(y.IP) && EqualMask(&x.Mask, &y.Mask)
}

func EqualMask(x *net.IPMask, y *net.IPMask) bool {
	xOnes, xBits := x.Size()
	yOnes, yBits := y.Size()
	return xOnes == yOnes && xBits == yBits
}

func SmallerMask(smaller *net.IPMask, larger *net.IPMask) bool {
	smallerOnes, _ := smaller.Size()
	largerOnes, _ := larger.Size()

	return smallerOnes > largerOnes
}
