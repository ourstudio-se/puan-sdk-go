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
			name:          "no selections returns empty",
			selections:    nil,
			solution:      Solution{"a": 1},
			wantAssume:    nil,
			wantAssumeNot: nil,
		},
		{
			name: "selected selection is assumed",
			selections: Selections{
				NewSelectionBuilder("a").Build(),
			},
			solution:      Solution{"a": 1},
			wantAssume:    []string{"a"},
			wantAssumeNot: nil,
		},
		{
			name: "unselected selection is assumed not",
			selections: Selections{
				NewSelectionBuilder("a").Build(),
			},
			solution:      Solution{"a": 0},
			wantAssume:    nil,
			wantAssumeNot: []string{"a"},
		},
		{
			name: "selection missing from solution is assumed not",
			selections: Selections{
				NewSelectionBuilder("a").Build(),
			},
			solution:      Solution{},
			wantAssume:    nil,
			wantAssumeNot: []string{"a"},
		},
		{
			name: "assumes selected sub-selections",
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
			name: "unselected sub-selections are neither assumed nor assumed not",
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
			name: "duplicate selected ids are assumed once",
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
			name: "id selected as selection and as sub-selection is assumed once",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAssume, gotAssumeNot := creator.getIDsToAssume(tt.selections, tt.solution)

			assert.ElementsMatch(t, tt.wantAssume, gotAssume)
			assert.ElementsMatch(t, tt.wantAssumeNot, gotAssumeNot)
		})
	}
}
