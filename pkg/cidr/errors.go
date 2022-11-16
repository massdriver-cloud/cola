package cidr

import (
	"errors"
)

var (
	ErrNoAvailableCidr    = errors.New("unable to find available CIDR range")
	ErrCidrAlreadyInUse   = errors.New("CIDR range already in use")
	ErrInvalidInputRanges = errors.New("input ranges invalid")
)
