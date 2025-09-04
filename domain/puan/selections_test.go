package puan

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/utils"
)

func Test_removeRedundantSelections_givenRedundantSelections_shouldReturnUniqueSelections(t *testing.T) {
	y := "y"
	selections := Selections{
		{
			id:             "x",
			subSelectionID: &y,
			action:         ADD,
		},
		{
			id:             "x",
			subSelectionID: &y,
			action:         ADD,
		},
	}

	actual := selections.removeRedundantSelections()
	expected := Selections{
		{
			id:             "x",
			subSelectionID: &y,
			action:         ADD,
		},
	}

	assert.Equal(t, expected, actual)
}

func Test_removeRedundantSelections_givenItemsWithOthersAndSingle_shouldReturnBothSelections(t *testing.T) {
	y := "y"
	selections := Selections{
		{
			id:             "x",
			subSelectionID: &y,
			action:         ADD,
		},
		{
			id:     "x",
			action: ADD,
		},
	}

	actual := selections.removeRedundantSelections()
	expected := Selections{
		{
			id:             "x",
			subSelectionID: &y,
			action:         ADD,
		},
		{
			id:     "x",
			action: ADD,
		},
	}

	assert.Equal(t, expected, actual)
}

func Test_removeRedundantSelections_givenSelectionsWithSameIDsInDifferentOrder_shouldReturnOneSelections(t *testing.T) {
	x := "x"
	y := "y"
	selections := Selections{
		{
			id:             "x",
			subSelectionID: &y,
			action:         ADD,
		},
		{
			id:             "y",
			subSelectionID: &x,
			action:         ADD,
		},
	}

	actual := selections.removeRedundantSelections()
	expected := Selections{
		{
			id:             "y",
			subSelectionID: &x,
			action:         ADD,
		},
	}

	assert.Equal(t, expected, actual)
}

func Test_removeRedundantSelections_givenRemoveSelection_shouldReturnEmptySelection(t *testing.T) {
	selections := Selections{
		{
			id:     "x",
			action: REMOVE,
		},
	}

	actual := selections.removeRedundantSelections()
	expected := Selections{}

	assert.Equal(t, expected, actual)
}

func Test_removeRedundantSelections_givenEmptySelection_shouldReturnEmptySelection(t *testing.T) {
	selections := Selections{}

	actual := selections.removeRedundantSelections()
	expected := Selections{}

	assert.Equal(t, expected, actual)
}

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
