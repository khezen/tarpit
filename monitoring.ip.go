package tarpit

import (
	"sync"
)

type ipAddresses struct {
	sync.RWMutex
	records map[string]*resources
}

func newstringAddresses() ipAddresses {
	return ipAddresses{
		sync.RWMutex{},
		make(map[string]*resources),
	}
}

func (i *ipAddresses) put(ip string, resources *resources) {
	i.Lock()
	defer i.Unlock()
	i.records[ip] = resources
}

func (i *ipAddresses) get(ip string) *resources {
	i.RLock()
	defer i.RUnlock()
	resources, ok := i.records[ip]
	if !ok {
		return nil
	}
	return resources
}

func (i *ipAddresses) cleanup() {
	i.Lock()
	defer i.Unlock()
	for ip, resources := range i.records {
		isEmpty := resources.cleanup()
		if isEmpty {
			delete(i.records, ip)
		}
	}
}
