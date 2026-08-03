package puan

import (
	"time"
)

type SolutionQuery struct {
	selections Selections
	ruleset    Ruleset
	from       *time.Time
	to         *time.Time
}

func (query SolutionQuery) validate() error {
	if err := validateRuleset(query.ruleset); err != nil {
		return err
	}

	if err := validateTimestamps(query.from, query.to); err != nil {
		return err
	}

	if err := validateSelections(query.ruleset, query.selections); err != nil {
		return err
	}

	return nil
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

func (b *SolutionQueryBuilder) Build() SolutionQuery {
	return SolutionQuery{
		selections: b.selections,
		ruleset:    b.ruleset,
		from:       b.from,
		to:         b.to,
	}
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
) NextSolutionsQuery {
	return NextSolutionsQuery{
		currentSelections: currentSelections,
		nextSelections:    nextSelections,
		ruleset:           ruleset,
		from:              from,
		to:                to,
	}
}

func (query NextSolutionsQuery) validate() error {
	if err := validateRuleset(query.ruleset); err != nil {
		return err
	}

	if err := validateTimestamps(query.from, query.to); err != nil {
		return err
	}

	if err := validateSelections(query.ruleset, query.currentSelections); err != nil {
		return err
	}

	if err := validateSelections(query.ruleset, query.nextSelections); err != nil {
		return err
	}

	return nil
}
