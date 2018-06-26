package tarpit

import "time"

const (
	// DefaultDelay - 1s
	DefaultDelay = time.Second
	// DefaultChunkPeriod - 2s
	DefaultChunkPeriod = 2 * time.Second
	// DefaultResetPeriod - 1m
	DefaultResetPeriod   = time.Minute
	defaultCleanupPeriod = 5 * time.Minute
)
