package glpk

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/puan"
)

func Test_solution_validate_givenSingleSolution_shouldBeValid(
	t *testing.T,
) {
	solution := Solution{
		Solution: fake.New[map[string]int](),
		Status:   "Optimal",
	}

	err := solution.validate()

	assert.NoError(t, err)
}

func Test_solutionResponse_getSingleSolution_givenMultipleSolutions_shouldReturnError(
	t *testing.T,
) {
	solution := SolutionResponse{
		Solutions: []Solution{
			{
				Solution: fake.New[map[string]int](),
				Status:   "Optimal",
			},
			{
				Solution: fake.New[map[string]int](),
				Status:   "Feasible",
			},
		},
	}

	_, err := solution.getSingleSolution()

	assert.Error(t, err)
}

func Test_solution_validate_givenUnexpectedStatus_shouldReturnError(
	t *testing.T,
) {
	solution := Solution{
		Solution: fake.New[map[string]int](),
		Status:   uuid.New().String(),
	}

	err := solution.validate()

	assert.Error(t, err)
}

func Test_solution_validate_givenError_shouldReturnThatError(
	t *testing.T,
) {
	solution := Solution{
		Solution: fake.New[map[string]int](),
		Status:   "Feasible",
		Error:    fake.New[*string](),
	}

	err := solution.validate()

	assert.Error(t, err)
}

func Test_solutionResponse_getManySolutions_shouldReturnSolutionsInOrder(
	t *testing.T,
) {
	response := SolutionResponse{
		Solutions: []Solution{
			{
				Solution: map[string]int{"x": 1},
				Status:   "Optimal",
			},
			{
				Solution: map[string]int{"y": 1},
				Status:   "Optimal",
			},
		},
	}

	solutions, err := response.getManySolutions(2)

	assert.NoError(t, err)
	assert.Equal(t, []puan.Solution{{"x": 1}, {"y": 1}}, solutions)
}

func Test_solutionResponse_getManySolutions_givenMismatchedCount_shouldReturnError(
	t *testing.T,
) {
	response := SolutionResponse{
		Solutions: []Solution{
			{
				Solution: fake.New[map[string]int](),
				Status:   "Optimal",
			},
		},
	}

	_, err := response.getManySolutions(2)

	assert.Error(t, err)
}

func Test_toSparseMatrix(t *testing.T) {
	shape := pldag.NewShape(
		fake.New[int](),
		fake.New[int](),
	)

	entity := pldag.NewSparseMatrix(
		fake.New[[]int](),
		fake.New[[]int](),
		fake.New[[]int](),
		shape,
	)

	sparseMatrix := toSparseMatrix(entity)

	assert.Equal(t, entity.Rows(), sparseMatrix.Rows)
	assert.Equal(t, entity.Columns(), sparseMatrix.Cols)
	assert.Equal(t, entity.Values(), sparseMatrix.Vals)
	assert.Equal(t, entity.Shape().NrOfRows(), sparseMatrix.Shape.Nrows)
	assert.Equal(t, entity.Shape().NrOfColumns(), sparseMatrix.Shape.Ncols)
}

func Test_toBooleanVariables(t *testing.T) {
	variableIDs := fake.New[[]string]()
	variables := toBooleanVariables(variableIDs)

	assert.Equal(t, len(variableIDs), len(variables))
	for i, v := range variables {
		assert.Equal(t, variableIDs[i], v.ID)
		assert.Equal(t, [2]int{0, 1}, v.Bound)
	}
}
