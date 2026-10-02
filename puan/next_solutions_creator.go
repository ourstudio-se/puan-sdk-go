package puan

type nextSolutionsCreator struct {
	SolverClient

	queryCreator          *solverQueryCreator
	singleSolutionCreator *singleSolutionCreator
}

func newNextSolutionsCreator(
	client SolverClient,
	queryCreator *solverQueryCreator,
	singleSolutionCreator *singleSolutionCreator,
) *nextSolutionsCreator {
	return &nextSolutionsCreator{
		SolverClient:          client,
		queryCreator:          queryCreator,
		singleSolutionCreator: singleSolutionCreator,
	}
}

func (c *nextSolutionsCreator) create(
	query NextSolutionsQuery,
) ([]SolutionBySelection, error) {
	solutionsForDependentSelections, err := c.createForDependentSelections(query)
	if err != nil {
		return nil, err
	}

	solutionsForIndependentSelections, err := c.createForIndependentSelections(query)
	if err != nil {
		return nil, err
	}

	var solutions []SolutionBySelection
	solutions = append(solutions, solutionsForDependentSelections...)
	solutions = append(solutions, solutionsForIndependentSelections...)

	return solutions, nil
}

func (c *nextSolutionsCreator) createForDependentSelections(
	query NextSolutionsQuery,
) ([]SolutionBySelection, error) {
	currentDependentSelections, currentIndependentSelections :=
		query.ruleset.CategorizeSelections(query.currentSelections)

	nextDependentSelections, _ := query.ruleset.CategorizeSelections(query.nextSelections)

	dependentQuery, err := NewNextSolutionsQuery(
		currentDependentSelections,
		nextDependentSelections,
		query.ruleset,
		query.from,
		query.to,
	)
	if err != nil {
		return nil, err
	}

	nextDependentSolutions, err := c.calculateDependentSolutions(dependentQuery)
	if err != nil {
		return nil, err
	}

	currentIndependentSolution := query.ruleset.calculateIndependentSolution(
		currentIndependentSelections,
	)

	solutionBySelection := make([]SolutionBySelection, len(nextDependentSolutions))
	for i, nextSolution := range nextDependentSolutions {
		mergedSolution := nextSolution.solution.merge(currentIndependentSolution)
		solutionBySelection[i] = SolutionBySelection{
			selection: nextSolution.selection,
			solution:  mergedSolution,
		}
	}

	return solutionBySelection, nil
}

func (c *nextSolutionsCreator) calculateDependentSolutions(
	query NextSolutionsQuery,
) ([]SolutionBySelection, error) {
	batchable, nonBatchable, err := query.splitByBatchability()
	if err != nil {
		return nil, err
	}

	batchedSolutions, err := c.calculateBatchableSolutions(batchable)
	if err != nil {
		return nil, err
	}

	nonBatchableSolutions, err := c.calculateNonBatchableSolutions(nonBatchable)
	if err != nil {
		return nil, err
	}

	var solutions []SolutionBySelection
	solutions = append(solutions, batchedSolutions...)
	solutions = append(solutions, nonBatchableSolutions...)

	return solutions, nil
}

func (c *nextSolutionsCreator) calculateBatchableSolutions(
	query NextSolutionsQuery,
) ([]SolutionBySelection, error) {
	// This check ensures that no extra solving with empty selections is performed.
	// Next selection can be empty if it splits all to 'nonBatchable',
	if query.hasEmptyNextSelections() {
		return nil, nil
	}

	solverQuery, err := c.queryCreator.newNextSolutionsQuery(query)
	if err != nil {
		return nil, err
	}

	solutions, err := c.SolveWithManyWeights(solverQuery)
	if err != nil {
		return nil, err
	}

	primitiveSolutions := query.ruleset.RemoveSupportVariablesForMany(solutions)

	return newSolutionsBySelection(primitiveSolutions, query.nextSelections)
}

func (c *nextSolutionsCreator) calculateNonBatchableSolutions(
	query NextSolutionsQuery,
) ([]SolutionBySelection, error) {
	solutionQueries, err := query.asSolutionQueries()
	if err != nil {
		return nil, err
	}

	solutions := make([]Solution, len(query.nextSelections))
	for i, solutionQuery := range solutionQueries {
		solution, err := c.singleSolutionCreator.create(solutionQuery)
		if err != nil {
			return nil, err
		}

		solutions[i] = solution
	}

	return newSolutionsBySelection(solutions, query.nextSelections)
}

func (c *nextSolutionsCreator) createForIndependentSelections(
	query NextSolutionsQuery,
) ([]SolutionBySelection, error) {
	currentQuery, err := NewSolutionQueryBuilder().
		WithRuleset(query.ruleset).
		WithFrom(query.from).
		WithTo(query.to).
		WithSelections(query.currentSelections).
		Build()
	if err != nil {
		return nil, err
	}

	currentSolution, err := c.singleSolutionCreator.create(currentQuery)
	if err != nil {
		return nil, err
	}

	_, nextIndependentSelections := query.ruleset.CategorizeSelections(
		query.nextSelections,
	)

	nextSolutions := c.calculateIndependentSolutionsFromCurrent(
		currentSolution,
		nextIndependentSelections,
	)

	return nextSolutions, nil
}

func (c *nextSolutionsCreator) calculateIndependentSolutionsFromCurrent(
	currentSolution Solution,
	selections Selections,
) []SolutionBySelection {
	solutions := make([]SolutionBySelection, len(selections))
	for i, selection := range selections {
		solution := currentSolution.copy()
		solution[selection.id] = selection.action.asInt()
		solutions[i] = SolutionBySelection{
			selection: selection,
			solution:  solution,
		}
	}
	return solutions
}
