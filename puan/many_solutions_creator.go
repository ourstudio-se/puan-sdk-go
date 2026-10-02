package puan

import "github.com/go-errors/errors"

type manySolutionsCreator struct {
	SolverClient
}

func newManySolutionsCreator(
	client SolverClient,
) *manySolutionsCreator {
	return &manySolutionsCreator{
		SolverClient: client,
	}
}

func (c *manySolutionsCreator) create(
	query ManySolutionsQuery,
) ([]SolutionForSelectionGroup, error) {
	dependentSelectionGroups := make([]Selections, len(query.selectionGroups))
	independentSelectionGroups := make([]Selections, len(query.selectionGroups))

	for i, selectionGroup := range query.selectionGroups {
		dependentSelections, independentSelections := query.ruleset.CategorizeSelections(selectionGroup)
		dependentSelectionGroups[i] = dependentSelections
		independentSelectionGroups[i] = independentSelections
	}

	dependentQuery, err := NewManySolutionsQuery(
		dependentSelectionGroups,
		query.ruleset,
		query.from,
		query.to,
	)
	if err != nil {
		return nil, err
	}

	dependentSolutions, err := c.createDependentSolutions(dependentQuery)
	if err != nil {
		return nil, err
	}

	independentSolutions := c.createIndependentSolutions(
		query.ruleset,
		independentSelectionGroups,
	)

	solutions, err := c.mergeSolutions(dependentSolutions, independentSolutions)
	if err != nil {
		return nil, err
	}

	solutionsBySelectionGroup, err := c.groupSolutions(query.selectionGroups, solutions)
	if err != nil {
		return nil, err
	}

	return solutionsBySelectionGroup, nil
}

func (c *manySolutionsCreator) createDependentSolutions(
	query ManySolutionsQuery,
) ([]Solution, error) {
	solverQuery, err := query.asSolverQuery()
	if err != nil {
		return nil, err
	}

	err = solverQuery.validateWeightLimit()
	if err != nil {
		return nil, err
	}

	solutions, err := c.SolveWithManyWeights(solverQuery)
	if err != nil {
		return nil, err
	}

	primitiveSolutions := query.ruleset.RemoveSupportVariablesForMany(solutions)
	return primitiveSolutions, nil
}

func (c *manySolutionsCreator) createIndependentSolutions(
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

func (c *manySolutionsCreator) mergeSolutions(
	dependentSolutions []Solution,
	independentSolutions []Solution,
) ([]Solution, error) {
	if len(dependentSolutions) != len(independentSolutions) {
		return nil, errors.Errorf(
			"mismatch in length of dependent (%d) and independent (%d) solutions",
			len(dependentSolutions),
			len(independentSolutions),
		)
	}

	mergedSolutions := make([]Solution, len(dependentSolutions))
	for i, dependentSolution := range dependentSolutions {
		independentSolution := independentSolutions[i]
		mergedSolution := dependentSolution.merge(independentSolution)
		mergedSolutions[i] = mergedSolution
	}
	return mergedSolutions, nil
}

func (c *manySolutionsCreator) groupSolutions(
	selectionGroups []Selections,
	solutions []Solution,
) ([]SolutionForSelectionGroup, error) {
	if len(selectionGroups) != len(solutions) {
		return nil, errors.Errorf(
			"mismatch in length of selection groups (%d) and solutions (%d)",
			len(selectionGroups),
			len(solutions),
		)
	}

	solutionsBySelectionGroup := make([]SolutionForSelectionGroup, len(selectionGroups))
	for i, selections := range selectionGroups {
		solutionForSelectionGroup := SolutionForSelectionGroup{
			selections: selections,
			solution:   solutions[i],
		}
		solutionsBySelectionGroup[i] = solutionForSelectionGroup
	}
	return solutionsBySelectionGroup, nil
}
