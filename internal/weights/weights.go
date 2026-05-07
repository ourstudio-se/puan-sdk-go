package weights

import (
	"maps"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
)

const NOT_SELECTED_WEIGHT = -2

// WEIGHTS_SATURATION_LIMIT is set to 2^32.
const WEIGHTS_SATURATION_LIMIT = 4294967296

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

func (w Weights) absSum() (int, error) {
	var values []int
	for _, weight := range w {
		values = append(values, abs(weight))
	}
	return absSum(values...)
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

// Weights can become very large. This can cause integer oveflows
// or problems for the external solver.
func (w Weights) WeightsTooLarge() bool {
	absSum, err := w.absSum()
	if err != nil {
		return true
	}
	tooLarge := absSum > WEIGHTS_SATURATION_LIMIT
	return tooLarge
}

func Calculate(
	selectableIDs []string,
	selections Selections,
	preferredIDs []string,
	periodIDs []string,
) (Weights, error) {
	notSelectedIDs := utils.Without(selectableIDs, selections.ids())

	notSelectedWeights := calculatedNotSelectedWeights(notSelectedIDs)
	notSelectedSum := notSelectedWeights.sum()

	preferredWeights := calculatePreferredWeights(preferredIDs, notSelectedSum)
	preferredSum := preferredWeights.sum()

	periodWeights, err := calculatePeriodWeights(
		periodIDs,
		notSelectedSum,
		preferredSum,
	)
	if err != nil {
		return Weights{}, err
	}

	maxPeriodWeight := periodWeights.maxWeight()

	selectedWeights, err := calculateSelectedWeights(
		selections,
		notSelectedSum,
		preferredSum,
		maxPeriodWeight,
	)
	if err != nil {
		return Weights{}, err
	}

	weights := notSelectedWeights.
		concat(selectedWeights).
		concat(preferredWeights).
		concat(periodWeights)

	return weights, nil
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
	notSelectedSum int,
	preferredWeightsSum int,
) (Weights, error) {
	periodWeights := make(Weights)

	threshold, err := absSum(notSelectedSum, preferredWeightsSum)
	if err != nil {
		return Weights{}, err
	}
	threshold = -threshold

	periodWeightSum := threshold
	for i, periodID := range periodIDs {
		if i == 0 {
			periodWeights[periodID] = 0

			continue
		}

		weight := periodWeightSum - 1
		periodWeights[periodID] = weight
		periodWeightSum += weight
	}

	return periodWeights, nil
}

func calculateSelectedWeights(
	selections Selections,
	notSelectedSum,
	preferredWeightsSum int,
	maxPeriodWeight int,
) (Weights, error) {
	selectedWeights := make(Weights)

	threshold, err := absSum(notSelectedSum, preferredWeightsSum, maxPeriodWeight)
	if err != nil {
		return Weights{}, err
	}

	selectionWeightSum := threshold
	for _, selection := range selections {
		weight := selectionWeightSum + 1
		if selection.action == ADD {
			selectedWeights[selection.id] = weight
		} else {
			selectedWeights[selection.id] = -weight
		}

		selectionWeightSum += weight
	}

	return selectedWeights, nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}

func absSum(terms ...int) (int, error) {
	sum := 0
	for _, term := range terms {
		absTerm := abs(term)
		newSum := sum + absTerm
		if newSum < sum {
			return 0, errors.New("weights sum overflow")
		}
		sum = newSum
	}

	return sum, nil
}
