package analytics

import (
	"context"
	"sync"
	"testing"
	"time"
)

type mockStore struct {
	mu             sync.Mutex
	purgeCalls     int
	aggregateCalls int
}

func (m *mockStore) RecordEvent(_ context.Context, _ Event) error { return nil }
func (m *mockStore) QueryPopularity(_ context.Context, _ string) (*PopularitySummary, error) {
	return &PopularitySummary{}, nil
}
func (m *mockStore) Close() error { return nil }

func (m *mockStore) PurgeOldEvents(_ context.Context, _ time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.purgeCalls++
	return nil
}

func (m *mockStore) AggregateDailyPopularity(_ context.Context, _ time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.aggregateCalls++
	return nil
}

func (m *mockStore) getPurgeCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.purgeCalls
}

func (m *mockStore) getAggregateCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.aggregateCalls
}

func TestStartRetentionJobs(t *testing.T) {
	store := &mockStore{}
	ctx, cancel := context.WithCancel(context.Background())

	StartRetentionJobs(ctx, store, RetentionConfig{
		RetainEventsDays: 30,
		RunInterval:      50 * time.Millisecond,
	})

	// Wait for at least one tick.
	time.Sleep(200 * time.Millisecond)
	cancel()

	// Give goroutine time to stop.
	time.Sleep(50 * time.Millisecond)

	if store.getPurgeCalls() < 1 {
		t.Errorf("purge calls = %d, want >= 1", store.getPurgeCalls())
	}
	if store.getAggregateCalls() < 1 {
		t.Errorf("aggregate calls = %d, want >= 1", store.getAggregateCalls())
	}
}

func TestRunMaintenance(t *testing.T) {
	store := &mockStore{}
	ctx := context.Background()

	runMaintenance(ctx, store, 90)

	if store.getPurgeCalls() != 1 {
		t.Errorf("purge calls = %d, want 1", store.getPurgeCalls())
	}
	if store.getAggregateCalls() != 1 {
		t.Errorf("aggregate calls = %d, want 1", store.getAggregateCalls())
	}
}
