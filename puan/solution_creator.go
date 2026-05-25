package puan

import (
	"time"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
	"github.com/ourstudio-se/puan-sdk-go/puanerror"
)

type SolverClient interface {
	Solve(query *Query) (Solution, error)
	SolveWithManyWeights(query *MultiWeightQuery) ([]Solution, error)
}

type SolutionCreator struct {
	SolverClient
}

func NewSolutionCreator(
	client SolverClient,
) *SolutionCreator {
	return &SolutionCreator{
		SolverClient: client,
	}
}

func (c *SolutionCreator) Create(
	selections Selections,
	ruleset Ruleset,
	from *time.Time,
) (SolutionEnvelope, error) {
	err := validateSelections(selections, ruleset)
	if err != nil {
		return SolutionEnvelope{}, err
	}

	dependentSelections, independentSelections :=
		categorizeSelections(selections, ruleset.independentVariables)

	dependentSolution, err := c.calculateDependentSolution(
		dependentSelections,
		ruleset,
		from,
	)
	if err != nil {
		err = updateSolveError(err, ruleset, from)
		return SolutionEnvelope{}, err
	}

	independentSolution := calculateIndependentSolution(
		ruleset.independentVariables,
		independentSelections,
	)

	solution := dependentSolution.merge(independentSolution)

	return SolutionEnvelope{
		solution: solution,
	}, nil
}

func (c *SolutionCreator) calculateDependentSolution(
	selections Selections,
	ruleset Ruleset,
	from *time.Time,
) (Solution, error) {
	query, err := newQuery(selections, ruleset, from)
	if err != nil {
		return Solution{}, err
	}

	tooLarge := query.weights.WeightsTooLarge()

	if tooLarge {
		return c.calculateSplitSolveSolution(selections, ruleset, from)
	}

	solution, err := c.Solve(query)
	if err != nil {
		return Solution{}, err
	}

	primitiveSolution := ruleset.RemoveSupportVariables(solution)

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
func (c *SolutionCreator) calculateSplitSolveSolution(
	selections Selections,
	ruleset Ruleset,
	from *time.Time,
) (Solution, error) {
	if len(selections) < 2 {
		return Solution{},
			errors.New("at least 2 selections are required for multi-solve")
	}

	remainingSelections, prioritisedSelections := selections.split()

	prioritisedSolution, err := c.calculateDependentSolution(prioritisedSelections, ruleset, from)
	if err != nil {
		return Solution{}, err
	}

	rulesetWithPrioritisedSolution, err := c.newRulesetWithAssumedSolution(
		ruleset,
		prioritisedSelections,
		prioritisedSolution,
	)
	if err != nil {
		return Solution{}, err
	}

	return c.calculateDependentSolution(remainingSelections, rulesetWithPrioritisedSolution, from)
}

func (c *SolutionCreator) newRulesetWithAssumedSolution(
	ruleset Ruleset,
	selections Selections,
	solution Solution,
) (Ruleset, error) {
	newRuleset := ruleset.copy()

	for _, selection := range selections {
		isSelected := solution.isSelected(selection.id)
		if isSelected {
			err := newRuleset.assume(selection.id)
			if err != nil {
				return Ruleset{}, err
			}
		} else {
			err := newRuleset.assumeNot(selection.id)
			if err != nil {
				return Ruleset{}, err
			}
		}
	}

	return newRuleset, nil
}

func calculateIndependentSolution(independentVariables []string, selections Selections) Solution {
	solution := make(Solution, len(independentVariables))
	for _, variable := range independentVariables {
		solution[variable] = independentSolutionValue(variable, selections)
	}

	return solution
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

func validateSelections(selections Selections, ruleset Ruleset) error {
	for _, selection := range selections {
		if !utils.ContainsAll(ruleset.selectableVariables, selection.IDs()) {
			return errors.Errorf(
				"%w: selection contains non-selectable variables: %v",
				puanerror.InvalidArgument,
				selection,
			)
		}

		hasSubSelection := len(selection.subSelectionIDs) > 0
		if hasSubSelection {
			if utils.ContainsAny(selection.IDs(), ruleset.independentVariables) {
				return errors.Errorf(
					"%w: independent variables cannot be part of a composite selections: %v",
					puanerror.InvalidArgument,
					selection,
				)
			}
		}
	}

	return nil
}

func categorizeSelections(
	selections Selections,
	independentVariables []string,
) (Selections, Selections) {
	var dependantSelections Selections
	var independentSelections Selections

	for _, selection := range selections {
		isIndependent := utils.Contains(independentVariables, selection.id)
		if isIndependent {
			independentSelections = append(independentSelections, selection)
		} else {
			dependantSelections = append(dependantSelections, selection)
		}
	}

	return dependantSelections, independentSelections
}

func newQuery(selections Selections, ruleset Ruleset, from *time.Time) (*Query, error) {
	preparedSelections := selections.prepareForMultiSelectionQuery()

	preparedRuleset, err := ruleset.prepareForQuery(preparedSelections, from)
	if err != nil {
		return nil, err
	}

	weights, err := newWeights(preparedRuleset, preparedSelections)
	if err != nil {
		return nil, err
	}

	query := NewQuery(
		preparedRuleset.polyhedron,
		preparedRuleset.dependentVariables,
		weights,
	)

	return query, nil
}

func newWeights(
	ruleset Ruleset,
	selections Selections,
) (weights.Weights, error) {
	dependentSelectableVariables := utils.Without(
		ruleset.selectableVariables,
		ruleset.independentVariables,
	)

	weightSelections, err := ruleset.newWeightSelections(selections)
	if err != nil {
		return nil, err
	}

	weights, err := weights.Calculate(
		dependentSelectableVariables,
		weightSelections,
		ruleset.preferredVariables,
		ruleset.periodVariables.ids(),
	)
	if err != nil {
		return nil, err
	}

	return weights, nil
}

func updateSolveError(
	err error,
	ruleset Ruleset,
	from *time.Time,
) error {
	solverFailed := errors.Is(err, puanerror.SolverFailed)
	if solverFailed {
		invalidTime := !ruleset.isValidFromTime(from)
		if invalidTime {
			return errors.Errorf(
				"%w: from '%s' is not valid for the ruleset",
				puanerror.InvalidArgument,
				from,
			)
		}
	}

	return err
}

func (c *SolutionCreator) CreateSolutionsBySelection(
	selections Selections,
	ruleset Ruleset,
	from *time.Time,
) (SolutionsBySelectionEnvelope, error) {
	err := validateSelections(selections, ruleset)
	if err != nil {
		return SolutionsBySelectionEnvelope{}, err
	}

	dependantSelections, independentSelections :=
		categorizeSelections(selections, ruleset.independentVariables)

	dependentSolutions, err := c.calculateDependentSolutionsBySelection(
		dependantSelections,
		ruleset,
		from,
	)
	if err != nil {
		err = updateSolveError(err, ruleset, from)
		return SolutionsBySelectionEnvelope{}, err
	}

	independentSolutions, err := c.calculateIndependentSolutionsBySelection(
		independentSelections,
		ruleset,
		from,
	)
	if err != nil {
		err = updateSolveError(err, ruleset, from)
		return SolutionsBySelectionEnvelope{}, err
	}

	var solutions []SolutionBySelection
	solutions = append(solutions, dependentSolutions...)
	solutions = append(solutions, independentSolutions...)

	return SolutionsBySelectionEnvelope{
		solutions: solutions,
	}, nil
}

func (c *SolutionCreator) calculateDependentSolutionsBySelection(
	selections Selections,
	ruleset Ruleset,
	from *time.Time,
) ([]SolutionBySelection, error) {
	query, err := newMultiWeightQuery(selections, ruleset, from)
	if err != nil {
		return nil, err
	}

	solutions, err := c.SolveWithManyWeights(query)
	if err != nil {
		return nil, err
	}

	solutionsBySelection := make([]SolutionBySelection, len(solutions))
	for i, solution := range solutions {
		primitiveSolution := ruleset.RemoveSupportVariables(solution)
		solutionsBySelection[i] = SolutionBySelection{
			selection: selections[i],
			solution:  primitiveSolution,
		}
	}

	return solutionsBySelection, nil
}

func newMultiWeightQuery(
	selections Selections,
	ruleset Ruleset,
	from *time.Time,
) (*MultiWeightQuery, error) {
	preparedRuleset, err := ruleset.prepareForQuery(selections, from)
	if err != nil {
		return nil, err
	}

	weightGroups := make([]weights.Weights, len(selections))
	for i, selection := range selections {
		weights, err := newWeights(preparedRuleset, Selections{selection})
		if err != nil {
			return nil, err
		}
		weightGroups[i] = weights
	}

	query := NewMultiWeightQuery(
		preparedRuleset.polyhedron,
		preparedRuleset.dependentVariables,
		weightGroups,
	)

	return query, nil
}

func (c *SolutionCreator) calculateIndependentSolutionsBySelection(
	selections Selections,
	ruleset Ruleset,
	from *time.Time,
) ([]SolutionBySelection, error) {
	defaultSolution, err := c.calculateDependentSolution(
		nil,
		ruleset,
		from,
	)
	if err != nil {
		return nil, err
	}

	solutionsBySelection := make([]SolutionBySelection, len(selections))
	for i, selection := range selections {
		solution := defaultSolution.withSelection(selection.id)
		solutionsBySelection[i] = SolutionBySelection{
			selection: selection,
			solution:  solution,
		}
	}

	return solutionsBySelection, nil
}
