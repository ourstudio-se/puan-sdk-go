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
