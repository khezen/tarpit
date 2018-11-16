package tarpit

import (
	"sync"
	"time"
)

type resources struct {
	sync.Mutex
	resetPeriod time.Duration
	requests    map[resourcePath]requests
}

func newResources(resetPeriod time.Duration) *resources {
	return &resources{
		sync.Mutex{},
		resetPeriod,
		make(map[resourcePath]requests),
	}
}

func (r *resources) increment(uri resourcePath) {
	now := time.Now().UTC()
	r.Lock()
	defer r.Unlock()
	rqs, ok := r.requests[uri]
	if !ok {
		rqs = requests{
			count:    1,
			latestAt: now,
		}
	} else {
		rqs.count++
		rqs.latestAt = now
	}
	r.requests[uri] = rqs
}

func (r *resources) get(uri resourcePath) requests {
	now := time.Now().UTC()
	r.Lock()
	defer r.Unlock()
	rqs, ok := r.requests[uri]
	if !ok {
		return requests{}
	}
	if rqs.count > 0 && rqs.latestAt.UnixNano()+int64(r.resetPeriod) <= now.UnixNano() {
		delete(r.requests, uri)
		return requests{}
	}
	return rqs
}

func (r *resources) cleanup() (isEmpty bool) {
	r.Lock()
	defer r.Unlock()
	now := time.Now().UTC()
	isEmpty = true
	wg := sync.WaitGroup{}
	wg.Add(len(r.requests))
	for uri, rqs := range r.requests {
		go func(uri resourcePath, rqs requests) {
			if rqs.latestAt.UnixNano()+int64(r.resetPeriod) <= now.UnixNano() {
				delete(r.requests, uri)
			} else {
				isEmpty = false
			}
			wg.Done()
		}(uri, rqs)
	}
	wg.Wait()
	return isEmpty
}
