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

func (q *MultiWeightSolverQuery) validateWeightLimit() error {
	for i, weights := range q.WeightGroups() {
		if weights.AboveSaturationLimit() {
			return errors.Errorf("weights too large at index %d", i)
		}
	}
	return nil
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

func calculateWeightGroups(
	ruleset Ruleset,
	selectionGroups []Selections,
) ([]weights.Weights, error) {
	weightGroups := make([]weights.Weights, len(selectionGroups))
	for i, selections := range selectionGroups {
		weights, err := newWeights(ruleset, selections)
		if err != nil {
			return nil, err
		}
		weightGroups[i] = weights
	}

	return weightGroups, nil
}
