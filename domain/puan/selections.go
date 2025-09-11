package puan

import (
	"github.com/ourstudio-se/puan-sdk-go/utils"
)

const ADD Action = "ADD"
const REMOVE Action = "REMOVE"

type Action string

type Selection struct {
	id              string
	subSelectionIDs []string
	action          Action
}

type querySelection struct {
	id     string
	action Action
}

type querySelections []querySelection

func (s querySelections) ids() []string {
	ids := make([]string, len(s))
	for i, selection := range s {
		ids[i] = selection.id
	}

	return ids
}

type Selections []Selection

func (s Selection) isComposite() bool {
	return len(s.subSelectionIDs) > 0
}

func newSelection(action Action, id string, subSelectionIDs []string) Selection {
	return Selection{
		id:              id,
		subSelectionIDs: subSelectionIDs,
		action:          action,
	}
}

func (s Selection) ID() string {
	return s.id
}

func (s Selection) ids() []string {
	ids := make([]string, len(s.subSelectionIDs)+1)
	ids[0] = s.id
	copy(ids[1:], s.subSelectionIDs)
	return ids
}

func getImpactingSelections(selectionsOrderedByOccurrence Selections) Selections {
	selectionsOrderedByPriority := utils.Reverse(selectionsOrderedByOccurrence)
	impactingSelectionsOrderedByPriority := filterOutRedundantSelections(selectionsOrderedByPriority)
	impactingSelections := utils.Reverse(impactingSelectionsOrderedByPriority)

	return impactingSelections
}

func filterOutRedundantSelections(
	selectionsOrderedByPriority Selections,
) Selections {
	var filtered Selections
	for _, selection := range selectionsOrderedByPriority {
		if selection.isRedundant(filtered) {
			continue
		}

		filtered = append(filtered, selection)
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
	if utils.ContainsAll(other.ids(), s.ids()) {
		return true
	}

	if s.action == REMOVE && utils.ContainsAny(other.ids(), s.subSelectionIDs) {
		return true
	}

	if s.id != other.id {
		return false
	}

	if utils.ContainsAll(other.ids(), s.subSelectionIDs) {
		return true
	}

	prioritisedIsNotComposite := !s.isComposite()

	return prioritisedIsNotComposite
}

func (s Selections) extendWithPrimaryPrimitiveSelections() Selections {
	extendedSelections := Selections{}
	for _, selection := range s {
		if selection.isComposite() && selection.action == ADD {
			primaryPrimitiveSelection := NewSelectionBuilder(selection.id).
				WithAction(selection.action).
				Build()
			extendedSelections = append(extendedSelections, primaryPrimitiveSelection)
		}

		extendedSelections = append(extendedSelections, selection)
	}

	return extendedSelections
}
