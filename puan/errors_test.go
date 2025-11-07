package puan

import (
	"testing"

	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
)

func Test_toPuanError(t *testing.T) {
	type testCase struct {
		name    string
		err     error
		wantErr error
	}

	cases := []testCase{
		{
			name:    "pldag invalid constraint argument",
			err:     pldag.ErrInvalidConstraintArgument,
			wantErr: ErrInvalidArgument,
		},
		{
			name:    "pldag duplicated variable",
			err:     pldag.ErrDuplicatedVariable,
			wantErr: ErrInvalidArgument,
		},
		{
			name:    "pldag empty variable",
			err:     pldag.ErrEmptyVariable,
			wantErr: ErrInvalidArgument,
		},
		{
			name:    "pldag invalid operands",
			err:     pldag.ErrInvalidOperands,
			wantErr: ErrInvalidArgument,
		},
		{
			name:    "pldag variable not exists",
			err:     pldag.ErrVariableNotExists,
			wantErr: ErrInvalidArgument,
		},
		{
			name:    "pldag variable already exists",
			err:     pldag.ErrVariableAlreadyExists,
			wantErr: ErrInvalidOperation,
		},
		{
			name:    "weights invalid action",
			err:     weights.ErrInvalidAction,
			wantErr: ErrInvalidArgument,
		},
		{
			name:    "unknown error",
			err:     errors.New(""),
			wantErr: ErrUnknown,
		},
		{
			name:    "nil error",
			err:     nil,
			wantErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := toPuanError(tc.err)
			assert.ErrorIs(t, gotErr, tc.wantErr)
		})
	}
}
