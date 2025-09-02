package puan

import "github.com/ourstudio-se/puan-sdk-go/utils"

func CalculateObjective3(
	primitives,
	selectedPrimitives,
	preferredIDs []string,
) Weights {
	notSelectedPrimitives := utils.Without(primitives, selectedPrimitives)
	notSelectedWeights := calculatedNotSelectedWeights(notSelectedPrimitives)
	notSelectedSum := notSelectedWeights.sum()
	preferenceWeights := calculatePreferredWeights3(preferredIDs, notSelectedSum)
	sumOfPreferredWeights := preferenceWeights.sum()
	selectedWeights := calculateSelectedWeights3(
		selectedPrimitives,
		notSelectedSum,
		sumOfPreferredWeights,
	)

	weights := notSelectedWeights.
		Concat(selectedWeights).
		Concat(preferenceWeights)

	return weights
}

func calculatePreferredWeights3(
	preferredIDs []string,
	notSelectedSum int,
) Weights {
	preferredWeights := make(Weights)

	if notSelectedSum == 0 {
		return preferredWeights
	}

	for _, preferredID := range preferredIDs {
		preferredWeights[preferredID] = notSelectedSum + 1
	}

	return preferredWeights
}

func calculateSelectedWeights3(
	selectedPrimitives []string,
	notSelectedSum,
	preferredWeightsSum int,
) Weights {
	selectedWeights := make(Weights)
	worstCase := -(notSelectedSum + preferredWeightsSum)
	previousSelectionWeightSum := worstCase
	for _, selectedPrimitive := range selectedPrimitives {
		weight := previousSelectionWeightSum + 1
		selectedWeights[selectedPrimitive] = weight
		previousSelectionWeightSum += weight
	}

	return selectedWeights
}
