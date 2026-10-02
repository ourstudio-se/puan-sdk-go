package puan

type solutionsBySelectionCreator struct {
	SolverClient

	queryCreator          *solverQueryCreator
	singleSolutionCreator *singleSolutionCreator
}

func newSolutionsBySelectionCreator(
	solverClient SolverClient,
	queryCreator *solverQueryCreator,
	singleSolutionCreator *singleSolutionCreator,
) *solutionsBySelectionCreator {
	return &solutionsBySelectionCreator{
		SolverClient:          solverClient,
		queryCreator:          queryCreator,
		singleSolutionCreator: singleSolutionCreator,
	}
}

func (c *solutionsBySelectionCreator) create(
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

	dependentSolutions, err := c.calculateDependentSolutions(dependentQuery)
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

	independentSolutions, err := c.calculateIndependentSolutions(independentQuery)
	if err != nil {
		return nil, err
	}

	var solutions []SolutionBySelection
	solutions = append(solutions, dependentSolutions...)
	solutions = append(solutions, independentSolutions...)

	return solutions, nil
}

func (c *solutionsBySelectionCreator) calculateDependentSolutions(
	query SolutionQuery,
) ([]SolutionBySelection, error) {
	solverQuery, err := c.queryCreator.newSolutionsBySelectionQuery(query)
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

	return newSolutionsBySelection(
		primitiveSolutions,
		query.selections,
	)
}

func (c *solutionsBySelectionCreator) calculateIndependentSolutions(
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

	solutions := c.calculateIndependentSolutionsFromDefault(
		defaultSolution,
		query.selections,
	)

	return solutions, nil
}

func (c *solutionsBySelectionCreator) calculateIndependentSolutionsFromDefault(
	defaultSolution Solution,
	selections Selections,
) []SolutionBySelection {
	solutions := make([]SolutionBySelection, len(selections))
	for i, selection := range selections {
		solution := defaultSolution.copy()
		solution[selection.id] = selection.action.asInt()
		solutions[i] = SolutionBySelection{
			selection: selection,
			solution:  solution,
		}
	}
	return solutions
}
