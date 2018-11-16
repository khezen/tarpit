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

func (i *requesters) put(requester string, resources *resources) {
	i.Lock()
	defer i.Unlock()
	i.resources[requester] = resources
}

func (i *requesters) get(requester string) *resources {
	i.RLock()
	defer i.RUnlock()
	resources, ok := i.resources[requester]
	if !ok {
		return nil
	}
	return resources
}

func (i *requesters) cleanup() {
	i.Lock()
	defer i.Unlock()
	for requester, resources := range i.resources {
		isEmpty := resources.cleanup()
		if isEmpty {
			delete(i.resources, requester)
		}
	}
}
