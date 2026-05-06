package puan

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
	for variable, value := range other {
		s[variable] = value
	}

	return s
}

func (s Solution) isSelected(variableID string) bool {
	return s[variableID] == 1
}

type SolutionEnvelope struct {
	solution        Solution
	weightsTooLarge bool
}

func (se SolutionEnvelope) Solution() Solution {
	return se.solution
}

func (se SolutionEnvelope) WeightsTooLarge() bool {
	return se.weightsTooLarge
}
