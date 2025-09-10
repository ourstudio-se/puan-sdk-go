package puan

import (
	"fmt"
	"strings"

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

func (s Selection) IDs() []string {
	ids := make([]string, len(s.subSelectionIDs)+1)
	ids[0] = s.id
	copy(ids[1:], s.subSelectionIDs)
	return ids
}

func (s Selection) Key() string {
	ids := s.IDs()
	key := fmt.Sprintf("%s:%s", s.action, strings.Join(ids, ","))
	return key
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

//nolint:gocyclo
func (s Selection) makesRedundant(other Selection) bool {
	if s.action == REMOVE && utils.ContainsAll(other.IDs(), s.IDs()) {
		return true
	}

	if s.action == REMOVE && utils.ContainsAny(other.IDs(), s.subSelectionIDs) {
		return true
	}

	if s.id != other.id {
		return false
	}

	if len(s.subSelectionIDs) == 0 {
		return true
	}

	if len(other.subSelectionIDs) == 0 {
		return false
	}

	if utils.ContainsAny(other.subSelectionIDs, s.subSelectionIDs) {
		return true
	}

	return false
}
