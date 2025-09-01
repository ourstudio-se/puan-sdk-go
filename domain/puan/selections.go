package puan

const ADD Action = "ADD"
const REMOVE Action = "REMOVE"

type Action string

type Selection struct {
	ID     string
	Action Action
}
type Selections []Selection

func (s Selections) GetImpactingSelectionIDS() []string {
	selections := s.removeRedundantSelections()
	ids := selections.ids()

	return ids
}

func (s Selections) removeRedundantSelections() Selections {
	lastActions := make(map[string]Action)
	for _, selection := range s {
		lastActions[selection.ID] = selection.Action
	}

	activeSelections := Selections{}
	for _, selection := range s {
		if action, ok := lastActions[selection.ID]; ok {
			if action == ADD {
				activeSelections = append(activeSelections, selection)
			}
		}
	}

	return activeSelections
}

func (s Selections) ids() []string {
	ids := make([]string, len(s))
	for i, selection := range s {
		ids[i] = selection.ID
	}

	return ids
}
