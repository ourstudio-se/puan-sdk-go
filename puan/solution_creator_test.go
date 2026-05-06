// nolint:lll
package puan

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
)

func Test_validateSelections_givenIndependentVariableInSubSelection_shouldReturnError(t *testing.T) {
	primaryID := fake.New[string]()
	subID := fake.New[string]()

	creator := NewRulesetCreator()
	_ = creator.AddPrimitives(primaryID, subID)
	_ = creator.Assume(primaryID)
	ruleset, _ := creator.Create()

	selections := Selections{
		NewSelectionBuilder(primaryID).WithSubSelectionID(subID).Build(),
	}

	err := validateSelections(selections, ruleset)

	assert.Error(t, err)
}

func Test_validateSelections_givenIndependentVariableSelectionWithSubSelection_shouldReturnError(t *testing.T) {
	primaryID := fake.New[string]()
	subID := fake.New[string]()

	creator := NewRulesetCreator()
	_ = creator.AddPrimitives(primaryID, subID)
	_ = creator.Assume(subID)
	ruleset, _ := creator.Create()

	selections := Selections{
		NewSelectionBuilder(primaryID).WithSubSelectionID(subID).Build(),
	}

	err := validateSelections(selections, ruleset)

	assert.Error(t, err)
}

func Test_validateSelections_givenNotExistingID_shouldReturnError(t *testing.T) {
	primaryID := fake.New[string]()
	subID := fake.New[string]()

	invalidID := fake.New[string]()
	creator := NewRulesetCreator()
	_ = creator.AddPrimitives(primaryID, subID)
	ruleset, _ := creator.Create()

	selections := Selections{
		NewSelectionBuilder(invalidID).Build(),
	}

	err := validateSelections(selections, ruleset)

	assert.Error(t, err)
}

func Test_validateSelections_givenEmptySelection_shouldReturnNoError(t *testing.T) {
	primaryID := fake.New[string]()
	subID := fake.New[string]()

	creator := NewRulesetCreator()
	_ = creator.AddPrimitives(primaryID, subID)
	ruleset, _ := creator.Create()

	selections := Selections{}

	err := validateSelections(selections, ruleset)

	assert.NoError(t, err)
}

func Test_categorizeSelections(t *testing.T) {
	independentID := fake.New[string]()
	dependentID := fake.New[string]()

	selections := Selections{
		NewSelectionBuilder(independentID).Build(),
		NewSelectionBuilder(dependentID).Build(),
	}

	independentVariables := []string{independentID}

	dependentSelections, independentSelections :=
		categorizeSelections(selections, independentVariables)

	assert.Len(t, dependentSelections, 1)
	assert.Equal(t, dependentID, dependentSelections[0].id)

	assert.Len(t, independentSelections, 1)
	assert.Equal(t, independentID, independentSelections[0].id)
}

func Test_getSelectedIDs(t *testing.T) {
	creator := &SolutionCreator{}

	type theory struct {
		name       string
		selections Selections
		solution   Solution
		want       []string
	}

	theories := []theory{
		{
			name:       "Given empty selections, returns nil",
			selections: Selections{},
			solution:   Solution{"x": 1},
			want:       nil,
		},
		{
			name: "Given no IDs selected in solution, returns nil",
			selections: Selections{
				NewSelectionBuilder("a").Build(),
				NewSelectionBuilder("b").Build(),
			},
			solution: Solution{"a": 0, "b": 0},
			want:     nil,
		},
		{
			name: "Given single primitive selected, returns the ID",
			selections: Selections{
				NewSelectionBuilder("x").Build(),
			},
			solution: Solution{"x": 1},
			want:     []string{"x"},
		},
		{
			name: "Given composite selection, returns the primary and sub-IDs",
			selections: Selections{
				NewSelectionBuilder("p").WithSubSelectionID("s1").WithSubSelectionID("s2").Build(),
			},
			solution: Solution{"p": 1, "s1": 0, "s2": 1},
			want:     []string{"p", "s2"},
		},
		{
			name: "Given 2, is not selected",
			selections: Selections{
				NewSelectionBuilder("q").Build(),
			},
			solution: Solution{"q": 2},
			want:     nil,
		},
		{
			name: "Given selected remove selection, returns the ID",
			selections: Selections{
				NewSelectionBuilder("x").WithAction(REMOVE).Build(),
			},
			solution: Solution{"x": 1},
			want:     []string{"x"},
		},
	}

	for _, tt := range theories {
		t.Run(tt.name, func(t *testing.T) {
			got := creator.getSelectedIDs(tt.selections, tt.solution)
			assert.Equal(t, tt.want, got)
		})
	}
}
