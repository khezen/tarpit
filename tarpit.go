// * simple HTTP middleware that purposely delays incoming connections
// * repeted calls to a given resource increase the delay
// * enable TCP keep alive to keep the client from timing out

// One typical use case is to protect authentication from brute force.

// ## example

// The following example applies tarpit based on IP address. It is possible to apply tarpit based on any data provided in the request.

// ```golang

// package main

// import (
//     "net/http"
//     "github.com/khezen/tarpit"
// )

// var tarpitMiddleware = tarpit.New(tarpit.DefaultFreeCallsCount, tarpit.DefaultDelay, tarpit.DefaultResetPeriod)

// func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
//       if r.Method != http.MethodGet{
//          w.WriteHeader(http.StatusMethodNotAllowed)
//          return
//     }
//     w.Write([]byte("OK"))
// }

// func handleGetMedicine(w http.ResponseWriter, r *http.Request) {
//     if r.Method != http.MethodGet{
//          w.WriteHeader(http.StatusMethodNotAllowed)
//          return
//     }
//     ipAddr := r.Header.Get(httpHeaderXForwardedFor)
//     err := tarpitMiddleware.Tar(ipAddr, w, r)
//     if err != nil {
//         w.WriteHeader(http.StatusInternalServerError)
//         w.Write([]byte(err.Error()))
//         return
//     }
//     w.Write([]byte("Here are your pills"))
// }

// func main() {
//     http.HandleFunc("/drugs-store/v1/health", handleHealthCheck)
//     http.HandleFunc("/drugs-store/v1/medicine", handleGetMedicine)
//     writeTimeout := 30*time.Second
//     err := tarpit.ListenAndServe(":80", nil, writeTimeout)
//     if err != nil {
//         panic(err)
//     }
// }
// ```

package tarpit

import (
	"errors"
	"net/http"
	"time"
)

const (
	// DefaultFreeCallsCount -
	DefaultFreeCallsCount = 10
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
// call Tar(w http.ResponseWriter, r *http.Request) to slow down repeted connection to the same resource.
type Interface interface {
	Tar(key string, w http.ResponseWriter, r *http.Request) error
	Close()
}

// New creates a new tarpit interface - delay is the unit period used to delay incoming connections.
// Repeted calls to the same resource increase the {delay};
// The delay is apply after more than {freeCallsCount} repeted calls to a given resources;
// Once a given resources is not called for more than {resetPeriod}, then the delay is reset.
func New(freeCallsCount int, delay, resetPeriod time.Duration) Interface {
	tarpit := tarpit{
		unitDelay:      delay,
		freeCallsCount: freeCallsCount,
		isClosed:       false,
		close:          make(chan struct{}),
		monitoring:     newMonitoring(resetPeriod),
	}
	go tarpit.monitoring.cleaner(defaultCleanupPeriod, tarpit.close)
	return &tarpit
}

type tarpit struct {
	unitDelay      time.Duration
	freeCallsCount int
	isClosed       bool
	close          chan struct{}
	monitoring     monitoring
}

func (t *tarpit) Tar(key string, w http.ResponseWriter, r *http.Request) error {
	if t.isClosed {
		return ErrClosedTarpit
	}
	uri := getURI(r)
	defer t.monitoring.increment(key, uri)
	calls := t.monitoring.get(key, uri)
	remainingDuration := time.Duration(calls.count-t.freeCallsCount) * t.unitDelay
	if remainingDuration <= 0 {
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
