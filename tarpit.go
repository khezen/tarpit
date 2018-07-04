package tarpit

import (
	"errors"
	"net/http"
	"time"
)

// Interface -
// call Tar(w http.ResponseWriter, r *http.Request) to slow down repeted connection to the same resource.
type Interface interface {
	Tar(w http.ResponseWriter, r *http.Request) error
	Close()
}

var (
	// ErrClosedTarpit -
	ErrClosedTarpit = errors.New("ErrClosedTarpit")
)

const (
	// DefaultDelay - 1s
	DefaultDelay = time.Second
	// DefaultResetPeriod - 1m
	DefaultResetPeriod   = time.Minute
	defaultCleanupPeriod = 5 * time.Minute
)

// New creates a new tarpit interface - delay is the unit period used to delay incoming connections.
// Repeted calls to the same resource from the same IP multiply this value;
// Once a given resources is not called from a given IP for more than resetPeriod, then the delay is reset.
func New(delay, resetPeriod time.Duration) Interface {
	tarpit := tarpit{
		unitDelay:  delay,
		isClosed:   false,
		close:      make(chan struct{}),
		monitoring: newMonitoring(resetPeriod),
	}
	go tarpit.monitoring.cleaner(defaultCleanupPeriod, tarpit.close)
	return &tarpit
}

type tarpit struct {
	unitDelay  time.Duration
	isClosed   bool
	close      chan struct{}
	monitoring monitoring
}

func (t *tarpit) Tar(w http.ResponseWriter, r *http.Request) error {
	if t.isClosed {
		return ErrClosedTarpit
	}
	ip := getCallerIP(r)
	uri := getURI(r)
	defer t.monitoring.increment(ip, uri)
	calls := t.monitoring.get(ip, uri)
	remainingDuration := time.Duration(calls.count) * t.unitDelay
	if remainingDuration == 0 {
		return nil
	}
	timer := time.NewTimer(remainingDuration)
	<-timer.C
	return nil
}

func (t *tarpit) Close() {
	t.isClosed = true
	t.close <- struct{}{}
}
