package transport

import "errors"

var (
	ErrNoPeer = errors.New("no peer found in gRPC context")
)
