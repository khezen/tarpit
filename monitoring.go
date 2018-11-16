package tarpit

import "time"

type resourcePath string

type requests struct {
	count    int
	latestAt time.Time
}

type monitoring struct {
	resetPeriod time.Duration
	requesters  requesters
}

func newMonitoring(resetPeriod time.Duration) monitoring {
	return monitoring{
		resetPeriod: resetPeriod,
		requesters:  newRequesters(),
	}
}

func (m *monitoring) get(requester string, uri resourcePath) requests {
	resources := m.requesters.get(requester)
	if resources == nil {
		return requests{}
	}
	return resources.get(uri)
}

func (m *monitoring) increment(requester string, uri resourcePath) {
	resources := m.requesters.get(requester)
	if resources == nil {
		resources := newResources(m.resetPeriod)
		resources.increment(uri)
		m.requesters.put(requester, resources)
	} else {
		resources.increment(uri)
	}
}

func (m *monitoring) cleaner(cleanupPeriod time.Duration, stop chan struct{}) {
	ticker := time.NewTicker(cleanupPeriod)
	for {
		select {
		case <-ticker.C:
			m.requesters.cleanup()
		case <-stop:
			return
		}
	}
}
