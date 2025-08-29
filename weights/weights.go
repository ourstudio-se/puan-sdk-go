package weights

import (
	"slices"
)

type OptionalPreferred struct {
	PrimitiveID string
	PreferredID string
}

// Helst inte, kan bli komplext.
// axuillary id
type OptionalPreferreds []OptionalPreferred

func (o *OptionalPreferreds) ExtractNonRedundantPreferredIDs(selectedIDs []string) []string {
	var preferredIDs []string
	for _, op := range *o {
		if !slices.Contains(selectedIDs, op.PrimitiveID) {
			continue
		}

		preferredIDs = append(preferredIDs, op.PreferredID)
	}

	return preferredIDs
}

type CompulsoryPreferred struct {
	PrimitiveID string
	PreferredID string
}

type CompulsoryPreferreds []CompulsoryPreferred

func (c *CompulsoryPreferreds) ExtractNonRedundantPreferredIDs(selectedIDs []string) []string {
	var preferredIDs []string
	for _, cp := range *c {
		if slices.Contains(selectedIDs, cp.PrimitiveID) {
			continue
		}

		preferredIDs = append(preferredIDs, cp.PreferredID)
	}

	return preferredIDs
}

type PrioritySelection []string

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

func Create(variables, selectedIDs, preferredIDs []string) Weights {
	weights := newWeights(variables)
	weights.setPreferredWeights(preferredIDs)
	weights.setSelectedWeights(selectedIDs)

	return weights
}

func (w Weights) setPreferredWeights(preferredIDs []string) {
	for _, preferredID := range preferredIDs {
		(w)[preferredID] = 0
	}

	negativeSum := w.calculateSumOfNonSelectedWeights()
	positiveSum := w.calculateSumOfSelectedWeights()

	for _, preferredID := range preferredIDs {
		(w)[preferredID] = negativeSum + positiveSum + 1
	}
}

func (w Weights) setSelectedWeights(selectedIDs []string) {
	for _, selectedID := range selectedIDs {
		(w)[selectedID] = 0
	}

	for _, selectedID := range selectedIDs {
		negativeSum := w.calculateSumOfNonSelectedWeights()
		positiveSum := w.calculateSumOfSelectedWeights()
		(w)[selectedID] = negativeSum + positiveSum + 1
	}
}

func (w Weights) calculateSumOfSelectedWeights() int {
	positiveSum := 0
	for _, weight := range w {
		if weight > 0 {
			positiveSum += weight
		}
	}

	return positiveSum
}

func (w Weights) calculateSumOfNonSelectedWeights() int {
	negativeSum := 0
	for _, weight := range w {
		if weight < 0 {
			negativeSum += -weight
		}
	}

	return negativeSum
}

func newWeights(variables []string) Weights {
	weights := make(Weights, len(variables))
	for _, v := range variables {
		weights[v] = -1
	}

	return weights
}
