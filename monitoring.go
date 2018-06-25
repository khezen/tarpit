package tarpit

import "time"

// Record -
type record struct {
	count    int
	latestAt time.Time
}

type monitoring struct {
	keepAlive, cleanupPeriod time.Duration
	records                  ipAddresses
}

func newMonitoring(keepAlive, cleanupPeriod time.Duration) monitoring {
	return monitoring{
		cleanupPeriod: cleanupPeriod,
		keepAlive:     keepAlive,
		records:       newstringAddresses(),
	}
}

func (m *monitoring) get(ip, uri string) record {
	resources := m.records.get(ip)
	if resources == nil {
		return record{}
	}
	return resources.get(uri)
}

func (m *monitoring) increment(ip, uri string) {
	resources := m.records.get(ip)
	if resources == nil {
		resources := newResources(m.keepAlive)
		resources.increment(uri)
		m.records.put(ip, resources)
	} else {
		resources.increment(uri)
	}
}

func (m *monitoring) cleaner(stop chan struct{}) {
	ticker := time.NewTicker(m.cleanupPeriod)
	for {
		select {
		case <-ticker.C:
			m.records.cleanup()
		case <-stop:
			return
		}
	}
}
