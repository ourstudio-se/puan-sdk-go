package puan

import "github.com/ourstudio-se/puan-sdk-go/utils"

type XORWithPreference2 struct {
	PreferredID string
	IDs         []string
}

func (xorWithPreference XORWithPreference2) getNotPreferredIDs() []string {
	notPreferredIDs := make([]string, len(xorWithPreference.IDs)-1)
	for i, id := range xorWithPreference.IDs {
		if id != xorWithPreference.PreferredID {
			notPreferredIDs[i] = id
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

	selectedWeights := calculateSelectedWeights2(selectedPrimitives, nonSelectedAndXORVariantsWeights)

	weights := nonSelectedAndXORVariantsWeights.Concat(selectedWeights)

	return weights
}

func getWeightsOfNonSelectedPrimitivesAndXORVariants(
	notSelectedPrimitives []string,
	xorWithPreferences []XORWithPreference2,
) Weights {
	var variantWeights Weights

	for _, xorWithPreference := range xorWithPreferences {
		for _, id := range xorWithPreference.IDs {
			variantWeights[id] = -1
		}
	}

	var primitivesWeights Weights
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
	nonSelectedAndXORVariantsWeights Weights,
) Weights {
	panic("not implemented")
}
