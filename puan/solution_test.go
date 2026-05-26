package puan

import (
	"reflect"
	"testing"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
	"github.com/stretchr/testify/assert"
)

func Test_Solution_Extract_validCases(t *testing.T) {
	tests := []struct {
		name      string
		solution  Solution
		variables []string
		expected  Solution
	}{
		{
			name: "givenSingleVariable_shouldReturnThatVariable",
			solution: Solution{
				"x": 10,
				"y": 20,
				"z": 30,
			},
			variables: []string{"x"},
			expected:  Solution{"x": 10},
		},
		{
			name: "givenMultipleVariables_shouldReturnThoseVariables",
			solution: Solution{
				"x": 10,
				"y": 20,
				"z": 30,
				"w": 40,
			},
			variables: []string{"x", "z"},
			expected: Solution{
				"x": 10,
				"z": 30,
			},
		},
		{
			name: "givenNoVariables_shouldReturnEmptySolution",
			solution: Solution{
				"x": 10,
				"y": 20,
			},
			variables: []string{},
			expected:  Solution{},
		},
		{
			name: "givenDuplicateVariables_shouldReturnUniqueValues",
			solution: Solution{
				"x": 10,
				"y": 20,
			},
			variables: []string{"x", "x", "y"},
			expected: Solution{
				"x": 10,
				"y": 20,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extracted := tt.solution.Extract(tt.variables...)

			if !reflect.DeepEqual(extracted, tt.expected) {
				t.Errorf("Expected %v, but got %v", tt.expected, extracted)
			}
		})
	}
}

func Test_Solution_merge(t *testing.T) {
	tests := []struct {
		name     string
		solution Solution
		other    Solution
		want     Solution
	}{
		{
			name: "Given non-overlapping keys, includes values from both solutions",
			solution: Solution{
				"x": 1,
			},
			other: Solution{
				"y": 0,
			},
			want: Solution{
				"x": 1,
				"y": 0,
			},
		},
		{
			name: "Given overlapping keys, values from other override existing",
			solution: Solution{
				"x": 1,
				"y": 1,
			},
			other: Solution{
				"y": 0,
				"z": 1,
			},
			want: Solution{
				"x": 1,
				"y": 0,
				"z": 1,
			},
		},
		{
			name: "Given empty other solution, keeps original values",
			solution: Solution{
				"x": 1,
			},
			other: Solution{},
			want: Solution{
				"x": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.solution.merge(tt.other)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_Solution_copy(t *testing.T) {
	original := Solution{
		"x": 1,
		"y": 0,
	}

	copied := original.copy()
	copied["x"] = 0
	copied["z"] = 1

	assert.Equal(t, Solution{"x": 1, "y": 0}, original)
	assert.Equal(t, Solution{"x": 0, "y": 0, "z": 1}, copied)
}

func Test_Solution_isSelected(t *testing.T) {
	variableID := fake.New[string]()
	tests := []struct {
		name     string
		solution Solution
		want     bool
	}{
		{
			name: "Given 1, is selected",
			solution: Solution{
				variableID: 1,
			},
			want: true,
		},
		{
			name: "Given 0, is not selected",
			solution: Solution{
				variableID: 0,
			},
			want: false,
		},
		{
			name:     "Given missing, is not selected",
			solution: Solution{},
			want:     false,
		},
		{
			name: "Given 2, is not selected",
			solution: Solution{
				variableID: 2,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.solution.isSelected(variableID)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_NewSolutionsBySelectionEnvelope(t *testing.T) {
	selection1 := NewSelectionBuilder("x").Build()
	solution1 := Solution{"x": 1}
	selection2 := NewSelectionBuilder("y").Build()
	solution2 := Solution{"y": 1}
	solutions := []SolutionBySelection{
		{selection: selection1, solution: solution1},
		{selection: selection2, solution: solution2},
	}

	envelope, err := NewSolutionsBySelectionEnvelope(solutions)

	assert.NoError(t, err)

	assert.Equal(
		t,
		map[string]SolutionBySelection{
			selection1.Hash(): {selection: selection1, solution: solution1},
			selection2.Hash(): {selection: selection2, solution: solution2},
		},
		envelope.solutionsBySelection,
	)
}

func Test_NewSolutionsBySelectionEnvelope_givenDuplicateSelection_shouldReturnError(
	t *testing.T,
) {
	selection := NewSelectionBuilder(fake.New[string]()).Build()
	solutions := []SolutionBySelection{
		{selection: selection},
		{selection: selection},
	}

	_, err := NewSolutionsBySelectionEnvelope(solutions)

	assert.Error(t, err)
}

func Test_SolutionsBySelectionEnvelope_GetSolutionBySelection(t *testing.T) {
	selection := NewSelectionBuilder("x").Build()
	expected := SolutionBySelection{
		selection: selection,
		solution:  Solution{"x": 1},
	}

	envelope, err := NewSolutionsBySelectionEnvelope([]SolutionBySelection{expected})
	assert.NoError(t, err)

	got, err := envelope.GetSolutionBySelection(selection)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

// nolint:lll
func Test_SolutionsBySelectionEnvelope_GetSolutionBySelection_givenMissingSelection_shouldReturnError(
	t *testing.T,
) {
	envelope := SolutionsBySelectionEnvelope{}

	_, err := envelope.GetSolutionBySelection(NewSelectionBuilder("y").Build())

	assert.Error(t, err)
}
