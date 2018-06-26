package tarpit

import (
	"errors"
	"net/http"
	"time"
)

// ErrClosedTarpit -
var ErrClosedTarpit = errors.New("ErrClosedTarpit")

// Interface -
// call Handle(w http.ResponseWriter, r *http.Request) if you want to tarpit an incoming connection.
type Interface interface {
	Handle(w http.ResponseWriter, r *http.Request) error
	Close()
}

// New creates a new tarpit interface - delay is the unit period used to delay incoming connections.
// Repeted calls to the same resource from the same IP multiply this value;
// The tarpit sends one byte of response every chunkPeriod to keep the client from timing out;
// you can disable this feature by setting chunkPeriod to <= 0;
// Once a given resources is not called from a given IP for more than resetPeriod, then the delay is reset.
func New(delay, chunkPeriod, resetPeriod time.Duration) Interface {
	cleanupPeriod := 15 * time.Minute
	tarpit := tarpit{
		unitDelay:   delay,
		chunkPeriod: chunkPeriod,
		isClosed:    false,
		close:       make(chan struct{}),
		monitoring:  newMonitoring(resetPeriod, cleanupPeriod),
	}
	go tarpit.monitoring.cleaner(tarpit.close)
	return &tarpit
}

type tarpit struct {
	unitDelay   time.Duration
	chunkPeriod time.Duration
	isClosed    bool
	close       chan struct{}
	monitoring  monitoring
}

func (t *tarpit) Handle(w http.ResponseWriter, r *http.Request) error {
	if t.isClosed {
		return ErrClosedTarpit
	}
	ip := getCallerIP(r)
	uri := getURI(r)
	defer t.monitoring.increment(ip, uri)
	rec := t.monitoring.get(ip, uri)
	delay := time.Duration(rec.count) * t.unitDelay
	if delay == 0 {
		return nil
	}
	remainingDuration := delay
	con, _, err := hijack(w)
	if err != nil {
		return err
	}
	var timer *time.Timer
	for {
		if remainingDuration > t.chunkPeriod {
			timer = time.NewTimer(t.chunkPeriod)
		} else {
			timer = time.NewTimer(remainingDuration)
		}
		<-timer.C
		// write a byte to prevent client timeout
		con.Write([]byte(" "))
		remainingDuration = remainingDuration - t.chunkPeriod
		if remainingDuration <= 0 {
			return nil
		}
	}
}

func (t *tarpit) Close() {
	t.isClosed = true
	t.close <- struct{}{}
}
