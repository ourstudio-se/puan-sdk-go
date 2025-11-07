package puan

import (
	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
)

var (
	ErrInvalidOperation = errors.New("invalid operation")
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrAmbiguous        = errors.New("ambiguous")
	ErrNotFound         = errors.New("not found")
	ErrUnknown          = errors.New("unknown error")
)

// nolint:gocyclo
func toPuanError(err error) error {
	switch {
	case errors.Is(err, pldag.ErrInvalidConstraintArgument):
		return errors.Wrap(errors.Errorf("%w: %w", ErrInvalidArgument, err), 1)
	case errors.Is(err, pldag.ErrDuplicatedVariable):
		return errors.Wrap(errors.Errorf("%w: %w", ErrInvalidArgument, err), 1)
	case errors.Is(err, pldag.ErrEmptyVariable):
		return errors.Wrap(errors.Errorf("%w: %w", ErrInvalidArgument, err), 1)
	case errors.Is(err, pldag.ErrInvalidOperands):
		return errors.Wrap(errors.Errorf("%w: %w", ErrInvalidArgument, err), 1)
	case errors.Is(err, pldag.ErrVariableNotExists):
		return errors.Wrap(errors.Errorf("%w: %w", ErrInvalidArgument, err), 1)

	case errors.Is(err, pldag.ErrVariableAlreadyExists):
		return errors.Wrap(errors.Errorf("%w: %w", ErrInvalidOperation, err), 1)

	case errors.Is(err, weights.ErrInvalidAction):
		return errors.Wrap(errors.Errorf("%w: %w", ErrInvalidArgument, err), 1)

	case err != nil:
		return errors.Wrap(errors.Errorf("%w: %w", ErrUnknown, err), 1)
	}

	return nil
}
