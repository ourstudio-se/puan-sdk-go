package puan

import (
	"reflect"
	"testing"

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
			extracted, err := tt.solution.Extract(tt.variables...)

			assert.NoError(t, err)
			if !reflect.DeepEqual(extracted, tt.expected) {
				t.Errorf("Expected %v, but got %v", tt.expected, extracted)
			}
		})
	}
}

func Test_Solution_Extract_invalidCases(t *testing.T) {
	tests := []struct {
		name      string
		solution  Solution
		variables []string
	}{
		{
			name: "givenNonExistentVariable_shouldReturnZeroValue",
			solution: Solution{
				"x": 10,
				"y": 20,
			},
			variables: []string{"non_existent"},
		},
		{
			name: "givenMixedExistingAndNonExistingVariables_shouldReturnMixedValues",
			solution: Solution{
				"x": 10,
				"y": 20,
			},
			variables: []string{"x", "non_existent", "y"},
		},
		{
			name:      "givenEmptySolution_shouldReturnEmptySolution",
			solution:  Solution{},
			variables: []string{"x", "y"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.solution.Extract(tt.variables...)
			assert.Error(t, err)
		})
	}
}
