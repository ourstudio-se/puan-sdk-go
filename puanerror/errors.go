package puanerror

import "github.com/go-errors/errors"

var (
	Ambiguous        = errors.New("ambiguous")
	InvalidArgument  = errors.New("invalid argument")
	InvalidOperation = errors.New("invalid operation")
	NotFound         = errors.New("not found")
)
