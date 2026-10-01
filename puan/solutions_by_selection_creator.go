package puan

import "github.com/go-errors/errors"

type SolutionsBySelectionCreator struct {
	SolverClient

	queryCreator          *solverQueryCreator
	singleSolutionCreator *singleSolutionCreator
}

func newSolutionsBySelectionCreator(
	solverClient SolverClient,
	queryCreator *solverQueryCreator,
	singleSolutionCreator *singleSolutionCreator,
) *SolutionsBySelectionCreator {
	return &SolutionsBySelectionCreator{
		SolverClient:          solverClient,
		queryCreator:          queryCreator,
		singleSolutionCreator: singleSolutionCreator,
	}
}

func (c *SolutionsBySelectionCreator) calculateSolutionsBySelection(
	query SolutionQuery,
) ([]SolutionBySelection, error) {
	dependantSelections, independentSelections :=
		query.ruleset.CategorizeSelections(query.selections)

	dependentQuery, err := NewSolutionQueryBuilder().
		fromQuery(query).
		WithSelections(dependantSelections).
		Build()
	if err != nil {
		return nil, err
	}

	dependentSolutions, err := c.calculateDependentSolutionsBySelection(dependentQuery)
	if err != nil {
		return nil, err
	}

	independentQuery, err := NewSolutionQueryBuilder().
		fromQuery(query).
		WithSelections(independentSelections).
		Build()
	if err != nil {
		return nil, err
	}

	independentSolutions, err := c.calculateIndependentSolutionsBySelection(independentQuery)
	if err != nil {
		return nil, err
	}

	var solutions []SolutionBySelection
	solutions = append(solutions, dependentSolutions...)
	solutions = append(solutions, independentSolutions...)

	return solutions, nil
}

func (c *SolutionsBySelectionCreator) calculateDependentSolutionsBySelection(
	query SolutionQuery,
) ([]SolutionBySelection, error) {
	solverQuery, err := c.queryCreator.newSolutionsBySelectionQuery(query)
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

	return newSolutionsBySelection(
		primitiveSolutions,
		query.selections,
	)
}

func (c *SolutionsBySelectionCreator) calculateIndependentSolutionsBySelection(
	query SolutionQuery,
) ([]SolutionBySelection, error) {
	defaultQuery, err := NewSolutionQueryBuilder().
		fromQuery(query).
		WithSelections(nil).
		Build()
	if err != nil {
		return nil, err
	}

	defaultSolution, err := c.singleSolutionCreator.create(defaultQuery)
	if err != nil {
		return nil, err
	}

	solutions := c.calculateManyIndependentSolutions(
		defaultSolution,
		query.selections,
	)

	return solutions, nil
}

func (c *SolutionsBySelectionCreator) calculateManyIndependentSolutions(
	solution Solution,
	selections Selections,
) []SolutionBySelection {
	solutions := make([]SolutionBySelection, len(selections))
	for i, selection := range selections {
		solution := solution.copy()
		solution[selection.id] = selection.action.asInt()
		solutions[i] = SolutionBySelection{
			selection: selection,
			solution:  solution,
		}
	}
	return solutions
}
