package puanerror

import "github.com/go-errors/errors"

var (
	InvalidArgument  = errors.New("invalid argument")
	InvalidOperation = errors.New("invalid operation")
	SolverFailed     = errors.New("solver failed")
	NotFound         = errors.New("not found")
)
