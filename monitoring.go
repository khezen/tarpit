package tarpit

import "time"

type resourcePath string

// Record -
type record struct {
	count    int
	latestAt time.Time
}

type monitoring struct {
	resetPeriod time.Duration
	records     keys
}

func newMonitoring(resetPeriod time.Duration) monitoring {
	return monitoring{
		resetPeriod: resetPeriod,
		records:     newstringAddresses(),
	}
}

func (m *monitoring) get(key string, uri resourcePath) record {
	resources := m.records.get(key)
	if resources == nil {
		return record{}
	}
	return resources.get(uri)
}

func (m *monitoring) increment(key string, uri resourcePath) {
	resources := m.records.get(key)
	if resources == nil {
		resources := newResources(m.resetPeriod)
		resources.increment(uri)
		m.records.put(key, resources)
	} else {
		resources.increment(uri)
	}
}

func (m *monitoring) cleaner(cleanupPeriod time.Duration, stop chan struct{}) {
	ticker := time.NewTicker(cleanupPeriod)
	for {
		select {
		case <-ticker.C:
			m.records.cleanup()
		case <-stop:
			return
		}
	}
}
