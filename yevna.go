package yevna

import (
	"context"
)

// Command returns a new Cmd with the global options
func Command(ctx context.Context, name string, args ...string) *Cmd {
	return NewContext(ctx).Command(name, args...)
}

// Pipe returns a new Cmd with the pipeline
func Pipe(ctx context.Context, commands ...[]string) *Cmd {
	return NewContext(ctx).Pipe(commands...)
}
