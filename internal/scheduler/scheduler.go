package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"template_sch/internal/config"
	"template_sch/timezone"
)

type Scheduler struct {
	cron   *cron.Cron
	logger *slog.Logger
}

func New(cfg config.SchedulerConfig, jobs []Job, logger *slog.Logger) (*Scheduler, error) {
	logger = logger.With("component", "scheduler")

	location, err := timezone.Load(cfg.Location)
	if err != nil {
		return nil, fmt.Errorf("load scheduler location: %w", err)
	}

	runner := cron.New(
		cron.WithLocation(location),
		cron.WithChain(cron.Recover(cronLogger{logger: logger})),
	)

	for _, job := range jobs {
		job := job
		if err := validateJob(job); err != nil {
			return nil, err
		}

		_, err := runner.AddFunc(job.Schedule, func() {
			ctx := context.Background()
			startedAt := time.Now()
			jobLogger := logger.With("job", job.Name)

			jobLogger.Info("job started")
			if err := job.Handler.Run(ctx); err != nil {
				jobLogger.Error("job failed", "duration", time.Since(startedAt).String(), "error", err)
				return
			}

			jobLogger.Info("job completed", "duration", time.Since(startedAt).String())
		})
		if err != nil {
			return nil, fmt.Errorf("register job %q: %w", job.Name, err)
		}
	}

	return &Scheduler{
		cron:   runner,
		logger: logger,
	}, nil
}

func (s *Scheduler) Start() {
	s.cron.Start()
	s.logger.Info("scheduler started")
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Info("scheduler stopped")
}

func validateJob(job Job) error {
	if job.Name == "" {
		return fmt.Errorf("job name is required")
	}

	if job.Schedule == "" {
		return fmt.Errorf("schedule is required for job %q", job.Name)
	}

	if job.Handler == nil {
		return fmt.Errorf("handler is required for job %q", job.Name)
	}

	return nil
}

type cronLogger struct {
	logger *slog.Logger
}

func (l cronLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info(msg, keysAndValues...)
}

func (l cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	args := append(keysAndValues, "error", err)
	l.logger.Error(msg, args...)
}
