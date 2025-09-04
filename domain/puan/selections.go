package puan

import (
	"sort"
	"strings"

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

func NewSelection(action Action, id string, subSelectionID *string) Selection {
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
	reversedSelections := utils.Reverse(s)
	seen := make(map[string][]*string)
	reversedImpactingSelections := Selections{}

	for _, selection := range reversedSelections {
		if _, ok := seen[selection.id]; ok {
			if utils.Contains(seen[selection.id], selection.subSelectionID) {
				continue
			}
			if utils.Contains(seen[selection.id], nil) {
				continue
			}
		}

		seen[selection.id] = append(seen[selection.id], selection.subSelectionID)
		if selection.action == ADD {
			reversedImpactingSelections = append(reversedImpactingSelections, selection)
		}
	}

	impactingSelections := utils.Reverse(reversedImpactingSelections)

	return impactingSelections
}

func createOrderIndependentID(id string, subSelectionID *string) string {
	var sorted []string
	sorted = append(sorted, id)
	if subSelectionID != nil {
		sorted = append(sorted, *subSelectionID)
	}
	sort.Strings(sorted)

	return strings.Join(sorted, ",")
}
