package puan

type SolutionExtended struct {
	Solution        Solution
	WeightSaturated bool
}
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
