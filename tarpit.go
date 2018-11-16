// Package tarpit :
//
// Simple HTTP middleware that purposely delays incoming connections.
// Repeted requests to a given resource increase the delay.
// Enables TCP keep alive to keep the client from timing out.
//
// One typical use case is to protect authentication from brute force.
//
// The following example applies tarpit based on IP address. It is possible to apply tarpit based on any data provided in the request.
//
// 	package main
//
// 	import (
// 		"net/http"
// 		"github.com/khezen/tarpit"
// 	)
//
// 	var tarpitMiddleware = tarpit.New(tarpit.DefaultFreeReqCount, tarpit.DefaultDelay, tarpit.DefaultResetPeriod)
//
// 	func handleGetMedicine(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodGet{
// 			w.WriteHeader(http.StatusMethodNotAllowed)
// 			return
// 		}
// 		ipAddr := r.Header.Get("X-Forwarded-For")
// 		err := tarpitMiddleware.Tar(ipAddr, w, r)
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			w.Write([]byte(err.Error()))
// 			return
// 		}
// 		w.Write([]byte("Here is your medicine"))
// 	}
//
// 	func main() {
// 		http.HandleFunc("/drugs-store/v1/medicine", handleGetMedicine)
// 		writeTimeout := 30*time.Second
// 		err := tarpit.ListenAndServe(":80", nil, writeTimeout)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
package tarpit

import (
	"errors"
	"math"
	"net/http"
	"time"
)

const (
	// DefaultFreeReqCount -
	DefaultFreeReqCount = 10
	// DefaultDelay - 1s
	DefaultDelay = time.Second
	// DefaultResetPeriod - 1m
	DefaultResetPeriod   = time.Minute
	defaultCleanupPeriod = 5 * time.Minute
)

var (
	// ErrClosedTarpit -
	ErrClosedTarpit = errors.New("ErrClosedTarpit")
)

// Interface -
// call Tar(requester string, w http.ResponseWriter, r *http.Request) to slow down repeted connection to the same resource.
type Interface interface {
	Tar(requester string, w http.ResponseWriter, r *http.Request) error
	Close()
}

// New creates a new tarpit interface - delay is the unit period used to delay incoming connections.
// Repeted requests to the same resource increase the {delay};
// The delay is apply after more than {freeReqCount} repeted requests to a given resources;
// Once a given resources is not called for more than {resetPeriod}, then the delay is reset.
func New(freeReqCount int, delay, resetPeriod time.Duration) Interface {
	tarpit := tarpit{
		unitDelay:    delay,
		freeReqCount: freeReqCount,
		isClosed:     false,
		close:        make(chan struct{}),
		monitoring:   newMonitoring(resetPeriod),
	}
	go tarpit.monitoring.cleaner(defaultCleanupPeriod, tarpit.close)
	return &tarpit
}

type tarpit struct {
	unitDelay    time.Duration
	freeReqCount int
	isClosed     bool
	close        chan struct{}
	monitoring   monitoring
}

func (t *tarpit) Tar(requester string, w http.ResponseWriter, r *http.Request) error {
	if t.isClosed {
		return ErrClosedTarpit
	}
	uri := resourcePath(r.URL.Path)
	defer t.monitoring.increment(requester, uri)
	requests := t.monitoring.get(requester, uri)
	if requests.count-t.freeReqCount <= 0 {
		return nil
	}
	remainingDuration := t.unitDelay * time.Duration(math.Pow(2, float64(requests.count-t.freeReqCount-1)))
	timer := time.NewTimer(remainingDuration)
	<-timer.C
	return nil
}

func (t *tarpit) Close() {
	t.isClosed = true
	t.close <- struct{}{}
}
