package puanerror

import "github.com/go-errors/errors"

var (
	ErrInvalidOperation = errors.New("invalid operation")
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrAmbiguous        = errors.New("ambiguous")
	ErrNotFound         = errors.New("not found")

	ErrInvalidConstraintArgument = errors.New("invalid constraint argument")
	ErrDuplicatedVariable        = errors.New("duplicated variable")
	ErrEmptyVariable             = errors.New("empty variable")
	ErrVariableAlreadyExists     = errors.New("variable already exists")
	ErrVariableNotExists         = errors.New("variable not exists")
	ErrInvalidOperands           = errors.New("invalid operands for operator")

	ErrInvalidAction = errors.New("invalid action")
)
