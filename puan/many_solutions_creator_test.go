package puan

import (
	"testing"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
	"github.com/stretchr/testify/assert"
)

func Test_manySolutionsCreator_mergeSolutions(t *testing.T) {
	creator := &manySolutionsCreator{}

	dependentSolutions := []Solution{
		{"a": 1, "b": 0},
		{"c": 1},
	}
	independentSolutions := []Solution{
		{"b": 1, "d": 0},
		{"e": 1},
	}

	got, err := creator.mergeSolutions(dependentSolutions, independentSolutions)

	assert.NoError(t, err)
	assert.Equal(
		t,
		[]Solution{
			{"a": 1, "b": 1, "d": 0},
			{"c": 1, "e": 1},
		},
		got,
	)
}

func Test_manySolutionsCreator_mergeSolutions_givenLengthMismatch_shouldReturnError(
	t *testing.T,
) {
	creator := &manySolutionsCreator{}

	dependentSolutions := []Solution{
		fake.New[Solution](),
	}
	independentSolutions := []Solution{
		fake.New[Solution](),
		fake.New[Solution](),
	}

	_, err := creator.mergeSolutions(dependentSolutions, independentSolutions)

	assert.Error(t, err)
}
