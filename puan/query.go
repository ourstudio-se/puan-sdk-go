package puan

import (
	"time"

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

func (c *queryCreator) create(
	selections Selections,
	ruleset Ruleset,
	from *time.Time,
	to *time.Time,
) (*Query, error) {
	preparedRuleset, err := ruleset.modifyForQuery(selections, from, to)
	if err != nil {
		return nil, err
	}

	weights, err := newWeights(preparedRuleset, selections)
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

func (c *queryCreator) newSolutionsBySelectionQuery(
	selections Selections,
	ruleset Ruleset,
	from *time.Time,
	to *time.Time,
) (*MultiWeightQuery, error) {
	preparedRuleset, err := ruleset.modifyForQuery(selections, from, to)
	if err != nil {
		return nil, err
	}

	weightGroups, err := c.calculateWeightsForSolutionsBySelection(preparedRuleset, selections)
	if err != nil {
		return nil, err
	}

	query := NewMultiWeightQuery(
		preparedRuleset.polyhedron,
		preparedRuleset.dependentVariables,
		weightGroups,
	)

	return query, nil
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
