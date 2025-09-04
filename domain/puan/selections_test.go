package puan

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/utils"
)

func Test_removeRedundantSelections(t *testing.T) {
	theories := []struct {
		name       string
		selections Selections
		expected   Selections
	}{
		{
			name: "subselection than only id remove selection",
			selections: Selections{
				{
					id:             "x",
					subSelectionID: utils.Pointer("y"),
					action:         ADD,
				},
				{
					id:     "x",
					action: REMOVE,
				},
			},
			expected: Selections{},
		},
		{
			name: "subselection two different ids",
			selections: Selections{
				{
					id:             "a",
					subSelectionID: utils.Pointer("x"),
					action:         ADD,
				},
				{
					id:             "a",
					subSelectionID: utils.Pointer("y"),
					action:         ADD,
				},
			},
			expected: Selections{
				{
					id:             "a",
					subSelectionID: utils.Pointer("x"),
					action:         ADD,
				},
				{
					id:             "a",
					subSelectionID: utils.Pointer("y"),
					action:         ADD,
				},
			},
		},
		{
			name: "subselection than only id selection",
			selections: Selections{
				{
					id:             "x",
					subSelectionID: utils.Pointer("y"),
					action:         ADD,
				},
				{
					id:     "x",
					action: ADD,
				},
			},
			expected: Selections{
				{
					id:     "x",
					action: ADD,
				},
			},
		},
		{
			name: "only id selection then subselection",
			selections: Selections{
				{
					id:     "x",
					action: ADD,
				},
				{
					id:             "x",
					subSelectionID: utils.Pointer("y"),
					action:         ADD,
				},
			},
			expected: Selections{
				{
					id:     "x",
					action: ADD,
				},
				{
					id:             "x",
					subSelectionID: utils.Pointer("y"),
					action:         ADD,
				},
			},
		},
		{
			name: "duplicate sub-selection",
			selections: Selections{
				NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
				NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
			},
			expected: Selections{
				NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
			},
		},
		{
			name: "reversed sub-selections",
			selections: Selections{
				NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
				NewSelectionBuilder("y").WithSubSelectionID("x").Build(),
			},
			expected: Selections{
				NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
				NewSelectionBuilder("y").WithSubSelectionID("x").Build(),
			},
		},
		{
			name: "Single remove",
			selections: Selections{
				NewSelectionBuilder("x").WithAction(REMOVE).Build(),
			},
			expected: Selections{},
		},
		{
			name:       "Empty selections",
			selections: Selections{},
			expected:   Selections{},
		},
		{
			name: "Add sub-selection, then remove it",
			selections: Selections{
				NewSelectionBuilder("x").WithSubSelectionID("y").WithAction(ADD).Build(),
				NewSelectionBuilder("x").WithSubSelectionID("y").WithAction(REMOVE).Build(),
			},
			expected: Selections{},
		},
		{
			name: "Add sub-selection, then add another",
			selections: Selections{
				NewSelectionBuilder("x").WithSubSelectionID("y").WithAction(ADD).Build(),
				NewSelectionBuilder("x").WithSubSelectionID("z").WithAction(ADD).Build(),
			},
			expected: Selections{
				NewSelectionBuilder("x").WithSubSelectionID("y").WithAction(ADD).Build(),
				NewSelectionBuilder("x").WithSubSelectionID("z").WithAction(ADD).Build(),
			},
		},
		{
			name: "Add sub-selection, then remove another sub-selection",
			selections: Selections{
				NewSelectionBuilder("x").WithSubSelectionID("y").WithAction(ADD).Build(),
				NewSelectionBuilder("x").WithSubSelectionID("z").WithAction(REMOVE).Build(),
			},
			expected: Selections{
				NewSelectionBuilder("x").WithSubSelectionID("y").WithAction(ADD).Build(),
			},
		},
	}

	for _, tt := range theories {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.selections.removeRedundantSelections()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_createOrderIndependentID_givenUnsortedIDs_shouldReturnSortedStringID(t *testing.T) {
	y := "y"
	actual := createOrderIndependentID("x", &y)
	expected := "x,y"

	assert.Equal(t, expected, actual)
}

func Test_createOrderIndependentID_givenSingleID_shouldReturnSameID(t *testing.T) {
	actual := createOrderIndependentID("z", nil)
	expected := "z"

	assert.Equal(t, expected, actual)
}
