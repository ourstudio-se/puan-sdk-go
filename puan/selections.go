package puan

import (
	"crypto/sha1"
	"fmt"

	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
)

const (
	ADD    Action = "ADD"
	REMOVE Action = "REMOVE"
)

type Action string

func (a Action) asInt() int {
	if a == ADD {
		return 1
	}
	return 0
}

type (
	Selection struct {
		id              string
		subSelectionIDs []string
		action          Action
	}
	Selections []Selection
)

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

func (s Selection) Action() Action {
	return s.action
}

func (s Selection) SubSelectionIDs() []string {
	return s.subSelectionIDs
}

func (s Selection) IsComposite() bool {
	return len(s.subSelectionIDs) > 0
}

func (s Selection) IDs() []string {
	ids := make([]string, len(s.subSelectionIDs)+1)
	ids[0] = s.id
	copy(ids[1:], s.subSelectionIDs)
	return ids
}

func (s Selection) Hash() string {
	h := sha1.New()
	h.Write([]byte(s.action))
	h.Write([]byte(s.id))
	for _, subID := range utils.Sorted(s.subSelectionIDs) {
		h.Write([]byte(subID))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (s Selection) Equals(other Selection) bool {
	return s.Hash() == other.Hash()
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

func (s Selections) Contains(selection Selection) bool {
	for _, s := range s {
		if s.Equals(selection) {
			return true
		}
	}
	return false
}

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

// Prepares selections for a query.
// Modifies, adds additional and cleans up redundant selections.
func (selectionsByOccurrence Selections) prepareForQuery() Selections {
	modified := selectionsByOccurrence.modifyForQuery()
	impacting := modified.getImpacting()

	return impacting
}

func (s Selections) modifyForQuery() Selections {
	modifiedSelections := Selections{}
	for _, selection := range s {
		modifiedSelections = append(
			modifiedSelections,
			selection.modifyForQuery()...,
		)
	}

	return modifiedSelections
}

func (s Selection) modifyForQuery() Selections {
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

func (selectionsByOccurrence Selections) getImpacting() Selections {
	byPriority := selectionsByOccurrence.reverse()
	impactingByPriority := byPriority.filterOutRedundant()
	impactingByOccurrence := impactingByPriority.reverse()

	return impactingByOccurrence
}

func (s Selections) reverse() Selections {
	return utils.Reverse(s)
}

func (selections Selections) copy() Selections {
	newSelections := make(Selections, len(selections))
	copy(newSelections, selections)
	return newSelections
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
