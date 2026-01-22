package puan

import (
	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
)

const (
	ADD    Action = "ADD"
	REMOVE Action = "REMOVE"
)

type Action string

type Selection struct {
	id              string
	subSelectionIDs []string
	action          Action
}

type Selections []Selection

func (s Selections) ids() []string {
	var ids []string
	seen := make(map[string]bool, len(s))
	for _, selection := range s {
		for _, id := range selection.ids() {
			if !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
	}

	return ids
}

func (s Selection) IsComposite() bool {
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

	prioritisedIsNotComposite := !s.IsComposite()

	return prioritisedIsNotComposite
}

func (s Selections) modifySelections() Selections {
	modifiedSelections := Selections{}
	for _, selection := range s {
		modifiedSelections = append(modifiedSelections, selection.modifySelection()...)
	}

	return modifiedSelections
}

func (s Selection) modifySelection() Selections {
	if s.action == REMOVE {
		removeSelection := NewSelectionBuilder(s.id).
			WithAction(REMOVE).
			Build()

		return Selections{removeSelection}
	}

	if s.IsComposite() {
		primaryPrimitiveSelection := NewSelectionBuilder(s.id).
			WithAction(s.action).
			Build()
		return Selections{primaryPrimitiveSelection, s}
	}

	return Selections{s}
}
