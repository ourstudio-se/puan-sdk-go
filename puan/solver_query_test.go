package puan

import (
	"testing"

	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
	"github.com/stretchr/testify/assert"
)

func Test_NewMultiWeightSolverQuery_givenValidWeights_shouldReturnQuery(t *testing.T) {
	polyhedron := pldag.NewPolyhedron([][]int{{1}}, []int{1})
	variables := []string{"x", "y"}
	weightGroups := []weights.Weights{
		{"x": 1, "y": 2},
		{"x": 3, "y": 4},
	}

	got, err := NewMultiWeightSolverQuery(polyhedron, variables, weightGroups)

	assert.NoError(t, err)
	assert.Equal(t, polyhedron, got.Polyhedron())
	assert.Equal(t, variables, got.Variables())
	assert.Equal(t, weightGroups, got.WeightGroups())
}

func Test_NewMultiWeightSolverQuery_givenWeightsTooLarge_shouldReturnError(t *testing.T) {
	polyhedron := pldag.NewPolyhedron([][]int{{1}}, []int{1})
	variables := []string{"x"}
	weightGroups := []weights.Weights{
		{"x": 1},
		{"x": weights.WEIGHTS_SATURATION_LIMIT + 1},
	}

	got, err := NewMultiWeightSolverQuery(polyhedron, variables, weightGroups)

	assert.Error(t, err)
	assert.Nil(t, got)
}
