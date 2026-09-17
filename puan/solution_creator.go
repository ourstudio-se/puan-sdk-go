package puan

import (
	"time"

	"github.com/go-errors/errors"
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

func (c *SolutionCreator) calculateDependentSolutionsBySelection(
	query SolutionQuery,
) ([]SolutionBySelection, error) {
	solverQuery, err := c.queryCreator.newSolutionsBySelectionQuery(query)
	if err != nil {
		return nil, err
	}

	for _, weights := range solverQuery.WeightGroups() {
		tooLarge := weights.WeightsTooLarge()
		if tooLarge {
			return nil, errors.Errorf(
				"%w: weights are too large to solve by selection",
				puanerror.InvalidArgument,
			)
		}
	}

	solutions, err := c.SolveWithManyWeights(solverQuery)
	if err != nil {
		return nil, err
	}

	primitiveSolutions := query.ruleset.RemoveSupportVariablesForMany(solutions)

	solutionsBySelection, err := c.groupSolutionsBySelection(
		primitiveSolutions,
		query.selections,
	)
	if err != nil {
		return nil, err
	}

	return solutionsBySelection, nil
}

func (c *SolutionCreator) groupSolutionsBySelection(
	solutions []Solution,
	selections Selections,
) ([]SolutionBySelection, error) {
	if len(solutions) != len(selections) {
		return nil, errors.Errorf(
			"Expected amount of solutions and selections to match. Got %d and %d",
			len(solutions),
			len(selections),
		)
	}

	solutionsBySelection := make([]SolutionBySelection, len(solutions))
	for i, solution := range solutions {
		selection := selections[i]

		solutionBySelection := SolutionBySelection{
			selection: selection,
			solution:  solution,
		}
		solutionsBySelection[i] = solutionBySelection
	}

	return solutionsBySelection, nil
}

func (c *SolutionCreator) calculateIndependentSolutionsBySelection(
	query SolutionQuery,
) ([]SolutionBySelection, error) {
	defaultQuery, err := NewSolutionQueryBuilder().
		fromQuery(query).
		WithSelections(nil).
		Build()
	if err != nil {
		return nil, err
	}

	defaultSolution, err := c.calculateDependentSolution(defaultQuery)
	if err != nil {
		return nil, err
	}

	solutions := c.calculateManyIndependentSolutions(
		defaultSolution,
		query.selections,
	)

	return solutions, nil
}

func (c *SolutionCreator) CreateNextSolutions(
	query NextSolutionsQuery,
) (SolutionsBySelectionEnvelope, error) {
	solutions, err := c.createNextSolutions(query)
	if err != nil {
		err = updateSolveError(err, query.ruleset, query.from)
		return SolutionsBySelectionEnvelope{}, err
	}

	return NewSolutionsBySelectionEnvelope(solutions)
}

func (c *SolutionCreator) createNextSolutions(
	query NextSolutionsQuery,
) ([]SolutionBySelection, error) {
	solutionsForDependentSelections, err := c.calculateNextSolutionsForDependentSelections(query)
	if err != nil {
		return nil, err
	}

	solutionsForIndependentSelections, err := c.calculateNextSolutionsForIndependentSelections(query)
	if err != nil {
		return nil, err
	}

	var solutions []SolutionBySelection
	solutions = append(solutions, solutionsForDependentSelections...)
	solutions = append(solutions, solutionsForIndependentSelections...)

	return solutions, nil
}

func (c *SolutionCreator) calculateNextSolutionsForDependentSelections(
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

	nextDependentSolutions, err := c.calculateNextDependentSolutions(dependentQuery)
	if err != nil {
		return nil, err
	}

	currentIndependentSolution := query.ruleset.calculateIndependentSolution(
		currentIndependentSelections,
	)

	nextSolutions := make([]SolutionBySelection, len(nextDependentSolutions))
	for i, nextSolution := range nextDependentSolutions {
		nextSolutions[i] = SolutionBySelection{
			selection: nextSolution.selection,
			solution:  nextSolution.solution.merge(currentIndependentSolution),
		}
	}

	return nextSolutions, nil
}

func (c *SolutionCreator) calculateNextDependentSolutions(
	query NextSolutionsQuery,
) ([]SolutionBySelection, error) {
	solverQuery, err := c.queryCreator.newNextSolutionsQuery(query)
	if err != nil {
		return nil, err
	}

	weighted, err := newWeightedSelections(query.nextSelections, solverQuery.WeightGroups())
	if err != nil {
		return nil, err
	}

	batchable, saturated := weighted.splitBySaturation()

	batchedSolutions, err := c.calculateBatchedNextSolutions(query.ruleset, solverQuery, batchable)
	if err != nil {
		return nil, err
	}

	saturatedSolutions, err := c.calculateManySaturatedNextSolutions(query, saturated.selections())
	if err != nil {
		return nil, err
	}

	var solutions []SolutionBySelection
	solutions = append(solutions, batchedSolutions...)
	solutions = append(solutions, saturatedSolutions...)

	return solutions, nil
}

func (c *SolutionCreator) calculateBatchedNextSolutions(
	ruleset Ruleset,
	solverQuery *MultiWeightSolverQuery,
	batchable weightedSelections,
) ([]SolutionBySelection, error) {
	if len(batchable) == 0 {
		return nil, nil
	}

	batchedQuery := NewMultiWeightSolverQuery(
		solverQuery.Polyhedron(),
		solverQuery.Variables(),
		batchable.weightGroups(),
	)

	solutions, err := c.SolveWithManyWeights(batchedQuery)
	if err != nil {
		return nil, err
	}

	primitiveSolutions := ruleset.RemoveSupportVariablesForMany(solutions)

	return c.groupSolutionsBySelection(primitiveSolutions, batchable.selections())
}

// Saturated weight groups cannot share a batched request. Each next selection
// outweighs all current selections, so the split has to start above the next
// selection and cannot be shared between the groups - solve them one at a time
// and let calculateDependentSolution split each one.
func (c *SolutionCreator) calculateManySaturatedNextSolutions(
	query NextSolutionsQuery,
	nextSelections Selections,
) ([]SolutionBySelection, error) {
	solutions := make([]Solution, len(nextSelections))
	for i, nextSelection := range nextSelections {
		solution, err := c.calculateSaturatedNextDependentSolution(query, nextSelection)
		if err != nil {
			return nil, err
		}

		solutions[i] = solution
	}

	return c.groupSolutionsBySelection(solutions, nextSelections)
}

func (c *SolutionCreator) calculateSaturatedNextDependentSolution(
	query NextSolutionsQuery,
	nextSelection Selection,
) (Solution, error) {
	selections := query.currentSelections.copy()

	selections = append(selections, nextSelection)

	selectionQuery, err := NewSolutionQueryBuilder().
		WithRuleset(query.ruleset).
		WithFrom(query.from).
		WithTo(query.to).
		WithSelections(selections).
		Build()
	if err != nil {
		return Solution{}, err
	}

	return c.calculateDependentSolution(selectionQuery)
}

func (c *SolutionCreator) calculateNextSolutionsForIndependentSelections(
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

	currentSolution, err := c.calculateSolution(currentQuery)
	if err != nil {
		return nil, err
	}

	_, nextIndependentSelections := query.ruleset.CategorizeSelections(
		query.nextSelections,
	)

	nextSolutions := c.calculateManyIndependentSolutions(
		currentSolution,
		nextIndependentSelections,
	)

	return nextSolutions, nil
}

func (c *SolutionCreator) calculateManyIndependentSolutions(
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

// Carries the weight-group and corresponding next selection.
type (
	weightedSelection struct {
		selection Selection
		weights   weights.Weights
	}
	weightedSelections []weightedSelection
)

func newWeightedSelections(
	selections Selections,
	weightGroups []weights.Weights,
) (weightedSelections, error) {
	if len(selections) != len(weightGroups) {
		return nil, errors.Errorf(
			"%w: expected %d weight groups, got %d",
			puanerror.InvalidArgument,
			len(selections),
			len(weightGroups),
		)
	}

	weighted := make(weightedSelections, len(selections))
	for i, selection := range selections {
		weighted[i] = weightedSelection{
			selection: selection,
			weights:   weightGroups[i],
		}
	}

	return weighted, nil
}

func (w weightedSelections) splitBySaturation() (weightedSelections, weightedSelections) {
	var batchable weightedSelections
	var saturated weightedSelections

	for _, weighted := range w {
		tooLarge := weighted.weights.WeightsTooLarge()
		if tooLarge {
			saturated = append(saturated, weighted)

			continue
		}

		batchable = append(batchable, weighted)
	}

	return batchable, saturated
}

func (w weightedSelections) selections() Selections {
	selections := make(Selections, len(w))
	for i, weighted := range w {
		selections[i] = weighted.selection
	}

	return selections
}

func (w weightedSelections) weightGroups() []weights.Weights {
	weightGroups := make([]weights.Weights, len(w))
	for i, weighted := range w {
		weightGroups[i] = weighted.weights
	}

	return weightGroups
}
