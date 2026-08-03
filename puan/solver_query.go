package puan

import (
	"time"

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

type WeightsForSelection struct {
	Selection Selection
	Weights   weights.Weights
}

type MultiWeightSolverQuery struct {
	polyhedron         *pldag.Polyhedron
	variables          []string
	weightsBySelection []WeightsForSelection
}

func NewMultiWeightSolverQuery(
	polyhedron *pldag.Polyhedron,
	variables []string,
	weightsBySelection []WeightsForSelection,
) *MultiWeightSolverQuery {
	return &MultiWeightSolverQuery{
		polyhedron:         polyhedron,
		variables:          variables,
		weightsBySelection: weightsBySelection,
	}
}

func (q *MultiWeightSolverQuery) Polyhedron() *pldag.Polyhedron {
	return q.polyhedron
}

func (q *MultiWeightSolverQuery) Variables() []string {
	return q.variables
}

func (q *MultiWeightSolverQuery) WeightsBySelection() []WeightsForSelection {
	return q.weightsBySelection
}

func (q *MultiWeightSolverQuery) validate() error {
	for _, weightsForSelection := range q.weightsBySelection {
		if weightsForSelection.Weights.WeightsTooLarge() {
			return errors.Errorf(
				"weights are too large for selection %v",
				weightsForSelection.Selection,
			)
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

	weightGroups, err := c.calculateWeightsForSolutionsBySelection(preparedRuleset, query.selections)
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

func (c *solverQueryCreator) calculateWeightsForSolutionsBySelection(
	ruleset Ruleset,
	selections Selections,
) ([]WeightsForSelection, error) {
	weightGroups := make([]WeightsForSelection, len(selections))
	for i, selection := range selections {
		modifiedSelections := Selections{selection}

		weights, err := newWeights(ruleset, modifiedSelections)
		if err != nil {
			return nil, err
		}
		weightsForSelection := WeightsForSelection{
			Selection: selection,
			Weights:   weights,
		}
		weightGroups[i] = weightsForSelection
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
	currentSelections Selections,
	nextSelections Selections,
	ruleset Ruleset,
	from *time.Time,
	to *time.Time,
) (*MultiWeightSolverQuery, error) {
	preparedRuleset, err := ruleset.modifyForQuery(currentSelections, from, to)
	if err != nil {
		return nil, err
	}

	if err := preparedRuleset.setCompositeSelectionConstraints(nextSelections); err != nil {
		return nil, err
	}

	weightGroups, err := c.calculateNextWeightGroups2(
		preparedRuleset,
		currentSelections,
		nextSelections,
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

func (c *solverQueryCreator) calculateNextWeightGroups2(
	ruleset Ruleset,
	currentSelections Selections,
	nextSelections Selections,
) ([]WeightsForSelection, error) {
	weightGroups := make([]WeightsForSelection, len(nextSelections))
	for i, nextSelection := range nextSelections {
		weightsForSelection, err := c.calculateNextWeights2(ruleset, currentSelections, nextSelection)
		if err != nil {
			return nil, err
		}
		weightGroups[i] = weightsForSelection
	}

	return weightGroups, nil
}

func (c *solverQueryCreator) calculateNextWeights2(
	ruleset Ruleset,
	currentSelections Selections,
	nextSelection Selection,
) (WeightsForSelection, error) {
	selections := currentSelections.copy()

	selections = append(selections, nextSelection)

	selections = selections.prepareForQuery()

	weights, err := newWeights(ruleset, selections)
	if err != nil {
		return WeightsForSelection{}, err
	}

	return WeightsForSelection{
		Selection: nextSelection,
		Weights:   weights,
	}, nil
}
