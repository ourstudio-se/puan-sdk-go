package puan

import "github.com/ourstudio-se/puan-sdk-go/utils"

type XORWithPreference2 struct {
	PreferredID string
	IDs         []string
}

func (xorWithPreference XORWithPreference2) getNotPreferredIDs() []string {
	var notPreferredIDs []string
	for _, id := range xorWithPreference.IDs {
		if id != xorWithPreference.PreferredID {
			notPreferredIDs = append(notPreferredIDs, id)
		}
	}
	return notPreferredIDs
}

func CalculateObjective2(
	primitives,
	selectedPrimitives []string,
	xorWithPreference []XORWithPreference2,
) Weights {
	notSelectedPrimitives := utils.Without(primitives, selectedPrimitives)

	notSelectedPrimitivesWeights := make(Weights)
	for _, xor := range notSelectedPrimitives {
		notSelectedPrimitivesWeights[xor] = -1
	}

	weightsOfNonSelectedPrimitivesAndXORVariants := getWeightsOfNonSelectedPrimitivesAndXORVariants(
		notSelectedPrimitives,
		xorWithPreference,
	)
	nonPreferredWeights := calculateNonPreferredWeights(
		xorWithPreference,
		weightsOfNonSelectedPrimitivesAndXORVariants,
	)
	nonSelectedAndXORVariantsWeights := weightsOfNonSelectedPrimitivesAndXORVariants.Concat(
		nonPreferredWeights,
	)

	selectedWeights := calculateSelectedWeights2(
		selectedPrimitives,
		xorWithPreference,
		nonSelectedAndXORVariantsWeights,
		notSelectedPrimitivesWeights,
	)

	weights := nonSelectedAndXORVariantsWeights.Concat(selectedWeights)

	return weights
}

func getWeightsOfNonSelectedPrimitivesAndXORVariants(
	notSelectedPrimitives []string,
	xorWithPreferences []XORWithPreference2,
) Weights {
	variantWeights := make(Weights)

	for _, xorWithPreference := range xorWithPreferences {
		for _, id := range xorWithPreference.IDs {
			variantWeights[id] = -1
		}
	}

	primitivesWeights := make(Weights)
	for _, xor := range notSelectedPrimitives {
		primitivesWeights[xor] = -1
	}

	weights := variantWeights.Concat(primitivesWeights)

	return weights
}

func calculateNonPreferredWeights(
	xorWithPreference []XORWithPreference2,
	weightsOfNonSelectedPrimitivesAndXORVariants Weights,
) Weights {
	sumOfNonSelectedPrimitivesAndXORVariants := weightsOfNonSelectedPrimitivesAndXORVariants.sum()

	notPreferredWeights := make(Weights)
	for _, xorWithPreference := range xorWithPreference {
		notPreferredIDs := xorWithPreference.getNotPreferredIDs()
		nrOfVariables := len(notPreferredIDs)
		for _, id := range notPreferredIDs {
			notPreferredWeights[id] = sumOfNonSelectedPrimitivesAndXORVariants + nrOfVariables - 2
		}
	}

	return notPreferredWeights
}

func calculateSelectedWeights2(
	selectedPrimitives []string,
	xorWithPreference []XORWithPreference2,
	nonSelectedAndXORVariantsWeights Weights,
	notSelectedPrimitivesWeights Weights,
) Weights {
	selectedWeights := make(Weights)

	sumOfNonPreferredWeightInXORs := 0
	for _, xorWithPreference := range xorWithPreference {
		notPreferredIDs := xorWithPreference.getNotPreferredIDs()
		firstNonPreferredID := notPreferredIDs[0]
		sumOfNonPreferredWeightInXORs += nonSelectedAndXORVariantsWeights[firstNonPreferredID]
	}

	sumOfNonSelectedAndXORVariantsWeights := notSelectedPrimitivesWeights.sum() +
		sumOfNonPreferredWeightInXORs

	previousSelectionWeightSum := sumOfNonSelectedAndXORVariantsWeights * -1
	for _, selectedPrimitive := range selectedPrimitives {
		weight := previousSelectionWeightSum + 1
		selectedWeights[selectedPrimitive] = weight
		previousSelectionWeightSum += weight
	}

	return selectedWeights
}
