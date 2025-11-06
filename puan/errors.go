package puan

import (
	"github.com/go-errors/errors"
)

var (
	ErrInvalidOperation = errors.New("invalid operation")
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrAmbiguous        = errors.New("ambiguous")
	ErrNotFound         = errors.New("not found")
	ErrUnknown          = errors.New("unknown error")
)
