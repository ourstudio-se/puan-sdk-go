package weights

import (
	"maps"

	"github.com/ourstudio-se/puan-sdk-go/utils"
)

const NOT_SELECTED_WEIGHT = -2

type XORWithPreference struct {
	XORID       string
	PreferredID string
}

func CalculateObjective(
	primitives,
	selectedPrimitives []string,
	xorWithPreference []XORWithPreference,
) Weights {
	notSelectedPrimitives := utils.Without(primitives, selectedPrimitives)
	notSelectedWeights := calculatedNotSelectedWeights(notSelectedPrimitives)
	notSelectedSum := notSelectedWeights.sum()
	xorWeights, preferenceWeights := calculatePreferredWeights(xorWithPreference, notSelectedSum)
	sumOfPreferredWeights := preferenceWeights.sum()
	selectedWeights := calculateSelectedWeights(
		selectedPrimitives,
		notSelectedSum,
		sumOfPreferredWeights,
	)

	weights := make(Weights)
	maps.Copy(weights, notSelectedWeights)
	maps.Copy(weights, selectedWeights)
	maps.Copy(weights, xorWeights)
	maps.Copy(weights, preferenceWeights)

	return weights
}

func (w *Weights) sum() int {
	sum := 0
	for _, weight := range *w {
		sum += weight
	}

	return sum
}

func calculateSelectedWeights(
	selectedPrimitives []string,
	notSelectedSum,
	preferredWeightsSum int,
) Weights {
	selectedWeights := make(Weights)
	worstCase := -notSelectedSum + preferredWeightsSum
	previousSelectionWeightSum := worstCase
	for _, selectedPrimitive := range selectedPrimitives {
		weight := previousSelectionWeightSum + 1
		selectedWeights[selectedPrimitive] = weight
		previousSelectionWeightSum += weight
	}

	return selectedWeights
}

func calculatedNotSelectedWeights(primitives []string) Weights {
	notSelectedWeights := make(Weights)
	for _, primitive := range primitives {
		notSelectedWeights[primitive] = NOT_SELECTED_WEIGHT
	}

	return notSelectedWeights
}

func calculatePreferredWeights(
	xorWithPreference []XORWithPreference,
	notSelectedSum int,
) (Weights, Weights) {
	xorWeights := make(Weights)
	preferenceWeights := make(Weights)

	if notSelectedSum == 0 {
		return xorWeights, preferenceWeights
	}

	for _, xor := range xorWithPreference {
		preferenceWeights[xor.PreferredID] = -notSelectedSum - 1
		xorWeights[xor.XORID] = notSelectedSum - NOT_SELECTED_WEIGHT
	}

	return xorWeights, preferenceWeights
}
