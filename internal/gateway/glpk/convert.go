package glpk

import (
	"bytes"
	"encoding/json"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/puan"
)

var VALID_STATUSES = map[string]any{
	"optimal":  nil,
	"feasible": nil,
}

func (response SolutionResponse) getSingleSolution() (puan.Solution, error) {
	if len(response.Solutions) != 1 {
		return puan.Solution{}, errors.Errorf(
			"got %d solutions, expected 1",
			len(response.Solutions),
		)
	}

	return response.Solutions[0].asEntity()
}

func (solution Solution) asEntity() (puan.Solution, error) {
	if err := solution.validate(); err != nil {
		return puan.Solution{}, err
	}

	return puan.Solution(solution.Solution), nil
}

func (response SolutionResponse) getManySolutions(
	wantCount int,
) ([]puan.Solution, error) {
	if len(response.Solutions) != wantCount {
		return nil, errors.Errorf(
			"got %d solutions, want %d",
			len(response.Solutions),
			wantCount,
		)
	}

	entities := make([]puan.Solution, wantCount)
	for i, solution := range response.Solutions {
		entity, err := solution.asEntity()
		if err != nil {
			return nil, err
		}

		entities[i] = entity
	}

	return entities, nil
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
