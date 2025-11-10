package puanerror

import "github.com/go-errors/errors"

var (
	Ambiguous             = errors.New("ambiguous")
	DuplicatedVariable    = errors.New("duplicated variable")
	EmptyVariable         = errors.New("empty variable")
	InvalidArgument       = errors.New("invalid argument")
	InvalidOperation      = errors.New("invalid operation")
	NotFound              = errors.New("not found")
	VariableAlreadyExists = errors.New("variable already exists")
	VariableNotExists     = errors.New("variable not exists")
)
