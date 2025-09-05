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
