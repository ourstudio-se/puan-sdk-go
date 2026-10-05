package puan

import (
	"maps"

	"github.com/go-errors/errors"
)

// Map of variable IDs and 0 or 1, representing whether the variable is selected or not
type Solution map[string]int

func (s Solution) Extract(variables ...string) Solution {
	extracted := make(Solution)
	for _, variable := range variables {
		if _, ok := s[variable]; ok {
			extracted[variable] = s[variable]
		}
	}

	return extracted
}

func (s Solution) merge(other Solution) Solution {
	maps.Copy(s, other)

	return s
}

func (s Solution) copy() Solution {
	copied := make(Solution)
	maps.Copy(copied, s)
	return copied
}

func (s Solution) isSelected(variableID string) bool {
	return s[variableID] == 1
}

type SolutionEnvelope struct {
	solution Solution
}

func (e SolutionEnvelope) Solution() Solution {
	return e.solution
}

type SolutionsBySelectionEnvelope struct {
	solutionsBySelection map[string]SolutionBySelection
}

func NewSolutionsBySelectionEnvelope(
	solutions []SolutionBySelection,
) (SolutionsBySelectionEnvelope, error) {
	solutionsBySelection := make(map[string]SolutionBySelection)
	for _, solution := range solutions {
		selectionHash := solution.selection.Hash()
		if _, exists := solutionsBySelection[selectionHash]; exists {
			return SolutionsBySelectionEnvelope{}, errors.Errorf(
				"duplicate solution for selection: %v",
				solution.selection,
			)
		}
		solutionsBySelection[selectionHash] = solution
	}

	return SolutionsBySelectionEnvelope{
		solutionsBySelection: solutionsBySelection,
	}, nil
}

func (e SolutionsBySelectionEnvelope) SolutionsBySelection() map[string]SolutionBySelection {
	return e.solutionsBySelection
}

func (e SolutionsBySelectionEnvelope) GetSolutionBySelection(
	selection Selection,
) (SolutionBySelection, error) {
	selectionHash := selection.Hash()
	solution, ok := e.solutionsBySelection[selectionHash]
	if !ok {
		return SolutionBySelection{}, errors.Errorf(
			"solution not found for selection: %v",
			selection,
		)
	}
	return solution, nil
}

type SolutionBySelection struct {
	selection Selection
	solution  Solution
}

func newSolutionsBySelection(
	solutions []Solution,
	selections Selections,
) ([]SolutionBySelection, error) {
	if len(solutions) != len(selections) {
		return nil, errors.Errorf(
			"Expected amount of solutions and selections to match. Got %d and %d",
			len(solutions),
			len(selections),
		)
	}

	solutionsBySelection := make([]SolutionBySelection, len(solutions))
	for i, solution := range solutions {
		selection := selections[i]

		solutionBySelection := SolutionBySelection{
			selection: selection,
			solution:  solution,
		}
		solutionsBySelection[i] = solutionBySelection
	}

	return solutionsBySelection, nil
}

func (s SolutionBySelection) Selection() Selection {
	return s.selection
}

func (s SolutionBySelection) Solution() Solution {
	return s.solution
}

type SolutionForSelectionGroup struct {
	selections Selections
	solution   Solution
}

func newSolutionsBySelectionGroup(
	solutions []Solution,
	selectionGroups []Selections,
) ([]SolutionForSelectionGroup, error) {
	if len(solutions) != len(selectionGroups) {
		return nil, errors.Errorf(
			"Expected amount of solutions and selection groups to match. Got %d and %d",
			len(solutions),
			len(selectionGroups),
		)
	}

	solutionsBySelectionGroup := make([]SolutionForSelectionGroup, len(selectionGroups))
	for i, selectionGroup := range selectionGroups {
		solutionsBySelectionGroup[i] = SolutionForSelectionGroup{
			selections: selectionGroup,
			solution:   solutions[i],
		}
	}

	return solutionsBySelectionGroup, nil
}

func (s SolutionForSelectionGroup) Selections() Selections {
	return s.selections
}

func (s SolutionForSelectionGroup) Solution() Solution {
	return s.solution
}

type SolutionsBySelectionGroupEnvelope struct {
	solutionsBySelectionGroup []SolutionForSelectionGroup
}

func NewSolutionsBySelectionGroupEnvelope(
	solutions []SolutionForSelectionGroup,
) SolutionsBySelectionGroupEnvelope {
	return SolutionsBySelectionGroupEnvelope{
		solutionsBySelectionGroup: solutions,
	}
}

func (e SolutionsBySelectionGroupEnvelope) SolutionsBySelectionGroup() []SolutionForSelectionGroup {
	return e.solutionsBySelectionGroup
}
