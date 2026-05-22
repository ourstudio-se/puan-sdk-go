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
		for _, id := range selection.IDs() {
			if !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
	}

	return ids
}

// split into two contiguous slices, preserving order
func (s Selections) split() (Selections, Selections) {
	n := len(s)
	switch n {
	case 0:
		return nil, nil
	case 1:
		return s, nil
	default:
		mid := (n + 1) / 2
		return s[:mid], s[mid:]
	}
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

func (s Selection) IDs() []string {
	ids := make([]string, len(s.subSelectionIDs)+1)
	ids[0] = s.id
	copy(ids[1:], s.subSelectionIDs)
	return ids
}

func (s Selection) makesRedundant(other Selection) bool {
	if utils.ContainsAll(other.IDs(), s.IDs()) {
		return true
	}

	if s.action == REMOVE && utils.ContainsAny(other.IDs(), s.subSelectionIDs) {
		return true
	}

	if s.id != other.id {
		return false
	}

	if utils.ContainsAll(other.IDs(), s.subSelectionIDs) {
		return true
	}

	prioritisedIsNotComposite := !s.IsComposite()

	return prioritisedIsNotComposite
}

// Prepares selections for queries that solves for many selections
// at the same time.
// Modifies, adds additional and cleans up redundant selections.
func (selectionsByOccurrence Selections) prepareForMultiSelectionQuery() Selections {
	modified := selectionsByOccurrence.modifyForMultiSelectionQuery()
	impacting := modified.getImpactingForMultiSelectionQuery()

	return impacting
}

func (s Selections) modifyForMultiSelectionQuery() Selections {
	modifiedSelections := Selections{}
	for _, selection := range s {
		modifiedSelections = append(modifiedSelections, selection.modifyForMultiSelectionQuery()...)
	}

	return modifiedSelections
}

func (s Selection) modifyForMultiSelectionQuery() Selections {
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

func (selectionsByOccurance Selections) getImpactingForMultiSelectionQuery() Selections {
	byPriority := selectionsByOccurance.reverse()
	impactingByPriority := byPriority.filterOutRedundant()
	impactingByOccurance := impactingByPriority.reverse()

	return impactingByOccurance
}

func (s Selections) reverse() Selections {
	return utils.Reverse(s)
}

func (selectionsByPriority Selections) filterOutRedundant() Selections {
	var filtered Selections
	for _, selection := range selectionsByPriority {
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
