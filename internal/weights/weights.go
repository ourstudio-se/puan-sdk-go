package weights

import (
	"maps"

	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
)

const NOT_SELECTED_WEIGHT = -2

// weights on periods are [0, -12, -24, ..., -12(n-1)]
// where n is the number of periods
// periods are assumed to be ordered by start time
//
// This constant can be tweaked to change the weight of periods
const PERIOD_WEIGHT_MULTIPLIER = NOT_SELECTED_WEIGHT * 6

// WEIGHT_SATURATION_LIMIT is set to 2^55.
// The limit is set lower than 2^63 to allow for some headroom
// and enable early detection before reaching the overflow limit.
// Note: Weights for selected variables increase exponentially with the number of selections.
// The highest weight is calculated as 2^(n-1) * (c + 1),
// where n is the number of selections and c is a constant
// derived from the sum of none selected primitives, preferred, and period weights.
const WEIGHT_SATURATION_LIMIT = 36028797018963968

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

func (w Weights) maxWeight() int {
	maxWeight := 0
	for _, weight := range w {
		absWeight := abs(weight)
		if absWeight > maxWeight {
			maxWeight = absWeight
		}
	}

	return maxWeight
}

func (w Weights) ContainsTooLargeWeight() bool {
	tooLarge := w.maxWeight() > WEIGHT_SATURATION_LIMIT
	return tooLarge
}

func Calculate(
	selectableIDs []string,
	selections Selections,
	preferredIDs []string,
	periodIDs []string,
) Weights {
	notSelectedIDs := utils.Without(selectableIDs, selections.ids())
	notSelectedWeights := calculatedNotSelectedWeights(notSelectedIDs)
	notSelectedSum := notSelectedWeights.sum()

	preferredWeights := calculatePreferredWeights(preferredIDs, notSelectedSum)
	sumOfPreferredWeights := preferredWeights.sum()

	periodWeights := calculatePeriodWeights(periodIDs)
	minPeriodWeight := calculateMinPeriodWeight(periodIDs)

	selectedWeights := calculateSelectedWeights(
		selections,
		notSelectedSum,
		sumOfPreferredWeights,
		minPeriodWeight,
	)

	weights := notSelectedWeights.
		concat(selectedWeights).
		concat(preferredWeights).
		concat(periodWeights)

	return weights
}

func calculatedNotSelectedWeights(selectableIDs []string) Weights {
	notSelectedWeights := make(Weights)
	for _, id := range selectableIDs {
		notSelectedWeights[id] = NOT_SELECTED_WEIGHT
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

func calculatePeriodWeights(
	periodIDs []string,
) Weights {
	periodWeights := make(Weights)
	for i, periodID := range periodIDs {
		periodWeights[periodID] = i * PERIOD_WEIGHT_MULTIPLIER
	}

	return periodWeights
}

func calculateMinPeriodWeight(
	periodIDs []string,
) int {
	if len(periodIDs) == 0 {
		return 0
	}

	return (len(periodIDs) - 1) * PERIOD_WEIGHT_MULTIPLIER
}

func calculateSelectedWeights(
	selections Selections,
	notSelectedSum,
	preferredWeightsSum int,
	minPeriodWeight int,
) Weights {
	selectedWeights := make(Weights)

	selectionThreshold := calculateSelectionThreshold(
		notSelectedSum,
		preferredWeightsSum,
		minPeriodWeight,
	)

	selectionWeightSum := selectionThreshold
	for _, selection := range selections {
		weight := selectionWeightSum + 1
		if selection.action == ADD {
			selectedWeights[selection.id] = weight
		} else {
			selectedWeights[selection.id] = -weight
		}

		selectionWeightSum += weight
	}

	return selectedWeights
}

func calculateSelectionThreshold(
	notSelectedSum,
	preferredWeightsSum,
	minPeriodWeight int,
) int {
	return -(notSelectedSum + preferredWeightsSum + minPeriodWeight)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}
