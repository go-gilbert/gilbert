package spec

import "errors"

var (
	ErrVersionMissing     = errors.New("missing version attribute")
	ErrUnsupportedVersion = errors.New("unsupported file version")
)
