package puan

import (
	"testing"

	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
	"github.com/stretchr/testify/assert"
)

func Test_NewMultiWeightSolverQuery(t *testing.T) {
	polyhedron := pldag.NewPolyhedron([][]int{{1}}, []int{1})
	variables := []string{"x", "y"}
	weightGroups := []weights.Weights{
		{"x": 1, "y": 2},
		{"x": 3, "y": 4},
	}

	got := NewMultiWeightSolverQuery(polyhedron, variables, weightGroups)

	assert.Equal(t, polyhedron, got.Polyhedron())
	assert.Equal(t, variables, got.Variables())
	assert.Equal(t, weightGroups, got.WeightGroups())
}
