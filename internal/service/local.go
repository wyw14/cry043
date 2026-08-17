package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Clock struct{}

func (Clock) Now() time.Time { return time.Now().UTC() }

type IDs struct{ n atomic.Uint64 }

func (i *IDs) NewID() string { return fmt.Sprintf("spec-%08d", i.n.Add(1)) }

type LocalScheduler struct {
	mu    sync.Mutex
	Tasks map[string]time.Time
}

func (s *LocalScheduler) Schedule(ctx context.Context, id string, at time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Tasks == nil {
		s.Tasks = map[string]time.Time{}
	}
	s.Tasks[id] = at
	return nil
}
