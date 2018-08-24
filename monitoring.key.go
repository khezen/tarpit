package tarpit

import (
	"sync"
)

type keys struct {
	sync.RWMutex
	records map[string]*resources
}

func newstringAddresses() keys {
	return keys{
		sync.RWMutex{},
		make(map[string]*resources),
	}
}

func (i *keys) put(key string, resources *resources) {
	i.Lock()
	defer i.Unlock()
	i.records[key] = resources
}

func (i *keys) get(key string) *resources {
	i.RLock()
	defer i.RUnlock()
	resources, ok := i.records[key]
	if !ok {
		return nil
	}
	return resources
}

func (i *keys) cleanup() {
	i.Lock()
	defer i.Unlock()
	for key, resources := range i.records {
		isEmpty := resources.cleanup()
		if isEmpty {
			delete(i.records, key)
		}
	}
}
