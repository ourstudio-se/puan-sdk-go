package puan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_singleSolutionCreator_getIDsToAssume(t *testing.T) {
	creator := &singleSolutionCreator{}

	tests := []struct {
		name          string
		selections    Selections
		solution      Solution
		wantAssume    []string
		wantAssumeNot []string
	}{
		{
			name:          "given no selections, returns empty",
			selections:    nil,
			solution:      Solution{"a": 1},
			wantAssume:    nil,
			wantAssumeNot: nil,
		},
		{
			name: "given selected selection, assumes its id",
			selections: Selections{
				NewSelectionBuilder("a").Build(),
			},
			solution:      Solution{"a": 1},
			wantAssume:    []string{"a"},
			wantAssumeNot: nil,
		},
		{
			name: "given unselected selection, assumes not its id",
			selections: Selections{
				NewSelectionBuilder("a").Build(),
			},
			solution:      Solution{"a": 0},
			wantAssume:    nil,
			wantAssumeNot: []string{"a"},
		},
		{
			name: "given selection missing from solution, assumes not its id",
			selections: Selections{
				NewSelectionBuilder("a").Build(),
			},
			solution:      Solution{},
			wantAssume:    nil,
			wantAssumeNot: []string{"a"},
		},
		{
			name: "given selected selection with selected sub-selections, assumes all",
			selections: Selections{
				NewSelectionBuilder("a").
					WithSubSelectionID("b").
					WithSubSelectionID("c").
					Build(),
			},
			solution: Solution{
				"a": 1,
				"b": 1,
				"c": 1,
			},
			wantAssume:    []string{"a", "b", "c"},
			wantAssumeNot: nil,
		},
		{
			name: "given selected selection with unselected sub-selections, assumes only parent",
			selections: Selections{
				NewSelectionBuilder("a").
					WithSubSelectionID("b").
					WithSubSelectionID("c").
					Build(),
			},
			solution: Solution{
				"a": 1,
				"b": 0,
				"c": 0,
			},
			wantAssume:    []string{"a"},
			wantAssumeNot: nil,
		},
		{
			name: "given selected selection with mixed sub-selections, assumes only selected ids",
			selections: Selections{
				NewSelectionBuilder("a").
					WithSubSelectionID("b").
					WithSubSelectionID("c").
					Build(),
			},
			solution: Solution{
				"a": 1,
				"b": 1,
				"c": 0,
			},
			wantAssume:    []string{"a", "b"},
			wantAssumeNot: nil,
		},
		{
			name: "given unselected selection, ignores its sub-selections",
			selections: Selections{
				NewSelectionBuilder("a").
					WithSubSelectionID("b").
					WithSubSelectionID("c").
					Build(),
			},
			solution: Solution{
				"a": 0,
				"b": 1,
				"c": 1,
			},
			wantAssume:    nil,
			wantAssumeNot: []string{"a"},
		},
		{
			name: "given mixed selections, partitions by solution",
			selections: Selections{
				NewSelectionBuilder("a").WithSubSelectionID("a1").Build(),
				NewSelectionBuilder("b").Build(),
				NewSelectionBuilder("c").WithSubSelectionID("c1").Build(),
			},
			solution: Solution{
				"a":  1,
				"a1": 1,
				"b":  0,
				"c":  0,
				"c1": 1,
			},
			wantAssume:    []string{"a", "a1"},
			wantAssumeNot: []string{"b", "c"},
		},
		{
			name: "given duplicate selected ids, assumes each id once",
			selections: Selections{
				NewSelectionBuilder("a").WithSubSelectionID("b").Build(),
				NewSelectionBuilder("a").WithSubSelectionID("b").Build(),
			},
			solution: Solution{
				"a": 1,
				"b": 1,
			},
			wantAssume:    []string{"a", "b"},
			wantAssumeNot: nil,
		},
		{
			name: "given id selected as selection and as sub-selection, assumes it once",
			selections: Selections{
				NewSelectionBuilder("parent").WithSubSelectionID("a").Build(),
				NewSelectionBuilder("a").Build(),
			},
			solution: Solution{
				"parent": 1,
				"a":      1,
			},
			wantAssume:    []string{"parent", "a"},
			wantAssumeNot: nil,
		},
		{
			name: "given sub-selection selected only under a selected parent, assumes it",
			selections: Selections{
				NewSelectionBuilder("skipped").WithSubSelectionID("shared").Build(),
				NewSelectionBuilder("kept").WithSubSelectionID("shared").Build(),
			},
			solution: Solution{
				"skipped": 0,
				"kept":    1,
				"shared":  1,
			},
			wantAssume:    []string{"kept", "shared"},
			wantAssumeNot: []string{"skipped"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAssume, gotAssumeNot := creator.getIDsToAssume(tt.selections, tt.solution)

			assert.ElementsMatch(t, tt.wantAssume, gotAssume)
			assert.ElementsMatch(t, tt.wantAssumeNot, gotAssumeNot)
		})
	}
}
