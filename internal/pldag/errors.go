package pldag

import "github.com/go-errors/errors"

var (
	ErrInvalidConstraintArgument = errors.New("invalid constraint argument")
	ErrDuplicatedVariable        = errors.New("duplicated variable")
	ErrEmptyVariable             = errors.New("empty variable")
	ErrAlreadyExists             = errors.New("variable already exists")
	ErrInvalidOperands           = errors.New("invalid operands for operator")
	ErrVariableNotFound          = errors.New("variable not found")
)
