package task

import (
	"log/slog"

	"template_sch/internal/config"
	"template_sch/internal/scheduler"
	"template_sch/internal/task/jobs"
)

type Registry struct {
	cfg    config.TaskConfig
	deps   Dependencies
	logger *slog.Logger
}

func NewRegistry(cfg config.TaskConfig, deps Dependencies, logger *slog.Logger) Registry {
	return Registry{
		cfg:    cfg,
		deps:   deps,
		logger: logger,
	}
}

func (r Registry) Jobs() []scheduler.Job {
	return []scheduler.Job{
		{
			Name:     "example-task",
			Schedule: r.cfg.ExampleSchedule,
			Handler:  jobs.NewExampleJob(r.deps.DB, r.logger.With("component", "job", "job", "example-task")),
		},
		// generator:task-registry-job
	}
}
