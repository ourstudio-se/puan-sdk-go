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
		*model.primities,
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
			variables: []string{"x", "y"},
			operation: OperationAnd,
			bias:      -2,
		},
		composite,
	)
}
