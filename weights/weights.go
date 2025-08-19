package weights

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
	ids := make([]string, len(*s))
	for _, selection := range *s {
		ids = append(ids, selection.ID)
	}

	return ids
}

type Weights map[string]int

func Create(variables []string, selectedIDs []string) Weights {
	weights := newWeights(variables)
	if len(selectedIDs) == 0 {
		return weights
	}

	weights.setWeights(selectedIDs)

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
		weights[v] = -1
	}

	return weights
}
