package yevna

import (
	"context"
)

// Command returns a new Cmd with the global options
func Command(ctx context.Context, name string, args ...string) *Cmd {
	return Default().Command(ctx, name, args...)
}

// Pipe returns a new Cmd with the pipeline
func Pipe(ctx context.Context, commands ...[]string) *Cmd {
	return Default().Pipe(ctx, commands...)
}

func init() {
	SetDefault(NewContext())
}
