package tarpit

import (
	"net/http"
	"time"
)

// Interface - tarpit interface
// call Handle(w http.ResponseWriter, r *http.Request) if you want to tarpit an incoming connection.
type Interface interface {
	Handle(w http.ResponseWriter, r *http.Request) error
	Close()
}

// New -
func New(unitDelay, chunkPeriod, keepAlive, cleanupPeriod time.Duration) Interface {
	tarpit := tarpit{
		unitDelay:   unitDelay,
		chunkPeriod: chunkPeriod,
		close:       make(chan struct{}),
		monitoring:  newMonitoring(keepAlive, cleanupPeriod),
	}
	go tarpit.monitoring.cleaner(tarpit.close)
	return &tarpit
}

type tarpit struct {
	unitDelay   time.Duration
	chunkPeriod time.Duration
	close       chan struct{}
	monitoring  monitoring
}

func (t *tarpit) Handle(w http.ResponseWriter, r *http.Request) error {
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
	t.close <- struct{}{}
}
