package pldag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Model_SetPrimities(t *testing.T) {
	model := New()

	model.SetPrimities([]string{"x", "y", "z"}...)

	assert.Equal(
		t,
		*model.variables,
		map[string]any{
			"x": nil,
			"y": nil,
			"z": nil,
		},
	)
}

func Test_Model_SetAnd(t *testing.T) {
	model := New()
	model.SetPrimities([]string{"x", "y"}...)

	id := model.SetAnd([]string{"x", "y"}...)

	composite := (*model.composites)[id]
	assert.Equal(
		t,
		Operation{
			variables: map[string]any{
				"x": nil,
				"y": nil,
			},
			operation: OperationAnd,
			bias:      -2,
		},
		composite,
	)
}

func Test_Model_NewLinearSystem_givenAnd(t *testing.T) {
	model := New()
	model.SetPrimities([]string{"x", "y"}...)
	model.SetAnd([]string{"x", "y"}...)

	linearSystem := model.NewLinearSystem()

	assert.Equal(
		t,
		linearSystem,
		LinearSystem{
			matrix:          [][]int{{1, 1, -2}, {-1, -1, 1}},
			rightHandVector: []int{0, -1},
		},
	)
}
