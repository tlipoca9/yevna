package yevna

import "github.com/cockroachdb/errors"

var (
	// ErrNameRequired is returned when the name is empty
	ErrNameRequired = errors.New("name is required")
	// ErrStdoutAlreadySet is returned when the stdout is not empty
	ErrStdoutAlreadySet = errors.New("stdout is already set")
)
