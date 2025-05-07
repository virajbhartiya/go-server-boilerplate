package jobs

import (
	"context"
	"sync"

	"go-server-boilerplate/internal/pkg/logger"

	"go.uber.org/zap"
)

// Dispatcher manages the worker pool and dispatches jobs
type Dispatcher struct {
	workerPool  []*Worker
	maxWorkers  int
	jobQueue    chan Job
	wg          *sync.WaitGroup
	scheduledWg *sync.WaitGroup
	shutdown    chan bool
	isRunning   bool
	mu          sync.Mutex
}

// NewDispatcher creates a new dispatcher
func NewDispatcher(maxWorkers int) *Dispatcher {
	jobQueue := make(chan Job, 100)
	return &Dispatcher{
		workerPool:  make([]*Worker, 0, maxWorkers),
		maxWorkers:  maxWorkers,
		jobQueue:    jobQueue,
		wg:          &sync.WaitGroup{},
		scheduledWg: &sync.WaitGroup{},
		shutdown:    make(chan bool),
		isRunning:   false,
	}
}

// Start starts the dispatcher and its workers
func (d *Dispatcher) Start() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.isRunning {
		return
	}

	logger.Info("Starting job dispatcher", zap.Int("max_workers", d.maxWorkers))

	// Start workers
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(i, d.jobQueue, d.wg)
		d.workerPool = append(d.workerPool, worker)
		d.wg.Add(1)
		worker.Start()
	}

	d.isRunning = true
}

// Stop stops the dispatcher and its workers
func (d *Dispatcher) Stop(ctx context.Context) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isRunning {
		return
	}

	logger.Info("Stopping job dispatcher")

	// Stop all workers
	for _, worker := range d.workerPool {
		worker.Stop()
	}

	// Wait for all workers to finish
	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()

	// Wait for either context cancellation or workers to finish
	select {
	case <-ctx.Done():
		logger.Warn("Context cancelled before all workers finished")
	case <-done:
		logger.Info("All workers finished successfully")
	}

	d.isRunning = false
}

// DispatchJob dispatches a job to be processed by a worker
func (d *Dispatcher) DispatchJob(job Job) {
	if !d.isRunning {
		logger.Warn("Cannot dispatch job: dispatcher is not running",
			zap.String("job_id", job.ID()),
			zap.String("job_name", job.Name()),
		)
		return
	}

	logger.Info("Dispatching job",
		zap.String("job_id", job.ID()),
		zap.String("job_name", job.Name()),
	)

	d.jobQueue <- job
}

// RunJob executes a job immediately in the current goroutine
func (d *Dispatcher) RunJob(ctx context.Context, job Job) error {
	logger.Info("Running job immediately",
		zap.String("job_id", job.ID()),
		zap.String("job_name", job.Name()),
	)

	return job.Execute(ctx)
}
