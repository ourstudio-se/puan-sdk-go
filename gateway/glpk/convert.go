package glpk

import (
	"strings"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/domain/puan"
)

var VALID_STATUSES = map[string]any{
	"optimal":  nil,
	"feasible": nil,
}

func (solution SolutionResponse) getSolutionEntity() (puan.Solution, error) {
	if err := solution.validate(); err != nil {
		return puan.Solution{}, err
	}

	return solution.asEntity(), nil
}

func (solution SolutionResponse) validate() error {
	if len(solution.Solutions) != 1 {
		return errors.Errorf("got %d solutions, expected 1", len(solution.Solutions))
	}

	status := strings.ToLower(solution.Solutions[0].Status)
	if _, ok := VALID_STATUSES[status]; !ok {
		return errors.Errorf(
			"got invalid status: %s, expected one of %v",
			status,
			VALID_STATUSES,
		)
	}

	if solution.Solutions[0].Error != nil {
		return errors.Errorf("got error: %s", *solution.Solutions[0].Error)
	}

	return nil
}

func (solution SolutionResponse) asEntity() puan.Solution {
	entity := puan.Solution(solution.Solutions[0].Solution)
	return entity
}
