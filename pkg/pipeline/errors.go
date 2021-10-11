package pipeline

import "errors"

var (
	ErrNotSupported        = errors.New("cannot add or remove rtmp outputs to non-stream recordings")
	ErrGhostPadFailed      = errors.New("failed to add ghost pad to bin")
	ErrOutputAlreadyExists = errors.New("output already exists")
	ErrOutputNotFound      = errors.New("output not found")
)
