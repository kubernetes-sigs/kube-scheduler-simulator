package util

import (
	"context"
	"runtime"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type ErrGroupWithSemaphore struct {
	Grp *errgroup.Group
	Sem *semaphore.Weighted
}

func NewErrGroupWithSemaphore(ctx context.Context) ErrGroupWithSemaphore {
	g, _ := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.GOMAXPROCS(0)))
	return ErrGroupWithSemaphore{
		Grp: g,
		Sem: sem,
	}
}
