package puan

import (
	"maps"

	"github.com/ourstudio-se/puan-sdk-go/utils"
)

const NOT_SELECTED_WEIGHT = -2

type Weights map[string]int

func (w Weights) concat(weightsToConcat Weights) Weights {
	weights := make(Weights)
	maps.Copy(weights, w)
	maps.Copy(weights, weightsToConcat)

	return weights
}

func (w Weights) sum() int {
	sum := 0
	for _, weight := range w {
		sum += weight
	}

	return sum
}

func calculateObjective(
	primitives,
	selectedIDs,
	preferredIDs []string,
) Weights {
	notSelectedPrimitives := utils.Without(primitives, selectedIDs)
	notSelectedWeights := calculatedNotSelectedWeights(notSelectedPrimitives)
	notSelectedSum := notSelectedWeights.sum()
	preferenceWeights := calculatePreferredWeights(preferredIDs, notSelectedSum)
	sumOfPreferredWeights := preferenceWeights.sum()
	selectedWeights := calculateSelectedWeights(
		selectedIDs,
		notSelectedSum,
		sumOfPreferredWeights,
	)

	weights := notSelectedWeights.
		concat(selectedWeights).
		concat(preferenceWeights)

	return weights
}

func calculatedNotSelectedWeights(primitives []string) Weights {
	notSelectedWeights := make(Weights)
	for _, primitive := range primitives {
		notSelectedWeights[primitive] = NOT_SELECTED_WEIGHT
	}

	return notSelectedWeights
}

func calculatePreferredWeights(
	preferredIDs []string,
	notSelectedSum int,
) Weights {
	preferredWeights := make(Weights)

	if notSelectedSum == 0 {
		return preferredWeights
	}

	weight := notSelectedSum + 1
	for _, preferredID := range preferredIDs {
		preferredWeights[preferredID] = weight
	}

	return preferredWeights
}

func calculateSelectedWeights(
	selectedPrimitives []string,
	notSelectedSum,
	preferredWeightsSum int,
) Weights {
	selectedWeights := make(Weights)
	selectionThreshold := -(notSelectedSum + preferredWeightsSum)
	selectionWeightSum := selectionThreshold
	for _, selectedPrimitive := range selectedPrimitives {
		weight := selectionWeightSum + 1
		selectedWeights[selectedPrimitive] = weight
		selectionWeightSum += weight
	}

	return selectedWeights
}
