package glpk

import (
	"strings"

	"github.com/go-errors/errors"
	"github.com/ourstudio-se/puan-sdk-go/puanerror"
)

func (solution Solution) validate() error {
	status := strings.ToLower(solution.Status)
	if _, ok := VALID_STATUSES[status]; !ok {
		var msg string
		if solution.Error != nil {
			msg = *solution.Error
		}

		if status == "mipfailed" {
			return errors.Errorf(
				"%w: message: %s",
				puanerror.SolverFailed,
				msg,
			)
		}

		return errors.Errorf(
			"got invalid status: %s, expected one of %v. Message: %s",
			status,
			VALID_STATUSES,
			msg,
		)
	}

	if solution.Error != nil {
		return errors.Errorf("got error: %s", *solution.Error)
	}

	return nil
}
