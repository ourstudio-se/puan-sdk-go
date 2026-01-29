package puanerror

import "github.com/go-errors/errors"

var (
	InvalidArgument  = errors.New("invalid argument")
	InvalidOperation = errors.New("invalid operation")
	NoSolutionFound  = errors.New("no solution found")
	NotFound         = errors.New("not found")
)
