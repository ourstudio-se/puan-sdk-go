package puan

import (
	"github.com/ourstudio-se/puan-sdk-go/utils"
)

const ADD Action = "ADD"
const REMOVE Action = "REMOVE"

type Action string

type Selection struct {
	id             string
	subSelectionID *string
	action         Action
}

type Selections []Selection

func newSelection(action Action, id string, subSelectionID *string) Selection {
	return Selection{
		id:             id,
		subSelectionID: subSelectionID,
		action:         action,
	}
}

func (s Selection) ID() string {
	return s.id
}

func (s Selections) getImpactingSelections() Selections {
	selections := s.removeRedundantSelections()

	return selections
}

func (s Selections) removeRedundantSelections() Selections {
	selectionsByPriority := utils.Reverse(s)

	impactingSelectionsByPriority := Selections{}
	for _, selection := range selectionsByPriority {
		if selection.isRedundant(impactingSelectionsByPriority) {
			continue
		}

		impactingSelectionsByPriority = append(impactingSelectionsByPriority, selection)
	}

	addSelectionsByPriority := Selections{}
	for _, selection := range impactingSelectionsByPriority {
		if selection.action == ADD {
			addSelectionsByPriority = append(addSelectionsByPriority, selection)
		}
	}

	impactingSelections := utils.Reverse(addSelectionsByPriority)

	return impactingSelections
}

func (s Selection) isRedundant(existingSelections Selections) bool {
	for _, existingSelection := range existingSelections {
		if existingSelection.makesRedundant(s) {
			return true
		}
	}

	return false
}

func (s Selection) makesRedundant(other Selection) bool {
	if s.id != other.id {
		return false
	}

	if s.subSelectionID == nil {
		return true
	}

	if other.subSelectionID == nil {
		return false
	}

	if *other.subSelectionID == *s.subSelectionID {
		return true
	}

	return false
}
