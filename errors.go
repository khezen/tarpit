package tarpit

import "errors"

var (
	// ErrHijackingUnsupported - webserver doesn't support hijacking
	ErrHijackingUnsupported = errors.New("ErrHijackingUnsupported - webserver doesn't support hijacking")
	// ErrClosedTarpit -
	ErrClosedTarpit = errors.New("ErrClosedTarpit")
)
