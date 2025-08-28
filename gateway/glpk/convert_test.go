package glpk

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/domain/pldag"
	"github.com/ourstudio-se/puan-sdk-go/domain/puan"
	"github.com/ourstudio-se/puan-sdk-go/fake"
)

func Test_solutionResponse_asEntity_givenSingleSolution_shouldReturnThatSolution(
	t *testing.T,
) {
	solution := SolutionResponse{
		Solutions: []Solution{
			{
				Solution: fake.New[map[string]int](),
			},
		},
	}

	entity := solution.asEntity()

	expected := puan.Solution(solution.Solutions[0].Solution)
	assert.Equal(t, expected, entity)
}

func Test_solutionResponse_asEntity_givenMultipleSolutions_shouldReturnFirstSolution(
	t *testing.T,
) {
	solution := SolutionResponse{
		Solutions: []Solution{
			{
				Solution: fake.New[map[string]int](),
			},
			{
				Solution: fake.New[map[string]int](),
			},
		},
	}

	entity := solution.asEntity()

	expected := puan.Solution(solution.Solutions[0].Solution)
	assert.Equal(t, expected, entity)
}

func Test_solutionResponse_validate_givenSingleSolution_shouldBeValid(
	t *testing.T,
) {
	solution := SolutionResponse{
		Solutions: []Solution{
			{
				Solution: fake.New[map[string]int](),
				Status:   "Optimal",
			},
		},
	}

	err := solution.validate()

	assert.NoError(t, err)
}

func Test_solutionResponse_validate_givenMultipleSolutions_shouldReturnError(
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
				Status:   "feasible",
			},
		},
	}

	err := solution.validate()

	assert.Error(t, err)
}

func Test_solutionResponse_validate_givenUnexpedStatus_shouldReturnError(
	t *testing.T,
) {
	solution := SolutionResponse{
		Solutions: []Solution{
			{
				Solution: fake.New[map[string]int](),
				Status:   faker.Word(),
			},
		},
	}

	err := solution.validate()

	assert.Error(t, err)
}

func Test_solutionResponse_validate_givenError_shouldReturnThatError(
	t *testing.T,
) {
	solution := SolutionResponse{
		Solutions: []Solution{
			{
				Solution: fake.New[map[string]int](),
				Status:   "Feasible",
				Error:    fake.New[*string](),
			},
		},
	}

	err := solution.validate()

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
