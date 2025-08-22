package weights

import (
	"math"

	"github.com/ourstudio-se/puan-sdk-go/utils"
)

const ADD Action = "ADD"
const REMOVE Action = "REMOVE"

type Action string

type Selection struct {
	ID     string
	Action Action
}
type Selections []Selection

func (s *Selections) ExtractActiveSelectionIDS() []string {
	selections := s.extractActiveSelections()
	ids := selections.extractSelectionsIDs()

	return ids
}

func (s *Selections) extractActiveSelections() Selections {
	lastActions := make(map[string]Action)
	for _, selection := range *s {
		lastActions[selection.ID] = selection.Action
	}

	activeSelections := Selections{}
	for _, selection := range *s {
		if action, ok := lastActions[selection.ID]; ok {
			if action == ADD {
				activeSelections = append(activeSelections, selection)
			}
		}
	}

	return activeSelections
}

func (s *Selections) extractSelectionsIDs() []string {
	ids := make([]string, 0, len(*s))
	for _, selection := range *s {
		ids = append(ids, selection.ID)
	}

	return ids
}

type Weights map[string]int

type XORWithPreference struct {
	XORID              string
	PreferredVariantID string
}

func Create(variables, selectedIDs []string, xorsWithPreference []XORWithPreference) Weights {
	notSelectedIDs := utils.Without(variables, selectedIDs)
	preferredWeights := calculatePreferredWeights(notSelectedIDs, xorsWithPreference)

	weights := newWeights(variables)

	weights.setWeights(selectedIDs)

	for id, weight := range preferredWeights {
		weights[id] = weight
	}

	return weights
}

func calculatePreferredWeights(notSelectedIDs []string, xorsWithPreference []XORWithPreference) Weights {
	weights := make(map[string]int)
	notSelectedSum := -2 * len(notSelectedIDs)
	for _, xor := range xorsWithPreference {
		preferenceWeight := math.Abs(float64(notSelectedSum)) - 1
		constraintWeight := notSelectedSum + 2

		weights[xor.PreferredVariantID] = int(preferenceWeight)
		weights[xor.XORID] = constraintWeight
	}

	return weights
}

func (w *Weights) setWeights(selectedIDs []string) {
	for _, selectedID := range selectedIDs {
		(*w)[selectedID] = 0
	}

	for _, selectedID := range selectedIDs {
		negativeSum := w.calculateSumOfNonSelectedWeights()
		positiveSum := w.calculateSumOfSelectedWeights()
		(*w)[selectedID] = negativeSum + positiveSum + 1
	}
}

func (w *Weights) calculateSumOfSelectedWeights() int {
	positiveSum := 0
	for _, weight := range *w {
		if weight > 0 {
			positiveSum += weight
		}
	}

	return positiveSum
}

func (w *Weights) calculateSumOfNonSelectedWeights() int {
	negativeSum := 0
	for _, weight := range *w {
		if weight < 0 {
			negativeSum += -weight
		}
	}

	return negativeSum
}

func newWeights(variables []string) Weights {
	weights := make(Weights, len(variables))
	for _, v := range variables {
		weights[v] = -2
	}

	return weights
}
