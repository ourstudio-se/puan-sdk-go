package puan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_solutionsBySelectionCreator_calculateIndependentSolutionsFromDefault(
	t *testing.T,
) {
	creator := &solutionsBySelectionCreator{}
	defaultSolution := Solution{
		"a": 0,
		"b": 1,
		"c": 0,
	}
	addSelection := NewSelectionBuilder("a").Build()
	removeSelection := NewSelectionBuilder("b").WithAction(REMOVE).Build()
	selections := Selections{addSelection, removeSelection}

	got := creator.calculateIndependentSolutionsFromDefault(
		defaultSolution,
		selections,
	)

	assert.Equal(
		t,
		[]SolutionBySelection{
			{
				selection: addSelection,
				solution:  Solution{"a": 1, "b": 1, "c": 0},
			},
			{
				selection: removeSelection,
				solution:  Solution{"a": 0, "b": 0, "c": 0},
			},
		},
		got,
	)
}
