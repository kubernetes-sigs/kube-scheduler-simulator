package util

import (
	"context"
	"runtime"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"golang.org/x/xerrors"
)

type SemaphoredErrGroup struct {
	g *errgroup.Group
	s *semaphore.Weighted
}

func NewErrGroupWithSemaphore(ctx context.Context) *SemaphoredErrGroup {
	g, _ := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.GOMAXPROCS(0)))
	return &SemaphoredErrGroup{
		g: g,
		s: sem,
	}
}

func (e *SemaphoredErrGroup) Go(fn func() error) error {
	ctx := context.Background()
	if err := e.s.Acquire(ctx, 1); err != nil {
		return xerrors.Errorf("acquire semaphore: %w", err)
	}

	e.g.Go(func() error {
		defer e.s.Release(1)
		return fn()
	})
	return nil
}

func (e *SemaphoredErrGroup) Wait() error {
	return e.g.Wait()
}
