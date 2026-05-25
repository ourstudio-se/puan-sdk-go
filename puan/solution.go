package puan

import "maps"

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

func (s Solution) withSelection(variableID string) Solution {
	newSolution := s.copy()
	newSolution.merge(Solution{variableID: 1})

	return newSolution
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
	solutions []SolutionBySelection
}

func (e SolutionsBySelectionEnvelope) Solutions() []SolutionBySelection {
	return e.solutions
}

type SolutionBySelection struct {
	selection Selection
	solution  Solution
}

func (s SolutionBySelection) Selection() Selection {
	return s.selection
}

func (s SolutionBySelection) Solution() Solution {
	return s.solution
}
