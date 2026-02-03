package glpk

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/puan"
	"github.com/ourstudio-se/puan-sdk-go/puanerror"
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

func (response SolutionResponse) validate() error {
	if len(response.Solutions) != 1 {
		return errors.Errorf(
			"got %d solutions, expected 1",
			len(response.Solutions),
		)
	}
	solution := response.Solutions[0]

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

	if response.Solutions[0].Error != nil {
		return errors.Errorf("got error: %s", *solution.Error)
	}

	return nil
}

func (solution SolutionResponse) asEntity() puan.Solution {
	entity := puan.Solution(solution.Solutions[0].Solution)
	return entity
}

func toSparseMatrix(entity pldag.SparseMatrix) SparseMatrix {
	return SparseMatrix{
		Rows: entity.Rows(),
		Cols: entity.Columns(),
		Vals: entity.Values(),
		Shape: Shape{
			Nrows: entity.Shape().NrOfRows(),
			Ncols: entity.Shape().NrOfColumns(),
		},
	}
}

func toBooleanVariables(variableIDs []string) []Variable {
	var variables []Variable
	for _, v := range variableIDs {
		variables = append(variables, Variable{
			ID:    v,
			Bound: [2]int{0, 1},
		})
	}

	return variables
}

func (s SolveRequest) asBufferedBytes() (*bytes.Buffer, error) {
	bodyBytes, err := json.Marshal(s)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	return bytes.NewBuffer(bodyBytes), nil
}
