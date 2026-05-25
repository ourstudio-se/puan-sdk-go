//nolint:lll
package puan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Selections_getImpacting(t *testing.T) {
	theories := []struct {
		name       string
		selections Selections
		expected   Selections
	}{
		{
			name: "Add primitive with sub-selection, then remove primary primitive",
			selections: Selections{
				NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
				NewSelectionBuilder("x").WithAction(REMOVE).Build(),
			},
			expected: Selections{
				NewSelectionBuilder("x").WithAction(REMOVE).Build(),
			},
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
			name: "Add primitive with sub-selection, then add primary primitive only",
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
				NewSelectionBuilder("y").WithSubSelectionID("x").Build(),
			},
		},
		{
			name: "Single remove",
			selections: Selections{
				NewSelectionBuilder("x").WithAction(REMOVE).Build(),
			},
			expected: Selections{
				NewSelectionBuilder("x").WithAction(REMOVE).Build(),
			},
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
			expected: Selections{NewSelectionBuilder("x").WithSubSelectionID("y").WithAction(REMOVE).Build()},
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
				NewSelectionBuilder("x").WithSubSelectionID("z").WithAction(REMOVE).Build(),
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
				NewSelectionBuilder("x").WithSubSelectionID("z").WithAction(REMOVE).Build(),
			},
		},
	}

	for _, tt := range theories {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.selections.getImpacting()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_Selections_filterOutRedundant(t *testing.T) {
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
			actual := tt.selections.filterOutRedundant()
			assert.Equal(t, tt.expected, actual)
		})
	}

	for _, tt := range theories {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.selections.filterOutRedundant()
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
		{
			name:        "Remove two different primitives",
			prioritised: NewSelectionBuilder("x").WithAction(REMOVE).Build(),
			other:       NewSelectionBuilder("y").WithAction(REMOVE).Build(),
			expected:    false,
		},
		{
			name: "Add composite selection, with a sub-set of sub ids",
			prioritised: NewSelectionBuilder("a").
				WithSubSelectionID("y").
				Build(),
			other: NewSelectionBuilder("a").
				WithSubSelectionID("x").
				WithSubSelectionID("y").
				Build(),
			expected: true,
		},
		{
			name:        "Remove with sub-selection, then add with reversed order",
			prioritised: NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
			other:       NewSelectionBuilder("y").WithSubSelectionID("x").WithAction(REMOVE).Build(),
			expected:    true,
		},
	}

	for _, tt := range theories {
		actual := tt.prioritised.makesRedundant(tt.other)
		assert.Equal(t, tt.expected, actual, tt.name)
	}
}

func Test_Selection_IDs(t *testing.T) {
	selection := NewSelectionBuilder("x").
		WithSubSelectionID("y").
		WithSubSelectionID("z").
		Build()
	assert.Equal(t, []string{"x", "y", "z"}, selection.IDs())
}

func Test_Selections_extendWithPrimaryPrimitiveSelections(t *testing.T) {
	selections := Selections{
		NewSelectionBuilder("x").WithSubSelectionID("y").WithAction(ADD).Build(),
		NewSelectionBuilder("z").Build(),
		NewSelectionBuilder("z").WithSubSelectionID("w").WithAction(REMOVE).Build(),
	}

	extended := selections.modifyForQuery()

	want := Selections{
		NewSelectionBuilder("x").WithAction(ADD).Build(),
		NewSelectionBuilder("x").WithSubSelectionID("y").WithAction(ADD).Build(),
		NewSelectionBuilder("z").Build(),
		NewSelectionBuilder("z").WithAction(REMOVE).Build(),
	}

	assert.Equal(t, want, extended)
}

func Test_Selections_ids(t *testing.T) {
	selections := Selections{
		NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
		NewSelectionBuilder("z").Build(),
		NewSelectionBuilder("z").WithSubSelectionID("w").Build(),
	}

	ids := selections.ids()
	assert.Equal(t, []string{"x", "y", "z", "w"}, ids)
}

func Test_Selections_split(t *testing.T) {
	type theory struct {
		name       string
		selections Selections
		wantFirst  Selections
		wantSecond Selections
	}

	theories := []theory{
		{
			name:       "empty",
			selections: Selections{},
			wantFirst:  nil,
			wantSecond: nil,
		},
		{
			name: "single element — second is nil",
			selections: Selections{
				NewSelectionBuilder("a").Build(),
			},
			wantFirst: Selections{
				NewSelectionBuilder("a").Build(),
			},
			wantSecond: nil,
		},
		{
			name: "two elements — one element each",
			selections: Selections{
				NewSelectionBuilder("a").Build(),
				NewSelectionBuilder("b").Build(),
			},
			wantFirst: Selections{
				NewSelectionBuilder("a").Build(),
			},
			wantSecond: Selections{
				NewSelectionBuilder("b").Build(),
			},
		},
		{
			name: "four elements - split evenly",
			selections: Selections{
				NewSelectionBuilder("a").Build(),
				NewSelectionBuilder("b").Build(),
				NewSelectionBuilder("c").Build(),
				NewSelectionBuilder("d").Build(),
			},
			wantFirst: Selections{
				NewSelectionBuilder("a").Build(),
				NewSelectionBuilder("b").Build(),
			},
			wantSecond: Selections{
				NewSelectionBuilder("c").Build(),
				NewSelectionBuilder("d").Build(),
			},
		},
		{
			name: "five elements — first half is larger when odd",
			selections: Selections{
				NewSelectionBuilder("a").Build(),
				NewSelectionBuilder("b").Build(),
				NewSelectionBuilder("c").Build(),
				NewSelectionBuilder("d").Build(),
				NewSelectionBuilder("e").Build(),
			},
			wantFirst: Selections{
				NewSelectionBuilder("a").Build(),
				NewSelectionBuilder("b").Build(),
				NewSelectionBuilder("c").Build(),
			},
			wantSecond: Selections{
				NewSelectionBuilder("d").Build(),
				NewSelectionBuilder("e").Build(),
			},
		},
	}

	for _, tt := range theories {
		t.Run(tt.name, func(t *testing.T) {
			first, second := tt.selections.split()
			assert.Equal(t, tt.wantFirst, first)
			assert.Equal(t, tt.wantSecond, second)
		})
	}
}

func Test_Selection_Hash(t *testing.T) {
	theories := []struct {
		name      string
		selection Selection
		expected  string
	}{
		{
			name:      "ADD selection",
			selection: NewSelectionBuilder("x").Build(),
			expected:  "c3b14e6c5ba76924b48df086f52c5e3237675ff5",
		},
		{
			name:      "REMOVE selection",
			selection: NewSelectionBuilder("x").WithAction(REMOVE).Build(),
			expected:  "235b76cf9d9c9334c4736b4c1f5439fe92f49327",
		},
		{
			name:      "different id same action",
			selection: NewSelectionBuilder("y").Build(),
			expected:  "390671eaeda30f8b9a6ee3dae4a357f47da8803b",
		},
		{
			name:      "composite with sub-selection",
			selection: NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
			expected:  "c93cf323b09bf5b73b27a92c4208b3c8c1f5bded",
		},
	}

	for _, tt := range theories {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.selection.Hash())
		})
	}
}

func Test_Selection_Equals(t *testing.T) {
	theories := []struct {
		name     string
		first    Selection
		second   Selection
		expected bool
	}{
		{
			name:     "same primitive selection",
			first:    NewSelectionBuilder("x").Build(),
			second:   NewSelectionBuilder("x").Build(),
			expected: true,
		},
		{
			name:     "different id",
			first:    NewSelectionBuilder("x").Build(),
			second:   NewSelectionBuilder("y").Build(),
			expected: false,
		},
		{
			name:     "different action",
			first:    NewSelectionBuilder("x").Build(),
			second:   NewSelectionBuilder("x").WithAction(REMOVE).Build(),
			expected: false,
		},
		{
			name:     "composite with same sub-selection",
			first:    NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
			second:   NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
			expected: true,
		},
		{
			name:     "composite without sub-selection",
			first:    NewSelectionBuilder("x").WithSubSelectionID("y").Build(),
			second:   NewSelectionBuilder("x").Build(),
			expected: false,
		},
		{
			name: "sub-selections in different order",
			first: NewSelectionBuilder("x").
				WithSubSelectionID("y").
				WithSubSelectionID("z").
				Build(),
			second: NewSelectionBuilder("x").
				WithSubSelectionID("z").
				WithSubSelectionID("y").
				Build(),
			expected: true,
		},
		{
			name: "different sub-selections",
			first: NewSelectionBuilder("x").
				WithSubSelectionID("y").
				Build(),
			second: NewSelectionBuilder("x").
				WithSubSelectionID("z").
				Build(),
			expected: false,
		},
	}

	for _, tt := range theories {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.first.Equals(tt.second))
		})
	}
}
