package puan

import (
	"github.com/go-errors/errors"
	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
)

type SolverQuery struct {
	polyhedron *pldag.Polyhedron
	variables  []string
	weights    weights.Weights
}

func NewSolverQuery(
	polyhedron *pldag.Polyhedron,
	variables []string,
	weights weights.Weights,
) *SolverQuery {
	return &SolverQuery{
		polyhedron: polyhedron,
		variables:  variables,
		weights:    weights,
	}
}

func (q *SolverQuery) Polyhedron() *pldag.Polyhedron {
	return q.polyhedron
}

func (q *SolverQuery) Variables() []string {
	return q.variables
}

func (q *SolverQuery) Weights() weights.Weights {
	return q.weights
}

type MultiWeightSolverQuery struct {
	polyhedron *pldag.Polyhedron
	variables  []string
	weights    []weights.Weights
}

func NewMultiWeightSolverQuery(
	polyhedron *pldag.Polyhedron,
	variables []string,
	weights []weights.Weights,
) *MultiWeightSolverQuery {
	return &MultiWeightSolverQuery{
		polyhedron: polyhedron,
		variables:  variables,
		weights:    weights,
	}
}

func (q *MultiWeightSolverQuery) Polyhedron() *pldag.Polyhedron {
	return q.polyhedron
}

func (q *MultiWeightSolverQuery) Variables() []string {
	return q.variables
}

func (q *MultiWeightSolverQuery) WeightGroups() []weights.Weights {
	return q.weights
}

func (q *MultiWeightSolverQuery) validate() error {
	for i, weights := range q.weights {
		if weights.WeightsTooLarge() {
			return errors.Errorf("weights are too large at index %d", i)
		}
	}
	return nil
}

type solverQueryCreator struct{}

func newSolverQueryCreator() *solverQueryCreator {
	return &solverQueryCreator{}
}

func (c *solverQueryCreator) new(query SolutionQuery) (*SolverQuery, error) {
	preparedRuleset, err := query.ruleset.modifyForQuery(query.selections, query.from, query.to)
	if err != nil {
		return nil, err
	}

	weights, err := newWeights(preparedRuleset, query.selections)
	if err != nil {
		return nil, err
	}

	solverQuery := NewSolverQuery(
		preparedRuleset.polyhedron,
		preparedRuleset.dependentVariables,
		weights,
	)

	return solverQuery, nil
}

func (c *solverQueryCreator) newSolutionsBySelectionQuery(
	query SolutionQuery,
) (*MultiWeightSolverQuery, error) {
	preparedRuleset, err := query.ruleset.modifyForQuery(query.selections, query.from, query.to)
	if err != nil {
		return nil, err
	}

	weightGroups, err := c.calculateWeightsForSelections(preparedRuleset, query.selections)
	if err != nil {
		return nil, err
	}

	solverQuery := NewMultiWeightSolverQuery(
		preparedRuleset.polyhedron,
		preparedRuleset.dependentVariables,
		weightGroups,
	)

	return solverQuery, nil
}

func (c *solverQueryCreator) calculateWeightsForSelections(
	ruleset Ruleset,
	selections Selections,
) ([]weights.Weights, error) {
	weightGroups := make([]weights.Weights, len(selections))
	for i, selection := range selections {
		modifiedSelections := Selections{selection}

		weights, err := newWeights(ruleset, modifiedSelections)
		if err != nil {
			return nil, err
		}
		weightGroups[i] = weights
	}

	return weightGroups, nil
}

func newWeights(
	ruleset Ruleset,
	selections Selections,
) (weights.Weights, error) {
	preparedSelections := selections.prepareForQuery()

	dependentSelectableVariables := ruleset.dependentSelectableVariables()

	weightSelections, err := ruleset.newWeightSelections(preparedSelections)
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

func (c *solverQueryCreator) newNextSolutionsQuery(
	query NextSolutionsQuery,
) (*MultiWeightSolverQuery, error) {
	preparedRuleset, err := query.ruleset.modifyForQuery(
		query.currentSelections,
		query.from,
		query.to,
	)
	if err != nil {
		return nil, err
	}

	if err := preparedRuleset.setCompositeSelectionConstraints(query.nextSelections); err != nil {
		return nil, err
	}

	weightGroups, err := c.calculateNextWeightGroups(
		preparedRuleset,
		query.currentSelections,
		query.nextSelections,
	)
	if err != nil {
		return nil, err
	}

	solverQuery := NewMultiWeightSolverQuery(
		preparedRuleset.polyhedron,
		preparedRuleset.dependentVariables,
		weightGroups,
	)

	return solverQuery, nil
}

func (c *solverQueryCreator) calculateNextWeightGroups(
	ruleset Ruleset,
	currentSelections Selections,
	nextSelections Selections,
) ([]weights.Weights, error) {
	weightGroups := make([]weights.Weights, len(nextSelections))
	for i, nextSelection := range nextSelections {
		weights, err := c.calculateNextWeights(ruleset, currentSelections, nextSelection)
		if err != nil {
			return nil, err
		}
		weightGroups[i] = weights
	}

	return weightGroups, nil
}

func (c *solverQueryCreator) calculateNextWeights(
	ruleset Ruleset,
	currentSelections Selections,
	nextSelection Selection,
) (weights.Weights, error) {
	selections := currentSelections.copy()

	selections = append(selections, nextSelection)

	selections = selections.prepareForQuery()

	weights, err := newWeights(ruleset, selections)
	if err != nil {
		return nil, err
	}

	return weights, nil
}
