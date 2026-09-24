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

func (q NextSolutionsQuery) prepareRuleset() (Ruleset, error) {
	preparedRuleset, err := q.ruleset.modifyForQuery(
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

func (q NextSolutionsQuery) batchableQuery(
	weightGroups []weights.Weights,
) (NextSolutionsQuery, error) {
	allSelections := q.nextSelections.copy()
	var batchableSelections Selections
	for i, group := range weightGroups {
		if group.WeightsTooLarge() {
			continue
		}

		selection := allSelections[i]
		batchableSelections = append(batchableSelections, selection)
	}

	query, err := NewNextSolutionsQuery(
		q.currentSelections,
		batchableSelections,
		q.ruleset,
		q.from,
		q.to,
	)
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
		if group.WeightsTooLarge() {
			selection := allSelections[i]
			nonBatchableSelections = append(nonBatchableSelections, selection)
		}
	}

	query, err := NewNextSolutionsQuery(
		q.currentSelections,
		nonBatchableSelections,
		q.ruleset,
		q.from,
		q.to,
	)
	if err != nil {
		return NextSolutionsQuery{}, err
	}

	return query, nil
}
