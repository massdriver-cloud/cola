package cidr

import (
	"errors"
)

var (
	ErrNoAvailableCIDR    = errors.New("unable to find available CIDR range")
	ErrInvalidInputRanges = errors.New("input ranges invalid")
)
