package cidr

import (
	"errors"
)

var (
	ErrNoAvailableCidr    = errors.New("unable to find available CIDR range")
	ErrInvalidInputRanges = errors.New("input ranges invalid")
)
