package tarpit

import (
	"sync"
	"time"
)

type resources struct {
	sync.Mutex
	resetPeriod time.Duration
	records     map[resourcePath]record
}

func newResources(resetPeriod time.Duration) *resources {
	return &resources{
		sync.Mutex{},
		resetPeriod,
		make(map[resourcePath]record),
	}
}

func (r *resources) increment(uri resourcePath) {
	now := time.Now().UTC()
	r.Lock()
	defer r.Unlock()
	rec, ok := r.records[uri]
	if !ok {
		rec = record{
			count:    1,
			latestAt: now,
		}
	} else {
		rec.count++
		rec.latestAt = now
	}
	r.records[uri] = rec
}

func (r *resources) get(uri resourcePath) record {
	now := time.Now().UTC()
	r.Lock()
	defer r.Unlock()
	rec, ok := r.records[uri]
	if !ok {
		return record{}
	}
	if rec.count > 0 && rec.latestAt.UnixNano()+int64(r.resetPeriod) <= now.UnixNano() {
		delete(r.records, uri)
		return record{}
	}
	return rec
}

func (r *resources) cleanup() (isEmpty bool) {
	now := time.Now().UTC()
	isEmpty = true
	r.Lock()
	defer r.Unlock()
	for uri, rec := range r.records {
		if rec.latestAt.UnixNano()+int64(r.resetPeriod) <= now.UnixNano() {
			delete(r.records, uri)
		} else {
			isEmpty = false
		}
	}
	return isEmpty
}
