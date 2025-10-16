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

func (s Solution) merge(other Solution) Solution {
	for variable, value := range other {
		s[variable] = value
	}

	return s
}
