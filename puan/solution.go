package puan

import (
	"github.com/go-errors/errors"
)

type Solution map[string]int

func (s Solution) Extract(variables ...string) (Solution, error) {
	extracted := make(Solution)
	for _, variable := range variables {
		if _, ok := s[variable]; !ok {
			return Solution{}, errors.Errorf("variable %s not found in solution", variable)
		}

		extracted[variable] = s[variable]
	}

	return extracted, nil
}

func (s Solution) applyIndependentVariables(
	independentVariables []string,
	selections Selections,
) Solution {
	for _, variable := range independentVariables {
		s[variable] = independentSolutionValue(variable, selections)
	}

	return s
}

func independentSolutionValue(variableID string, selections Selections) int {
	// reverse loop for prioritizing the latest selection action
	for i := len(selections) - 1; i >= 0; i-- {
		selection := selections[i]
		if selection.id == variableID {
			if selection.action == ADD {
				return 1
			}

			return 0
		}
	}

	return 0
}
