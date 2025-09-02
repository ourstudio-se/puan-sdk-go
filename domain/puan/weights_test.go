package puan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Weights_concat(t *testing.T) {
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

func Test_calculatedNotSelectedWeights_shouldReturnEmptyWeights(t *testing.T) {
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

func Test_calculatePreferredWeights_shouldReturnEmptyWeights(t *testing.T) {
	notSelectedSum := 0
	actual := calculatePreferredWeights(nil, notSelectedSum)
	expected := Weights{}

	assert.Equal(t, expected, actual)
}

func Test_calculateSelectedWeights_oneSelected_shouldReturnWeights(t *testing.T) {
	selectedPrimitives := []string{"a"}
	notSelectedSum := -2
	preferredWeightsSum := -1

	actual := calculateSelectedWeights(selectedPrimitives, notSelectedSum, preferredWeightsSum)
	expected := Weights{
		"a": 4,
	}

	assert.Equal(t, expected, actual)
}

func Test_calculateSelectedWeights_twoSelected_shouldReturnWeights(t *testing.T) {
	selectedPrimitives := []string{"a", "b"}
	notSelectedSum := -4
	preferredWeightsSum := -2

	actual := calculateSelectedWeights(selectedPrimitives, notSelectedSum, preferredWeightsSum)
	expected := Weights{
		"a": 7,
		"b": 14,
	}

	assert.Equal(t, expected, actual)
}

func Test_calculateSelectedWeights_noSelection_shouldReturnEmptyWeights(t *testing.T) {
	var selectedPrimitives []string
	notSelectedSum := -1
	preferredWeightsSum := -1

	actual := calculateSelectedWeights(selectedPrimitives, notSelectedSum, preferredWeightsSum)
	expected := Weights{}

	assert.Equal(t, expected, actual)
}

func Test_CalculateObjective(t *testing.T) {
	primitives := []string{"a", "b", "c"}
	selectedPrimitives := []string{"a"}
	preferredIDs := []string{"e"}

	actual := CalculateObjective(primitives, selectedPrimitives, preferredIDs)
	expected := Weights{
		"a": 8,
		"b": -2,
		"c": -2,
		"e": -3,
	}

	assert.Equal(t, expected, actual)
}
