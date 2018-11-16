package tarpit

import (
	"sync"
)

type requesters struct {
	sync.RWMutex
	resources map[string]*resources
}

func newRequesters() requesters {
	return requesters{
		sync.RWMutex{},
		make(map[string]*resources),
	}
}

func (r *requesters) put(requester string, resources *resources) {
	r.Lock()
	defer r.Unlock()
	r.resources[requester] = resources
}

func (r *requesters) get(requester string) *resources {
	r.RLock()
	defer r.RUnlock()
	resources, ok := r.resources[requester]
	if !ok {
		return nil
	}
	return resources
}

func (r *requesters) cleanup() {
	r.Lock()
	defer r.Unlock()
	rLen := len(r.resources)
	if rLen == 0 {
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(rLen)
	for requester, resrcs := range r.resources {
		go func(requester string, resrcs *resources) {
			isEmpty := resrcs.cleanup()
			if isEmpty {
				delete(r.resources, requester)
			}
			wg.Done()
		}(requester, resrcs)
	}
	wg.Wait()
}
