package puan

import "github.com/go-errors/errors"

type manySolutionsCreator struct {
	SolverClient

	queryCreator *solverQueryCreator
}

func newManySolutionsCreator(
	client SolverClient,
	queryCreator *solverQueryCreator,
) *manySolutionsCreator {
	return &manySolutionsCreator{
		SolverClient: client,
		queryCreator: queryCreator,
	}
}

func (c *manySolutionsCreator) createManySolutions(
	query ManySolutionQueries,
) ([]SolutionForSelectionGroup, error) {
	dependentSelectionGroups := make([]Selections, len(query.selectionGroups))
	indipendentSelectionGroups := make([]Selections, len(query.selectionGroups))

	for i, selectionGroup := range query.selectionGroups {
		dependentSelections, independentSelections := query.ruleset.CategorizeSelections(selectionGroup)
		dependentSelectionGroups[i] = dependentSelections
		indipendentSelectionGroups[i] = independentSelections
	}

	dependentQuery := NewManySolutionQueries(
		dependentSelectionGroups,
		query.ruleset,
		query.from,
		query.to,
	)

	dependentSolutions, err := c.createManyDependentSolutions(dependentQuery)
	if err != nil {
		return nil, err
	}

	independentSolutions := c.createManyIndependentSolutions(
		query.ruleset,
		indipendentSelectionGroups,
	)

	solutionsBySelectionGroup := make([]SolutionForSelectionGroup, len(query.selectionGroups))
	for i, selections := range query.selectionGroups {
		dependentSolution := dependentSolutions[i]
		independentSolution := independentSolutions[i]
		mergedSolution := dependentSolution.merge(independentSolution)

		solutionForSelectionGroup := SolutionForSelectionGroup{
			selections: selections,
			solution:   mergedSolution,
		}
		solutionsBySelectionGroup[i] = solutionForSelectionGroup
	}

	return solutionsBySelectionGroup, nil
}

func (c *manySolutionsCreator) createManyDependentSolutions(
	query ManySolutionQueries,
) ([]Solution, error) {
	solverQuery, err := c.queryCreator.newManySolutionsQuery(query)
	if err != nil {
		return nil, err
	}

	for i, weights := range solverQuery.WeightGroups() {
		tooLarge := weights.WeightsTooLarge()
		if tooLarge {
			return nil, errors.Errorf("weights are too large at index %d", i)
		}
	}

	solutions, err := c.SolveWithManyWeights(solverQuery)
	if err != nil {
		return nil, err
	}

	primitiveSolutions := query.ruleset.RemoveSupportVariablesForMany(solutions)
	return primitiveSolutions, nil
}

func (c *manySolutionsCreator) createManyIndependentSolutions(
	ruleset Ruleset,
	selectionGroups []Selections,
) []Solution {
	solutions := make([]Solution, len(selectionGroups))
	for i, selections := range selectionGroups {
		independentSolution := ruleset.calculateIndependentSolution(selections)
		solutions[i] = independentSolution
	}
	return solutions
}
