package weights

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_concat_noSharedVariables(t *testing.T) {
	w1 := Weights{
		"x": 1,
		"y": 2,
	}

	w2 := Weights{
		"z": 3,
	}

	expected := Weights{
		"x": 1,
		"y": 2,
		"z": 3,
	}

	actual := w1.concat(w2)
	assert.Equal(t, expected, actual)
}

func Test_concat_withSharedVariables(t *testing.T) {
	w1 := Weights{
		"x": 1,
		"y": 2,
	}

	w2 := Weights{
		"x": 3,
	}

	expected := Weights{
		"x": 3,
		"y": 2,
	}

	actual := w1.concat(w2)
	assert.Equal(t, expected, actual)
}

func Test_weights_sum(t *testing.T) {
	w := Weights{
		"x": 1,
		"y": 2,
		"z": -4,
	}

	actual := w.sum()
	expected := -1
	assert.Equal(t, expected, actual)
}

func Test_calculatedNotSelectedWeights_shouldReturnWeights(t *testing.T) {
	notSelectedVariables := []string{"x", "y", "z"}
	actual := calculatedNotSelectedWeights(notSelectedVariables)
	expected := Weights{
		"x": -2,
		"y": -2,
		"z": -2,
	}

	assert.Equal(t, expected, actual)
}

func Test_calculatedNotSelectedWeights_givenNoVariables_shouldReturnEmptyWeights(t *testing.T) {
	var notSelectedVariables []string
	actual := calculatedNotSelectedWeights(notSelectedVariables)
	expected := Weights{}

	assert.Equal(t, expected, actual)
}

func Test_calculatePreferredWeights_shouldReturnPreferredWeights(t *testing.T) {
	preferredIDs := []string{"x", "y", "z"}
	notSelectedSum := -10

	actual := calculatePreferredWeights(preferredIDs, notSelectedSum)
	expected := Weights{
		"x": -9,
		"y": -9,
		"z": -9,
	}

	assert.Equal(t, expected, actual)
}

func Test_calculatePreferredWeights_givenNoPreferredIDs_shouldReturnEmptyWeights(t *testing.T) {
	notSelectedSum := 0
	actual := calculatePreferredWeights(nil, notSelectedSum)
	expected := Weights{}

	assert.Equal(t, expected, actual)
}

func Test_calculateSelectedWeights_oneSelected_shouldReturnWeights(t *testing.T) {
	selections := Selections{
		{
			id:     "a",
			action: ADD,
		},
	}
	notSelectedSum := -2
	preferredWeightsSum := -1

	actual := calculateSelectedWeights(selections, notSelectedSum, preferredWeightsSum, 0)
	expected := Weights{
		"a": 4,
	}

	assert.Equal(t, expected, actual)
}

func Test_calculateSelectedWeights_twoSelected_shouldReturnWeights(t *testing.T) {
	selections := Selections{
		{
			id:     "a",
			action: ADD,
		},
		{
			id:     "b",
			action: ADD,
		},
	}
	notSelectedSum := -4
	preferredWeightsSum := -2

	actual := calculateSelectedWeights(selections, notSelectedSum, preferredWeightsSum, 0)
	expected := Weights{
		"a": 7,
		"b": 14,
	}

	assert.Equal(t, expected, actual)
}

func Test_calculateSelectedWeights_twoSelected_withRemoveAction(t *testing.T) {
	selections := Selections{
		{
			id:     "a",
			action: ADD,
		},
		{
			id:     "b",
			action: REMOVE,
		},
	}
	notSelectedSum := -4
	preferredWeightsSum := -2

	actual := calculateSelectedWeights(selections, notSelectedSum, preferredWeightsSum, 0)
	expected := Weights{
		"a": 7,
		"b": -14,
	}

	assert.Equal(t, expected, actual)
}

func Test_calculateSelectedWeights_noSelection_shouldReturnEmptyWeights(t *testing.T) {
	notSelectedSum := -1
	preferredWeightsSum := -1

	actual := calculateSelectedWeights(nil, notSelectedSum, preferredWeightsSum, 0)
	expected := Weights{}

	assert.Equal(t, expected, actual)
}

func Test_calculateSelectedWeights_givenSelectionsPreferredWeightsAndPeriodWeight(
	t *testing.T,
) {
	selections := Selections{
		{
			id:     "a",
			action: ADD,
		},
		{
			id:     "b",
			action: ADD,
		},
	}
	notSelectedSum := -4
	preferredWeightsSum := -2
	minPeriodWeight := -8

	actual := calculateSelectedWeights(
		selections,
		notSelectedSum,
		preferredWeightsSum,
		minPeriodWeight,
	)
	expected := Weights{
		"a": 15,
		"b": 30,
	}

	assert.Equal(t, expected, actual)
}

func Test_calculateWeights(t *testing.T) {
	primitives := []string{"a", "b", "c"}
	preferredIDs := []string{"e"}
	selections := Selections{
		{
			id:     "a",
			action: ADD,
		},
	}

	actual := Calculate(primitives, selections, preferredIDs, nil)
	expected := Weights{
		"a": 8,
		"b": -2,
		"c": -2,
		"e": -3,
	}

	assert.Equal(t, expected, actual)
}

func Test_calculateSelectionThreshold(t *testing.T) {
	testCases := []struct {
		name                string
		notSelectedSum      int
		preferredWeightsSum int
		minPeriodWeight     int
		expected            int
	}{
		{
			name:                "all zeros",
			notSelectedSum:      0,
			preferredWeightsSum: 0,
			minPeriodWeight:     0,
			expected:            0,
		},
		{
			name:                "all negative values",
			notSelectedSum:      -10,
			preferredWeightsSum: -5,
			minPeriodWeight:     -3,
			expected:            18,
		},
		{
			name:                "all positive values",
			notSelectedSum:      10,
			preferredWeightsSum: 5,
			minPeriodWeight:     3,
			expected:            -18,
		},
		{
			name:                "mixed values",
			notSelectedSum:      -4,
			preferredWeightsSum: -2,
			minPeriodWeight:     0,
			expected:            6,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := calculateSelectionThreshold(
				tc.notSelectedSum,
				tc.preferredWeightsSum,
				tc.minPeriodWeight,
			)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func Test_abs(t *testing.T) {
	theories := []struct {
		input    int
		expected int
	}{
		{input: 1, expected: 1},
		{input: -1, expected: 1},
		{input: 0, expected: 0},
	}

	for _, theory := range theories {
		actual := abs(theory.input)
		assert.Equal(t, theory.expected, actual)
	}
}

func Test_Weights_MaxWeight(t *testing.T) {
	theories := []struct {
		weights  Weights
		expected int
	}{
		{
			weights: Weights{
				"a": 1,
				"b": 2,
			},
			expected: 2,
		},
		{
			weights: Weights{
				"a": -1,
				"b": -2,
			},
			expected: 2,
		},
		{
			weights: Weights{
				"a": -2,
				"b": 1,
			},
			expected: 2,
		},
	}

	for _, theory := range theories {
		actual := theory.weights.MaxWeight()
		assert.Equal(t, theory.expected, actual)
	}
}
