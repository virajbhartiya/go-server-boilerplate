package jobs

import (
	"context"
	"sync"
	"time"

	"go-server-boilerplate/internal/pkg/logger"

	"go.uber.org/zap"
)

// Job represents a background job to be processed
type Job interface {
	// Execute runs the job
	Execute(ctx context.Context) error
	// Name returns the name of the job
	Name() string
	// ID returns the unique identifier of the job
	ID() string
}

// Worker processes jobs in the background
type Worker struct {
	id         int
	jobQueue   chan Job
	quit       chan bool
	wg         *sync.WaitGroup
	maxRetries int
}

// NewWorker creates a new worker
func NewWorker(id int, jobQueue chan Job, wg *sync.WaitGroup) *Worker {
	return &Worker{
		id:         id,
		jobQueue:   jobQueue,
		quit:       make(chan bool),
		wg:         wg,
		maxRetries: 3,
	}
}

// Start starts the worker
func (w *Worker) Start() {
	go func() {
		defer w.wg.Done()

		for {
			select {
			case job := <-w.jobQueue:
				logger.Info("Processing job",
					zap.String("job_id", job.ID()),
					zap.String("job_name", job.Name()),
					zap.Int("worker_id", w.id),
				)

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				err := job.Execute(ctx)
				cancel()

				if err != nil {
					logger.Error("Failed to process job",
						zap.String("job_id", job.ID()),
						zap.String("job_name", job.Name()),
						zap.Int("worker_id", w.id),
						zap.Error(err),
					)
				} else {
					logger.Info("Job completed successfully",
						zap.String("job_id", job.ID()),
						zap.String("job_name", job.Name()),
						zap.Int("worker_id", w.id),
					)
				}

			case <-w.quit:
				logger.Info("Worker is shutting down", zap.Int("worker_id", w.id))
				return
			}
		}
	}()
}

// Stop stops the worker
func (w *Worker) Stop() {
	w.quit <- true
}
