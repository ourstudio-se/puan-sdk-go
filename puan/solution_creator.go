package puan

import (
	"time"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
	"github.com/ourstudio-se/puan-sdk-go/puanerror"
)

type SolverClient interface {
	Solve(query *SolverQuery) (Solution, error)
	SolveWithManyWeights(query *MultiWeightSolverQuery) ([]Solution, error)
}

type SolutionCreator struct {
	SolverClient
	queryCreator *solverQueryCreator
}

func NewSolutionCreator(
	client SolverClient,
) *SolutionCreator {
	queryCreator := newSolverQueryCreator()
	return &SolutionCreator{
		SolverClient: client,
		queryCreator: queryCreator,
	}
}

func (c *SolutionCreator) Create(
	query SolutionQuery,
) (SolutionEnvelope, error) {
	err := query.validate()
	if err != nil {
		return SolutionEnvelope{}, err
	}

	solution, err := c.calculateSolution(query)
	if err != nil {
		err = updateSolveError(err, query.ruleset, query.from)
		return SolutionEnvelope{}, err
	}

	return SolutionEnvelope{
		solution: solution,
	}, nil
}

func (c *SolutionCreator) calculateSolution(
	query SolutionQuery,
) (Solution, error) {
	dependentSelections, independentSelections :=
		categorizeSelections(query.selections, query.ruleset.independentVariables)

	dependentQuery := NewSolutionQueryBuilder().
		fromQuery(query).
		WithSelections(dependentSelections).
		Build()
	dependentSolution, err := c.calculateDependentSolution(dependentQuery)
	if err != nil {
		return Solution{}, err
	}

	independentSolution := calculateIndependentSolution(
		query.ruleset.independentVariables,
		independentSelections,
	)

	solution := dependentSolution.merge(independentSolution)

	return solution, nil
}

func (c *SolutionCreator) calculateDependentSolution(
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
func (c *SolutionCreator) calculateSplitDependentSolution(
	query SolutionQuery,
) (Solution, error) {
	if len(query.selections) < 2 {
		return Solution{},
			errors.New("at least 2 selections are required for split solving")
	}

	remainingSelections, prioritisedSelections := query.selections.split()

	prioritisedQuery := NewSolutionQueryBuilder().
		fromQuery(query).
		WithSelections(prioritisedSelections).
		Build()
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

	remainingQuery := NewSolutionQueryBuilder().
		fromQuery(query).
		WithSelections(remainingSelections).
		WithRuleset(rulesetWithPrioritisedSolution).
		Build()
	return c.calculateDependentSolution(remainingQuery)
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
	query SolutionQuery,
) (SolutionsBySelectionEnvelope, error) {
	err := query.validate()
	if err != nil {
		return SolutionsBySelectionEnvelope{}, err
	}

	solutions, err := c.calculateSolutionsBySelection(query)
	if err != nil {
		err = updateSolveError(err, query.ruleset, query.from)
		return SolutionsBySelectionEnvelope{}, err
	}

	return NewSolutionsBySelectionEnvelope(solutions)
}

func (c *SolutionCreator) calculateSolutionsBySelection(
	query SolutionQuery,
) ([]SolutionBySelection, error) {
	dependantSelections, independentSelections :=
		categorizeSelections(query.selections, query.ruleset.independentVariables)

	dependentQuery := NewSolutionQueryBuilder().
		fromQuery(query).
		WithSelections(dependantSelections).
		Build()
	dependentSolutions, err := c.calculateDependentSolutionsBySelection(dependentQuery)
	if err != nil {
		return nil, err
	}

	independentQuery := NewSolutionQueryBuilder().
		fromQuery(query).
		WithSelections(independentSelections).
		Build()
	independentSolutions, err := c.calculateIndependentSolutionsBySelection(independentQuery)
	if err != nil {
		return nil, err
	}

	var solutions []SolutionBySelection
	solutions = append(solutions, dependentSolutions...)
	solutions = append(solutions, independentSolutions...)

	return solutions, nil
}

func (c *SolutionCreator) calculateDependentSolutionsBySelection(
	query SolutionQuery,
) ([]SolutionBySelection, error) {
	solverQuery, err := c.queryCreator.newSolutionsBySelectionQuery(query)
	if err != nil {
		return nil, err
	}

	solutions, err := c.SolveWithManyWeights(solverQuery)
	if err != nil {
		return nil, err
	}

	primitiveSolutions := query.ruleset.RemoveSupportVariablesForMany(solutions)

	solutionsBySelection := make([]SolutionBySelection, len(solutions))
	for i := range primitiveSolutions {
		selection := query.selections[i]
		solution := primitiveSolutions[i]
		solutionsBySelection[i] = SolutionBySelection{
			selection: selection,
			solution:  solution,
		}
	}

	return solutionsBySelection, nil
}

func (c *SolutionCreator) calculateIndependentSolutionsBySelection(
	query SolutionQuery,
) ([]SolutionBySelection, error) {
	defaultQuery := NewSolutionQueryBuilder().
		fromQuery(query).
		WithSelections(nil).
		Build()
	defaultSolution, err := c.calculateDependentSolution(defaultQuery)
	if err != nil {
		return nil, err
	}

	solutionsBySelection := make([]SolutionBySelection, len(query.selections))
	for i, selection := range query.selections {
		solution := defaultSolution.withSelection(selection.id)
		solutionsBySelection[i] = SolutionBySelection{
			selection: selection,
			solution:  solution,
		}
	}

	return solutionsBySelection, nil
}

func (c *SolutionCreator) CreateNextSolutions(
	query SolutionQuery,
) (SolutionsBySelectionEnvelope, error) {
	err := query.validate()
	if err != nil {
		return SolutionsBySelectionEnvelope{}, err
	}

	solutions, err := c.createNextSolutions(query)
	if err != nil {
		err = updateSolveError(err, query.ruleset, query.from)
		return SolutionsBySelectionEnvelope{}, err
	}

	return NewSolutionsBySelectionEnvelope(solutions)
}

func (c *SolutionCreator) createNextSolutions(
	query SolutionQuery,
) ([]SolutionBySelection, error) {
	dependentSolutions, err := c.calculateNextDependentSolutions(query)
	if err != nil {
		return nil, err
	}

	independentSolutions, err := c.calculateNextIndependentSolutions(query)
	if err != nil {
		return nil, err
	}

	var solutions []SolutionBySelection
	solutions = append(solutions, dependentSolutions...)
	solutions = append(solutions, independentSolutions...)

	return solutions, nil
}

func (c *SolutionCreator) calculateNextDependentSolutions(
	query SolutionQuery,
) ([]SolutionBySelection, error) {
	dependentSelections, independentSelections :=
		categorizeSelections(query.selections, query.ruleset.independentVariables)

	preparedRuleset, err := query.ruleset.modifyForQuery(
		dependentSelections,
		query.from,
		query.to,
	)
	if err != nil {
		return nil, err
	}

	type SelectionWithWeights struct {
		selection Selection
		weights   weights.Weights
	}

	selectableVariables := preparedRuleset.dependentSelectableVariables()
	selectionWithWeights := make([]SelectionWithWeights, len(selectableVariables))
	for i, variable := range selectableVariables {
		newSelections := make(Selections, len(dependentSelections))
		copy(newSelections, dependentSelections)

		selectionBuilder := NewSelectionBuilder(variable)
		isSelected := utils.Contains(newSelections.ids(), variable)
		if isSelected {
			selectionBuilder.WithAction(REMOVE)
		}
		selection := selectionBuilder.Build()
		newSelections = append(newSelections, selection)

		newSelections = newSelections.prepareForQuery()

		weights, err := newWeights(preparedRuleset, newSelections)
		if err != nil {
			return nil, err
		}
		selectionWithWeights[i] = SelectionWithWeights{
			selection: selection,
			weights:   weights,
		}
	}

	weightGroups := make([]weights.Weights, len(selectionWithWeights))
	for i, selectionWithWeight := range selectionWithWeights {
		weightGroups[i] = selectionWithWeight.weights
	}

	solverQuery := NewMultiWeightSolverQuery(
		preparedRuleset.polyhedron,
		preparedRuleset.dependentVariables,
		weightGroups,
	)

	solutions, err := c.SolveWithManyWeights(solverQuery)
	if err != nil {
		return nil, err
	}

	independentSolution := calculateIndependentSolution(
		query.ruleset.independentVariables,
		independentSelections,
	)

	primitiveSolutions := query.ruleset.RemoveSupportVariablesForMany(solutions)

	solutionsBySelection := make([]SolutionBySelection, len(solutions))
	for i := range primitiveSolutions {
		selectionWithWeight := selectionWithWeights[i]
		selection := selectionWithWeight.selection
		solution := primitiveSolutions[i].merge(independentSolution)
		solutionsBySelection[i] = SolutionBySelection{
			selection: selection,
			solution:  solution,
		}
	}

	return solutionsBySelection, nil
}

func (c *SolutionCreator) calculateNextIndependentSolutions(
	query SolutionQuery,
) ([]SolutionBySelection, error) {
	dependentSelections, independentSelections :=
		categorizeSelections(query.selections, query.ruleset.independentVariables)

	dependentQuery := NewSolutionQueryBuilder().
		fromQuery(query).
		WithSelections(dependentSelections).
		Build()
	defaultDependentSolution, err := c.calculateDependentSolution(dependentQuery)
	if err != nil {
		return nil, err
	}

	independentDefaultSolution := calculateIndependentSolution(
		query.ruleset.independentVariables,
		independentSelections,
	)

	defaultSolution := defaultDependentSolution.merge(independentDefaultSolution)

	independentVariables := query.ruleset.independentVariables

	solutions := make([]SolutionBySelection, len(independentVariables))
	for i, variable := range independentVariables {
		selection := Selection{id: variable}
		solution := defaultSolution.copy()

		isSelected := utils.Contains(query.selections.ids(), variable)
		if isSelected {
			solution[variable] = 0
			selection.action = REMOVE
		} else {
			solution[variable] = 1
			selection.action = ADD
		}

		solutionBySelection := SolutionBySelection{
			selection: selection,
			solution:  solution,
		}
		solutions[i] = solutionBySelection
	}

	return solutions, nil
}
