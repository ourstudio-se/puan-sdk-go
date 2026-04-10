package solve

import (
	"testing"

	"github.com/ourstudio-se/puan-sdk-go/puan"
	"github.com/stretchr/testify/assert"
)

type solutionAsserter struct {
	puan.Solution
}

func newSolutionAsserter(solution puan.Solution) solutionAsserter {
	return solutionAsserter{solution}
}

func (s solutionAsserter) assertActive(t *testing.T, variables ...string) {
	solution := s.Extract(variables...)
	for _, variable := range variables {
		value, ok := solution[variable]
		if !ok {
			assert.Failf(t, "variable %s not found in solution", variable)
		}

		assert.Equal(t, 1, value, "expected %s to be active", variable)
	}
}

func (s solutionAsserter) assertInactive(t *testing.T, variables ...string) {
	solution := s.Extract(variables...)
	for _, variable := range variables {
		value, ok := solution[variable]
		if !ok {
			assert.Failf(t, "variable %s not found in solution", variable)
		}

		assert.Equal(t, 0, value, "expected %s to be inactive", variable)
	}
}
