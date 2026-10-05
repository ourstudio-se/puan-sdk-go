package puan

import (
	"time"

	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
)

type SolutionQuery struct {
	selections Selections
	ruleset    Ruleset
	from       *time.Time
	to         *time.Time
}

func NewSolutionQuery(
	selections Selections,
	ruleset Ruleset,
	from *time.Time,
	to *time.Time,
) (SolutionQuery, error) {
	if err := validateRuleset(ruleset); err != nil {
		return SolutionQuery{}, err
	}

	if err := validateTimestamps(from, to); err != nil {
		return SolutionQuery{}, err
	}

	if err := validateSelections(ruleset, selections); err != nil {
		return SolutionQuery{}, err
	}

	return SolutionQuery{
		selections: selections,
		ruleset:    ruleset,
		from:       from,
		to:         to,
	}, nil
}

func (query SolutionQuery) asSolverQuery() (*SolverQuery, error) {
	preparedRuleset, err := query.ruleset.prepareForSolve(query.selections, query.from, query.to)
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

func (query SolutionQuery) asSolutionsBySelectionSolverQuery() (*MultiWeightSolverQuery, error) {
	preparedRuleset, err := query.ruleset.prepareForSolve(query.selections, query.from, query.to)
	if err != nil {
		return nil, err
	}

	weightGroups, err := calculateWeightsBySelection(preparedRuleset, query.selections)
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

func calculateWeightsBySelection(
	ruleset Ruleset,
	selections Selections,
) ([]weights.Weights, error) {
	selectionGroups := make([]Selections, len(selections))
	for i, selection := range selections {
		selectionGroups[i] = Selections{selection}
	}

	return calculateWeightGroups(ruleset, selectionGroups)
}

type SolutionQueryBuilder struct {
	selections Selections
	ruleset    Ruleset
	from       *time.Time
	to         *time.Time
}

func NewSolutionQueryBuilder() *SolutionQueryBuilder {
	return &SolutionQueryBuilder{}
}

func (b *SolutionQueryBuilder) fromQuery(
	query SolutionQuery,
) *SolutionQueryBuilder {
	b.selections = query.selections
	b.ruleset = query.ruleset
	b.from = query.from
	b.to = query.to
	return b
}

func (b *SolutionQueryBuilder) WithSelections(selections Selections) *SolutionQueryBuilder {
	b.selections = selections
	return b
}

func (b *SolutionQueryBuilder) WithRuleset(ruleset Ruleset) *SolutionQueryBuilder {
	b.ruleset = ruleset
	return b
}

func (b *SolutionQueryBuilder) WithFrom(from *time.Time) *SolutionQueryBuilder {
	b.from = from
	return b
}

func (b *SolutionQueryBuilder) WithTo(to *time.Time) *SolutionQueryBuilder {
	b.to = to
	return b
}

func (b *SolutionQueryBuilder) Build() (SolutionQuery, error) {
	return NewSolutionQuery(b.selections, b.ruleset, b.from, b.to)
}
