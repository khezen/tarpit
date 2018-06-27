package tarpit

import (
	"net/http"
	"time"
)

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
	tarpit := tarpit{
		unitDelay:   delay,
		chunkPeriod: chunkPeriod,
		isClosed:    false,
		close:       make(chan struct{}),
		monitoring:  newMonitoring(resetPeriod, defaultCleanupPeriod),
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
	calls := t.monitoring.get(ip, uri)
	remainingDuration := time.Duration(calls.count) * t.unitDelay
	if remainingDuration == 0 {
		return nil
	}
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
		remainingDuration = remainingDuration - t.chunkPeriod
		if remainingDuration <= 0 {
			return nil
		}
		// write a byte to prevent client timeout
		con.Write([]byte(" "))
	}
}

func (t *tarpit) Close() {
	t.isClosed = true
	t.close <- struct{}{}
}
