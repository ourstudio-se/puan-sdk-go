package puan

import (
	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
)

type Query struct {
	polyhedron *pldag.Polyhedron
	variables  []string
	weights    weights.Weights
}

func NewQuery(polyhedron *pldag.Polyhedron, variables []string, weights weights.Weights) *Query {
	return &Query{
		polyhedron: polyhedron,
		variables:  variables,
		weights:    weights,
	}
}

func (q *Query) Polyhedron() *pldag.Polyhedron {
	return q.polyhedron
}

func (q *Query) Variables() []string {
	return q.variables
}

func (q *Query) Weights() weights.Weights {
	return q.weights
}

type MultiWeightQuery struct {
	polyhedron   *pldag.Polyhedron
	variables    []string
	weightGroups []weights.Weights
}

func NewMultiWeightQuery(
	polyhedron *pldag.Polyhedron,
	variables []string,
	weightGroups []weights.Weights,
) *MultiWeightQuery {
	return &MultiWeightQuery{
		polyhedron:   polyhedron,
		variables:    variables,
		weightGroups: weightGroups,
	}
}

func (q *MultiWeightQuery) Polyhedron() *pldag.Polyhedron {
	return q.polyhedron
}

func (q *MultiWeightQuery) Variables() []string {
	return q.variables
}

func (q *MultiWeightQuery) WeightGroups() []weights.Weights {
	return q.weightGroups
}

type queryCreator struct{}

func newQueryCreator() *queryCreator {
	return &queryCreator{}
}

func (c *queryCreator) create(query SolutionQuery) (*Query, error) {
	preparedRuleset, err := query.ruleset.modifyForQuery(query.selections, query.from, query.to)
	if err != nil {
		return nil, err
	}

	weights, err := newWeights(preparedRuleset, query.selections)
	if err != nil {
		return nil, err
	}

	solverQuery := NewQuery(
		preparedRuleset.polyhedron,
		preparedRuleset.dependentVariables,
		weights,
	)

	return solverQuery, nil
}

func (c *queryCreator) newSolutionsBySelectionQuery(
	query SolutionQuery,
) (*MultiWeightQuery, error) {
	preparedRuleset, err := query.ruleset.modifyForQuery(query.selections, query.from, query.to)
	if err != nil {
		return nil, err
	}

	weightGroups, err := c.calculateWeightsForSolutionsBySelection(preparedRuleset, query.selections)
	if err != nil {
		return nil, err
	}

	solverQuery := NewMultiWeightQuery(
		preparedRuleset.polyhedron,
		preparedRuleset.dependentVariables,
		weightGroups,
	)

	return solverQuery, nil
}

func (c *queryCreator) calculateWeightsForSolutionsBySelection(
	ruleset Ruleset,
	selections Selections,
) ([]weights.Weights, error) {
	weightsBySelection := make([]weights.Weights, len(selections))
	for i, selection := range selections {
		modifiedSelections := Selections{selection}.prepareForQuery()

		weights, err := newWeights(ruleset, modifiedSelections)
		if err != nil {
			return nil, err
		}
		weightsBySelection[i] = weights
	}

	return weightsBySelection, nil
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
