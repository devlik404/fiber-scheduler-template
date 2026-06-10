package scheduler

import "context"

type Job struct {
	Name     string
	Schedule string
	Handler  Handler
}

type Handler interface {
	Run(ctx context.Context) error
}
