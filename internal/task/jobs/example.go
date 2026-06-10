package jobs

import (
	"context"
	"log/slog"

	"template_sch/internal/platform/database"
)

type ExampleJob struct {
	db     *database.Connections
	logger *slog.Logger
}

func NewExampleJob(db *database.Connections, logger *slog.Logger) ExampleJob {
	return ExampleJob{
		db:     db,
		logger: logger,
	}
}

func (j ExampleJob) Run(ctx context.Context) error {
	// Ganti isi method ini dengan logic utama scheduler Anda.
	if j.db == nil || j.db.Primary == nil {
		j.logger.Debug("primary database is disabled")
	}

	j.logger.Info("example task logic executed")
	return nil
}
