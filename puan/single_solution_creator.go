package puan

import "github.com/go-errors/errors"

type singleSolutionCreator struct {
	SolverClient
	queryCreator *solverQueryCreator
}

func newSingleSolutionCreator(
	client SolverClient,
	queryCreator *solverQueryCreator,
) *singleSolutionCreator {
	return &singleSolutionCreator{
		SolverClient: client,
		queryCreator: queryCreator,
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
	solverQuery, err := c.queryCreator.new(query)
	if err != nil {
		return Solution{}, err
	}

	tooLarge := solverQuery.weights.WeightsTooLarge()
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

// nolint:gocyclo
func (c *singleSolutionCreator) newRulesetWithAssumedSolution(
	ruleset Ruleset,
	selections Selections,
	solution Solution,
) (Ruleset, error) {
	newRuleset := ruleset.copy()
	for _, selection := range selections {
		isNotSelected := !solution.isSelected(selection.id)
		if isNotSelected {
			// Large selection sets are split and solved by priority, locking
			// higher-priority solution first. Unselected selections in a split must be explicitly
			// assumeNot, or unwanted behavior can occur, e.g. an ADD and REMOVE ending
			// up in different splits could let the lower-priority ADD take effect.
			// Note: subSelectionIDs are excluded from assumeNot here, since they may
			// appear in other selections/subselections.
			err := newRuleset.assumeNot(selection.id)
			if err != nil {
				return Ruleset{}, err
			}

			continue
		}

		err := newRuleset.assume(selection.id)
		if err != nil {
			return Ruleset{}, err
		}

		for _, subSelection := range selection.subSelectionIDs {
			isSubSelected := solution.isSelected(subSelection)
			if isSubSelected {
				err = newRuleset.assume(subSelection)
				if err != nil {
					return Ruleset{}, err
				}
			}
		}
	}

	return newRuleset, nil
}
