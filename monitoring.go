package tarpit

import "time"

type ipAddress string

type resourcePath string

// Record -
type record struct {
	count    int
	latestAt time.Time
}

type monitoring struct {
	resetPeriod time.Duration
	records     ipAddresses
}

func newMonitoring(resetPeriod time.Duration) monitoring {
	return monitoring{
		resetPeriod: resetPeriod,
		records:     newstringAddresses(),
	}
}

func (m *monitoring) get(ip ipAddress, uri resourcePath) record {
	resources := m.records.get(ip)
	if resources == nil {
		return record{}
	}
	return resources.get(uri)
}

func (m *monitoring) increment(ip ipAddress, uri resourcePath) {
	resources := m.records.get(ip)
	if resources == nil {
		resources := newResources(m.resetPeriod)
		resources.increment(uri)
		m.records.put(ip, resources)
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
