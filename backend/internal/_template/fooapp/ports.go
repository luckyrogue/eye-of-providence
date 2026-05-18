package fooapp

import "context"

// Clock — example application port (non-domain infrastructure).
type Clock interface {
	NowUTC() context.Context
}
