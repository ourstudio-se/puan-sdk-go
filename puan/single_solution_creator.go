package puan

import "github.com/go-errors/errors"

type singleSolutionCreator struct {
	SolverClient
}

func newSingleSolutionCreator(
	client SolverClient,
) *singleSolutionCreator {
	return &singleSolutionCreator{
		SolverClient: client,
	}
}

func (c *singleSolutionCreator) create(
	query SolutionQuery,
) (Solution, error) {
	dependentSelections, independentSelections :=
		query.ruleset.CategorizeSelections(query.selections)

	dependentQuery, err := NewSolutionQueryBuilder().
		fromQuery(query).
		WithSelections(dependentSelections).
		Build()
	if err != nil {
		return Solution{}, err
	}

	dependentSolution, err := c.calculateDependentSolution(dependentQuery)
	if err != nil {
		return Solution{}, err
	}

	independentSolution := query.ruleset.calculateIndependentSolution(independentSelections)

	solution := dependentSolution.merge(independentSolution)

	return solution, nil
}

func (c *singleSolutionCreator) calculateDependentSolution(
	query SolutionQuery,
) (Solution, error) {
	solverQuery, err := query.asSolverQuery()
	if err != nil {
		return Solution{}, err
	}

	tooLarge := solverQuery.weights.AboveSaturationLimit()
	if tooLarge {
		return c.calculateSplitDependentSolution(query)
	}

	solution, err := c.Solve(solverQuery)
	if err != nil {
		return Solution{}, err
	}

	primitiveSolution := query.ruleset.RemoveSupportVariables(solution)

	return primitiveSolution, nil
}

// When weights are very large, we need to solve many times sequentially
//
// 1. Split selections into prioritised and remaining
// 2. Solve with prioritised selections
// 3. Create new ruleset, assuming the prioritised solution
// 4. Solve with remaining selections using the new ruleset
//
// this can happen many times recursively until all selections are solved
func (c *singleSolutionCreator) calculateSplitDependentSolution(
	query SolutionQuery,
) (Solution, error) {
	if len(query.selections) < 2 {
		return Solution{},
			errors.New("at least 2 selections are required for split solving")
	}

	remainingSelections, prioritisedSelections := query.selections.split()

	prioritisedQuery, err := NewSolutionQueryBuilder().
		fromQuery(query).
		WithSelections(prioritisedSelections).
		Build()
	if err != nil {
		return Solution{}, err
	}

	prioritisedSolution, err := c.calculateDependentSolution(prioritisedQuery)
	if err != nil {
		return Solution{}, err
	}

	rulesetWithPrioritisedSolution, err := c.newRulesetWithAssumedSolution(
		query.ruleset,
		prioritisedSelections,
		prioritisedSolution,
	)
	if err != nil {
		return Solution{}, err
	}

	remainingQuery, err := NewSolutionQueryBuilder().
		fromQuery(query).
		WithSelections(remainingSelections).
		WithRuleset(rulesetWithPrioritisedSolution).
		Build()
	if err != nil {
		return Solution{}, err
	}

	return c.calculateDependentSolution(remainingQuery)
}

func (c *singleSolutionCreator) newRulesetWithAssumedSolution(
	ruleset Ruleset,
	selections Selections,
	solution Solution,
) (Ruleset, error) {
	idsToAssume, idsToAssumeNot := c.getIDsToAssume(selections, solution)

	newRuleset := ruleset.copy()
	for _, id := range idsToAssume {
		err := newRuleset.assume(id)
		if err != nil {
			return Ruleset{}, err
		}
	}
	for _, id := range idsToAssumeNot {
		err := newRuleset.assumeNot(id)
		if err != nil {
			return Ruleset{}, err
		}
	}

	return newRuleset, nil
}

// nolint:gocyclo
func (c *singleSolutionCreator) getIDsToAssume(
	selections Selections,
	solution Solution,
) ([]string, []string) {
	assumeByID := make(map[string]bool)
	for _, selection := range selections {
		isSelected := solution.isSelected(selection.id)
		if isSelected {
			assumeByID[selection.id] = true

			for _, subSelection := range selection.subSelectionIDs {
				isSubSelected := solution.isSelected(subSelection)
				if isSubSelected {
					assumeByID[subSelection] = true
				}
			}
		} else {
			assumeByID[selection.id] = false
		}
	}

	var idsToAssume []string
	var idsToAssumeNot []string
	for id, shouldAssume := range assumeByID {
		if shouldAssume {
			idsToAssume = append(idsToAssume, id)
		} else {
			idsToAssumeNot = append(idsToAssumeNot, id)
		}
	}

	return idsToAssume, idsToAssumeNot
}
