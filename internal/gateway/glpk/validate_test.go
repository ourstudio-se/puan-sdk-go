package glpk

import (
	"testing"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
	"github.com/ourstudio-se/puan-sdk-go/puanerror"
	"github.com/stretchr/testify/assert"
)

func Test_solution_validate_givenValidStatusAndNoError_shouldReturnNil(t *testing.T) {
	solution := Solution{Status: "Optimal"}

	err := solution.validate()

	assert.NoError(t, err)
}

func Test_solution_validate_givenInvalidStatus_shouldReturnError(t *testing.T) {
	solution := Solution{Status: "Unknown"}

	err := solution.validate()

	assert.Error(t, err)
}

func Test_solution_validate_givenMipFailedStatus_shouldReturnSolverFailedError(t *testing.T) {
	solution := Solution{
		Status: "MIPFAILED",
	}

	err := solution.validate()

	assert.Error(t, err)
	assert.ErrorIs(t, err, puanerror.SolverFailed)
}

func Test_solution_validate_givenValidStatusAndError_shouldReturnError(t *testing.T) {
	msg := fake.New[string]()
	solution := Solution{
		Status: "feasible",
		Error:  &msg,
	}

	err := solution.validate()

	assert.Error(t, err)
}
