package puan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getImpactingSelections(t *testing.T) {
	theories := []struct {
		name       string
		selections Selections
		expected   Selections
	}{
		{
			name: "subselection than only id remove selection",
			selections: Selections{
				NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
				NewSelectionBuilder("x").WithAction(REMOVE).Build(),
			},
			expected: Selections{},
		},
		{
			name: "subselection two different sub ids",
			selections: Selections{
				NewSelectionBuilder("a").WithSubSelectionID("x").Build(),
				NewSelectionBuilder("a").WithSubSelectionID("y").Build(),
			},
			expected: Selections{
				NewSelectionBuilder("a").WithSubSelectionID("x").Build(),
				NewSelectionBuilder("a").WithSubSelectionID("y").Build(),
			},
		},
		{
			name: "subselection than only id selection",
			selections: Selections{
				NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
				NewSelectionBuilder("x").Build(),
			},
			expected: Selections{
				NewSelectionBuilder("x").Build(),
			},
		},
		{
			name: "only id selection then subselection",
			selections: Selections{
				NewSelectionBuilder("x").Build(),
				NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
			},
			expected: Selections{
				NewSelectionBuilder("x").Build(),
				NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
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
		{
			name: "Add selection, then subselection, then remove sub-selection",
			selections: Selections{
				NewSelectionBuilder("x").Build(),
				NewSelectionBuilder("x").WithSubSelectionID("z").Build(),
				NewSelectionBuilder("x").WithSubSelectionID("z").WithAction(REMOVE).Build(),
			},
			expected: Selections{
				NewSelectionBuilder("x").Build(),
			},
		},
	}

	for _, tt := range theories {
		t.Run(tt.name, func(t *testing.T) {
			actual := getImpactingSelections(tt.selections)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_filterOutRedundantSelections(t *testing.T) {
	theories := []struct {
		name       string
		selections Selections
		expected   Selections
	}{
		{
			name: "remove duplicated selection",
			selections: Selections{
				NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
				NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
			},
			expected: Selections{
				NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
			},
		},
		{
			name: "remove multiple duplicated selections",
			selections: Selections{
				NewSelectionBuilder("x").Build(),
				NewSelectionBuilder("x").Build(),
				NewSelectionBuilder("x").Build(),
				NewSelectionBuilder("x").Build(),
			},
			expected: Selections{
				NewSelectionBuilder("x").Build(),
			},
		},
		{
			name: "should not remove selections",
			selections: Selections{
				NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
				NewSelectionBuilder("x").WithSubSelectionID("z").WithAction(REMOVE).Build(),
			},
			expected: Selections{
				NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
				NewSelectionBuilder("x").WithSubSelectionID("z").WithAction(REMOVE).Build(),
			},
		},
		{
			name: "remove duplicated independent of action",
			selections: Selections{
				NewSelectionBuilder("x").WithAction(REMOVE).Build(),
				NewSelectionBuilder("x").Build(),
			},
			expected: Selections{
				NewSelectionBuilder("x").WithAction(REMOVE).Build(),
			},
		},
	}

	for _, tt := range theories {
		t.Run(tt.name, func(t *testing.T) {
			actual := filterOutRedundantSelections(tt.selections)
			assert.Equal(t, tt.expected, actual, tt.name)
		})
	}
}

func Test_makesRedundant(t *testing.T) {
	theories := []struct {
		name        string
		prioritised Selection
		other       Selection
		expected    bool
	}{
		{
			name:        "not redundant different ids",
			prioritised: NewSelectionBuilder("x").Build(),
			other:       NewSelectionBuilder("y").Build(),
			expected:    false,
		},
		{
			name:        "not redundant different sub ids",
			prioritised: NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
			other:       NewSelectionBuilder("x").WithSubSelectionID("z").Build(),
			expected:    false,
		},
		{
			name:        "redundant same sub ids",
			prioritised: NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
			other:       NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
			expected:    true,
		},
		{
			name:        "re-select same variable, but without sub ids",
			prioritised: NewSelectionBuilder("x").Build(),
			other:       NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
			expected:    true,
		},
		{
			name:        "redundant selection has sub ids",
			prioritised: NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
			other:       NewSelectionBuilder("x").Build(),
			expected:    false,
		},
		{
			name:        "Remove prior sub-selection",
			prioritised: NewSelectionBuilder("y").WithAction(REMOVE).Build(),
			other:       NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
			expected:    true,
		},
		{
			name:        "Remove prior composite selection",
			prioritised: NewSelectionBuilder("x").WithSubSelectionID("y").WithAction(REMOVE).Build(),
			other:       NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
			expected:    true,
		},
		{
			name:        "Remove composite selection, with same primary id",
			prioritised: NewSelectionBuilder("x").WithSubSelectionID("z").WithAction(REMOVE).Build(),
			other:       NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
			expected:    false,
		},
		{
			name:        "Remove composite selection, with same sub-selection id",
			prioritised: NewSelectionBuilder("a").WithSubSelectionID("y").WithAction(REMOVE).Build(),
			other:       NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
			expected:    true,
		},
	}

	for _, tt := range theories {
		actual := tt.prioritised.makesRedundant(tt.other)
		assert.Equal(t, tt.expected, actual, tt.name)
	}
}

func Test_filterOutRemoveSelections(t *testing.T) {
	theories := []struct {
		name       string
		selections Selections
		expected   Selections
	}{
		{
			name: "remove all selections",
			selections: Selections{
				NewSelectionBuilder("x").WithAction(REMOVE).Build(),
				NewSelectionBuilder("y").WithAction(REMOVE).Build(),
			},
			expected: nil,
		},
		{
			name: "remove only one selections",
			selections: Selections{
				NewSelectionBuilder("x").WithAction(REMOVE).Build(),
				NewSelectionBuilder("y").Build(),
			},
			expected: Selections{
				NewSelectionBuilder("y").Build(),
			},
		},
		{
			name: "no remove selections",
			selections: Selections{
				NewSelectionBuilder("x").Build(),
				NewSelectionBuilder("y").Build(),
			},
			expected: Selections{
				NewSelectionBuilder("x").Build(),
				NewSelectionBuilder("y").Build(),
			},
		},
	}

	for _, tt := range theories {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.selections.filterOutRemoveSelections()
			assert.Equal(t, tt.expected, actual)
		})
	}
}
