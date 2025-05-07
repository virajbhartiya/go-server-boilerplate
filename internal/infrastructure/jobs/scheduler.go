package jobs

import (
	"context"
	"sync"
	"time"

	"go-server-boilerplate/internal/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ScheduledJob represents a job that runs on a schedule
type ScheduledJob struct {
	id          string
	name        string
	interval    time.Duration
	executeFunc func(ctx context.Context) error
	ticker      *time.Ticker
	dispatcher  *Dispatcher
	ctx         context.Context
	cancel      context.CancelFunc
	isRunning   bool
	mu          sync.Mutex
}

// NewScheduledJob creates a new scheduled job
func NewScheduledJob(name string, interval time.Duration, executeFunc func(ctx context.Context) error, dispatcher *Dispatcher) *ScheduledJob {
	ctx, cancel := context.WithCancel(context.Background())
	return &ScheduledJob{
		id:          uuid.New().String(),
		name:        name,
		interval:    interval,
		executeFunc: executeFunc,
		dispatcher:  dispatcher,
		ctx:         ctx,
		cancel:      cancel,
		isRunning:   false,
	}
}

// Start starts the scheduled job
func (s *ScheduledJob) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return
	}

	logger.Info("Starting scheduled job",
		zap.String("job_id", s.id),
		zap.String("job_name", s.name),
		zap.Duration("interval", s.interval),
	)

	s.ticker = time.NewTicker(s.interval)
	s.isRunning = true

	go func() {
		// Run immediately on start
		s.dispatchJob()

		// Then run on schedule
		for {
			select {
			case <-s.ticker.C:
				s.dispatchJob()
			case <-s.ctx.Done():
				logger.Info("Scheduled job stopped",
					zap.String("job_id", s.id),
					zap.String("job_name", s.name),
				)
				return
			}
		}
	}()
}

// Stop stops the scheduled job
func (s *ScheduledJob) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return
	}

	logger.Info("Stopping scheduled job",
		zap.String("job_id", s.id),
		zap.String("job_name", s.name),
	)

	s.ticker.Stop()
	s.cancel()
	s.isRunning = false
}

// dispatchJob dispatches the job to the dispatcher
func (s *ScheduledJob) dispatchJob() {
	job := &jobWrapper{
		id:      uuid.New().String(),
		name:    s.name,
		execute: s.executeFunc,
	}
	s.dispatcher.DispatchJob(job)
}

// jobWrapper wraps a function in a Job interface
type jobWrapper struct {
	id      string
	name    string
	execute func(ctx context.Context) error
}

// Execute runs the job
func (j *jobWrapper) Execute(ctx context.Context) error {
	return j.execute(ctx)
}

// Name returns the name of the job
func (j *jobWrapper) Name() string {
	return j.name
}

// ID returns the unique identifier of the job
func (j *jobWrapper) ID() string {
	return j.id
}
