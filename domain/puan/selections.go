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

func removeRedundantSelections(selectionsOrderedByOccurence Selections) Selections {
	selectionsOrderedByPriority := utils.Reverse(selectionsOrderedByOccurence)

	impactingSelectionsOrderedByPriority := filterOutRedundantSelections(selectionsOrderedByPriority)

	addSelectionsOrderedByPriority := impactingSelectionsOrderedByPriority.filterOutRemoveSelections()

	impactingSelections := utils.Reverse(addSelectionsOrderedByPriority)

	return impactingSelections
}

func filterOutRedundantSelections(
	selectionsOrderedByPriority Selections,
) Selections {
	var filtered Selections
	for _, selection := range selectionsOrderedByPriority {
		if selection.isRedundant(selectionsOrderedByPriority) {
			continue
		}

		filtered = append(filtered, selection)
	}

	return filtered
}

func (s Selections) filterOutRemoveSelections() Selections {
	var filtered Selections
	for _, selection := range s {
		isRemove := selection.action == REMOVE
		if !isRemove {
			filtered = append(filtered, selection)
		}
	}

	return filtered
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
