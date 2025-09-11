package puan

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
	selections := querySelections{
		{
			id:     "a",
			action: ADD,
		},
	}
	notSelectedSum := -2
	preferredWeightsSum := -1

	actual := calculateSelectedWeights(selections, notSelectedSum, preferredWeightsSum)
	expected := Weights{
		"a": 4,
	}

	assert.Equal(t, expected, actual)
}

func Test_calculateSelectedWeights_twoSelected_shouldReturnWeights(t *testing.T) {
	selections := querySelections{
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

	actual := calculateSelectedWeights(selections, notSelectedSum, preferredWeightsSum)
	expected := Weights{
		"a": 7,
		"b": 14,
	}

	assert.Equal(t, expected, actual)
}

func Test_calculateSelectedWeights_noSelection_shouldReturnEmptyWeights(t *testing.T) {
	notSelectedSum := -1
	preferredWeightsSum := -1

	actual := calculateSelectedWeights(nil, notSelectedSum, preferredWeightsSum)
	expected := Weights{}

	assert.Equal(t, expected, actual)
}

func Test_CalculateObjective(t *testing.T) {
	primitives := []string{"a", "b", "c"}
	preferredIDs := []string{"e"}
	selections := querySelections{
		{
			id:     "a",
			action: ADD,
		},
	}

	actual := calculateObjective(primitives, selections, preferredIDs)
	expected := Weights{
		"a": 8,
		"b": -2,
		"c": -2,
		"e": -3,
	}

	assert.Equal(t, expected, actual)
}
