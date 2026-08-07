package puan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SolutionCreator_groupSolutionsBySelection(t *testing.T) {
	creator := &SolutionCreator{}

	selection1 := NewSelectionBuilder("x").Build()
	selection2 := NewSelectionBuilder("y").WithAction(REMOVE).Build()
	solution1 := Solution{"x": 1, "y": 0}
	solution2 := Solution{"x": 0, "y": 1}

	got, err := creator.groupSolutionsBySelection(
		[]Solution{solution1, solution2},
		Selections{selection1, selection2},
	)

	assert.NoError(t, err)
	assert.Equal(t, []SolutionBySelection{
		{selection: selection1, solution: solution1},
		{selection: selection2, solution: solution2},
	}, got)
}

func Test_SolutionCreator_groupSolutionsBySelection_givenLengthMismatch_shouldReturnError(
	t *testing.T,
) {
	creator := &SolutionCreator{}

	got, err := creator.groupSolutionsBySelection(
		[]Solution{{"x": 1}},
		Selections{
			NewSelectionBuilder("x").Build(),
			NewSelectionBuilder("y").Build(),
		},
	)

	assert.Nil(t, got)
	assert.Error(t, err)
}
