package puan

import (
	"reflect"
	"testing"
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
