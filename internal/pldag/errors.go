package pldag

import "github.com/go-errors/errors"

var (
	ErrInvalidConstraintArgument = errors.New("invalid constraint argument")
	ErrDuplicatedVariable        = errors.New("duplicated variable")
	ErrEmptyVariable             = errors.New("empty variable")
	ErrVariableAlreadyExists     = errors.New("variable already exists")
	ErrVariableNotExists         = errors.New("variable not exists")
	ErrInvalidOperands           = errors.New("invalid operands for operator")
)
