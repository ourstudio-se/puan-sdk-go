package puan

import (
	"time"

	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
)

type NextSolutionsQuery struct {
	currentSelections Selections
	nextSelections    Selections
	ruleset           Ruleset
	from              *time.Time
	to                *time.Time
}

func NewNextSolutionsQuery(
	currentSelections Selections,
	nextSelections Selections,
	ruleset Ruleset,
	from *time.Time,
	to *time.Time,
) (NextSolutionsQuery, error) {
	if err := validateRuleset(ruleset); err != nil {
		return NextSolutionsQuery{}, err
	}

	if err := validateTimestamps(from, to); err != nil {
		return NextSolutionsQuery{}, err
	}

	if err := validateSelections(ruleset, currentSelections); err != nil {
		return NextSolutionsQuery{}, err
	}

	if err := validateSelections(ruleset, nextSelections); err != nil {
		return NextSolutionsQuery{}, err
	}

	return NextSolutionsQuery{
		currentSelections: currentSelections,
		nextSelections:    nextSelections,
		ruleset:           ruleset,
		from:              from,
		to:                to,
	}, nil
}

type NextSolutionsQueryBuilder struct {
	currentSelections Selections
	nextSelections    Selections
	ruleset           Ruleset
	from              *time.Time
	to                *time.Time
}

func NewNextSolutionsQueryBuilder() *NextSolutionsQueryBuilder {
	return &NextSolutionsQueryBuilder{}
}

func (b *NextSolutionsQueryBuilder) fromQuery(
	query NextSolutionsQuery,
) *NextSolutionsQueryBuilder {
	b.currentSelections = query.currentSelections
	b.nextSelections = query.nextSelections
	b.ruleset = query.ruleset
	b.from = query.from
	b.to = query.to
	return b
}

func (b *NextSolutionsQueryBuilder) WithCurrentSelections(
	selections Selections,
) *NextSolutionsQueryBuilder {
	b.currentSelections = selections
	return b
}

func (b *NextSolutionsQueryBuilder) WithNextSelections(
	selections Selections,
) *NextSolutionsQueryBuilder {
	b.nextSelections = selections
	return b
}

func (b *NextSolutionsQueryBuilder) WithRuleset(ruleset Ruleset) *NextSolutionsQueryBuilder {
	b.ruleset = ruleset
	return b
}

func (b *NextSolutionsQueryBuilder) WithFrom(from *time.Time) *NextSolutionsQueryBuilder {
	b.from = from
	return b
}

func (b *NextSolutionsQueryBuilder) WithTo(to *time.Time) *NextSolutionsQueryBuilder {
	b.to = to
	return b
}

func (b *NextSolutionsQueryBuilder) Build() (NextSolutionsQuery, error) {
	return NewNextSolutionsQuery(
		b.currentSelections,
		b.nextSelections,
		b.ruleset,
		b.from,
		b.to,
	)
}

func (q NextSolutionsQuery) hasEmptyNextSelections() bool {
	return len(q.nextSelections) == 0
}

func (q NextSolutionsQuery) asSolutionQueries() ([]SolutionQuery, error) {
	queries := make([]SolutionQuery, len(q.nextSelections))
	for i, nextSelection := range q.nextSelections {
		selections := q.currentSelections.copy()
		selections = append(selections, nextSelection)

		query, err := NewSolutionQueryBuilder().
			WithRuleset(q.ruleset).
			WithFrom(q.from).
			WithTo(q.to).
			WithSelections(selections).
			Build()
		if err != nil {
			return nil, err
		}

		queries[i] = query
	}

	return queries, nil
}

func (q NextSolutionsQuery) splitByBatchability() (
	NextSolutionsQuery, NextSolutionsQuery, error,
) {
	ruleset, err := q.prepareRuleset()
	if err != nil {
		return NextSolutionsQuery{}, NextSolutionsQuery{}, err
	}

	weightGroups, err := calculateNextWeightGroups(
		ruleset,
		q.currentSelections,
		q.nextSelections,
	)
	if err != nil {
		return NextSolutionsQuery{}, NextSolutionsQuery{}, err
	}

	batchable, err := q.batchableQuery(weightGroups)
	if err != nil {
		return NextSolutionsQuery{}, NextSolutionsQuery{}, err
	}

	nonBatchable, err := q.nonBatchableQuery(weightGroups)
	if err != nil {
		return NextSolutionsQuery{}, NextSolutionsQuery{}, err
	}

	return batchable, nonBatchable, nil
}

func (q NextSolutionsQuery) prepareRuleset() (Ruleset, error) {
	preparedRuleset, err := q.ruleset.prepareForSolve(
		q.currentSelections,
		q.from,
		q.to,
	)
	if err != nil {
		return Ruleset{}, err
	}

	if err = preparedRuleset.setCompositeSelectionConstraints(q.nextSelections); err != nil {
		return Ruleset{}, err
	}

	return preparedRuleset, nil
}

func (q NextSolutionsQuery) batchableQuery(
	weightGroups []weights.Weights,
) (NextSolutionsQuery, error) {
	allSelections := q.nextSelections.copy()
	var batchableSelections Selections
	for i, group := range weightGroups {
		if group.AboveSaturationLimit() {
			continue
		}

		selection := allSelections[i]
		batchableSelections = append(batchableSelections, selection)
	}

	query, err := NewNextSolutionsQueryBuilder().
		fromQuery(q).
		WithNextSelections(batchableSelections).
		Build()
	if err != nil {
		return NextSolutionsQuery{}, err
	}

	return query, nil
}

func (q NextSolutionsQuery) nonBatchableQuery(
	weightGroups []weights.Weights,
) (NextSolutionsQuery, error) {
	allSelections := q.nextSelections.copy()
	var nonBatchableSelections Selections
	for i, group := range weightGroups {
		if group.AboveSaturationLimit() {
			selection := allSelections[i]
			nonBatchableSelections = append(nonBatchableSelections, selection)
		}
	}

	query, err := NewNextSolutionsQueryBuilder().
		fromQuery(q).
		WithNextSelections(nonBatchableSelections).
		Build()
	if err != nil {
		return NextSolutionsQuery{}, err
	}

	return query, nil
}

func (query NextSolutionsQuery) asSolverQuery() (*MultiWeightSolverQuery, error) {
	preparedRuleset, err := query.prepareRuleset()
	if err != nil {
		return nil, err
	}

	weightGroups, err := calculateNextWeightGroups(
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

func calculateNextWeightGroups(
	ruleset Ruleset,
	currentSelections Selections,
	nextSelections Selections,
) ([]weights.Weights, error) {
	selectionGroups := make([]Selections, len(nextSelections))
	for i, nextSelection := range nextSelections {
		selections := currentSelections.copy()
		selections = append(selections, nextSelection)
		selectionGroups[i] = selections
	}

	return calculateWeightGroups(ruleset, selectionGroups)
}
