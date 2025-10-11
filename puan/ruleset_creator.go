package puan

import (
	"fmt"
	"time"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
)

type RuleSetCreator struct {
	model              *pldag.Model
	preferredVariables []string
	assumedVariables   []string

	period                    *Period
	timeBoundAssumedVariables timeBoundVariables
}

func NewRuleSetCreator() *RuleSetCreator {
	return &RuleSetCreator{
		model: pldag.New(),
	}
}

func (c *RuleSetCreator) AddPrimitives(primitives ...string) error {
	return c.model.SetPrimitives(primitives...)
}

func (c *RuleSetCreator) SetAnd(variables ...string) (string, error) {
	return c.model.SetAnd(variables...)
}

func (c *RuleSetCreator) SetOr(variables ...string) (string, error) {
	return c.model.SetOr(variables...)
}

func (c *RuleSetCreator) SetNot(variable ...string) (string, error) {
	return c.model.SetNot(variable...)
}

func (c *RuleSetCreator) SetImply(condition, consequence string) (string, error) {
	return c.model.SetImply(condition, consequence)
}

func (c *RuleSetCreator) SetXor(variables ...string) (string, error) {
	return c.model.SetXor(variables...)
}

func (c *RuleSetCreator) SetOneOrNone(variables ...string) (string, error) {
	return c.model.SetOneOrNone(variables...)
}

func (c *RuleSetCreator) SetEquivalent(variableOne, variableTwo string) (string, error) {
	return c.model.SetEquivalent(variableOne, variableTwo)
}

func (c *RuleSetCreator) Prefer(ids ...string) error {
	dedupedIDs := utils.Dedupe(ids)
	unpreferredIDs := utils.Without(dedupedIDs, c.preferredVariables)

	err := c.model.ValidateVariables(unpreferredIDs...)
	if err != nil {
		return err
	}

	negatedIDs, err := c.negatePreferreds(unpreferredIDs)
	if err != nil {
		return err
	}

	c.preferredVariables = append(c.preferredVariables, negatedIDs...)

	return nil
}

func (c *RuleSetCreator) negatePreferreds(ids []string) ([]string, error) {
	negatedIDs := make([]string, len(ids))
	for i, id := range ids {
		negatedID, err := c.model.SetNot(id)
		if err != nil {
			return nil, err
		}

		negatedIDs[i] = negatedID
	}

	return negatedIDs, nil
}

func (c *RuleSetCreator) Assume(ids ...string) error {
	dedupedIDs := utils.Dedupe(ids)
	unassumedIDs := utils.Without(dedupedIDs, c.assumedVariables)

	err := c.model.ValidateVariables(unassumedIDs...)
	if err != nil {
		return err
	}

	c.assumedVariables = append(c.assumedVariables, unassumedIDs...)

	return nil
}

func (c *RuleSetCreator) AssumeInPeriod(
	id string,
	from, to time.Time,
) error {
	variable, err := c.newTimeBoundVariable(id, from, to)
	if err != nil {
		return err
	}

	c.timeBoundAssumedVariables = append(c.timeBoundAssumedVariables, variable)
	return nil
}

func (c *RuleSetCreator) newTimeBoundVariable(
	id string,
	from, to time.Time,
) (timeBoundVariable, error) {
	if c.period == nil {
		return timeBoundVariable{}, errors.New("time support not enabled. Call EnableTime() first")
	}

	period, err := NewPeriod(from, to)
	if err != nil {
		return timeBoundVariable{}, err
	}

	if !c.period.contains(period) {
		return timeBoundVariable{},
			errors.Errorf(
				"period %v is outside of enabled period %v",
				period,
				*c.period,
			)
	}

	return timeBoundVariable{
		variable: id,
		period:   period,
	}, nil
}

func (c *RuleSetCreator) EnableTime(
	from, to time.Time,
) error {
	period, err := NewPeriod(from, to)
	if err != nil {
		return err
	}

	c.period = &period

	return nil
}

func (c *RuleSetCreator) Create() (*RuleSet, error) {
	periodVariables, err := c.newPeriodVariables()
	if err != nil {
		return nil, err
	}

	err = c.createPeriodConstraints(periodVariables)
	if err != nil {
		return nil, err
	}

	err = c.createAssumeConstraints()
	if err != nil {
		return nil, err
	}

	// Sort variables and constraints to ensure
	// consistent order in the polyhedron,
	// this to facilitate testing
	sortedVariables := utils.Sorted(c.model.Variables())
	sortedConstraints := utils.SortedBy(
		c.model.Constraints(),
		func(c pldag.Constraint) string {
			return c.ID()
		},
	)

	assumedConstraints := c.model.AssumedConstraints()

	polyhedron := pldag.CreatePolyhedron(
		sortedVariables,
		sortedConstraints,
		assumedConstraints,
	)

	selectableVariables := utils.Without(c.model.PrimitiveVariables(), periodVariables.ids())

	return &RuleSet{
		polyhedron:          polyhedron,
		selectableVariables: selectableVariables,
		variables:           sortedVariables,
		preferredVariables:  c.preferredVariables,
		periodVariables:     periodVariables,
	}, nil
}

func (c *RuleSetCreator) newPeriodVariables() (timeBoundVariables, error) {
	if len(c.timeBoundAssumedVariables) == 0 {
		return nil, nil
	}

	nonOverlappingPeriods := calculateCompletePeriods(
		c.periods(),
	)

	// Create variable for each period
	periodVariables := make(timeBoundVariables, len(nonOverlappingPeriods))
	for i, period := range nonOverlappingPeriods {
		period := timeBoundVariable{
			variable: fmt.Sprintf("period_%d", i),
			period:   period,
		}
		periodVariables[i] = period
		if err := c.model.SetPrimitives(period.variable); err != nil {
			return nil, err
		}
	}

	return periodVariables, nil
}

func (c *RuleSetCreator) periods() []Period {
	periods := []Period{}
	periods = append(periods, *c.period)
	periods = append(periods, c.timeBoundAssumedVariables.periods()...)
	return periods
}

func (c *RuleSetCreator) createPeriodConstraints(periodVariables timeBoundVariables) error {
	if len(c.timeBoundAssumedVariables) == 0 {
		return nil
	}

	groupedByPeriods, err := groupByPeriods(periodVariables, c.timeBoundAssumedVariables)
	if err != nil {
		return err
	}

	var constraintIDs []string
	for serializedPeriodIDs, assumedIDs := range groupedByPeriods {
		periodIDs := serializedPeriodIDs.ids()
		constraintID, err := c.setTimeBoundConstraint(periodIDs, assumedIDs)
		if err != nil {
			return err
		}
		constraintIDs = append(constraintIDs, constraintID)
	}

	// Choose exactly one period
	exactlyOnePeriod, err := c.model.SetXor(periodVariables.ids()...)
	if err != nil {
		return err
	}
	constraintIDs = append(constraintIDs, exactlyOnePeriod)

	return c.Assume(constraintIDs...)
}

func (c *RuleSetCreator) setTimeBoundConstraint(
	periodIDs []string,
	assumedIDs []string,
) (string, error) {
	combinedPeriodsID, err := c.setSingleOrOR(periodIDs...)
	if err != nil {
		return "", err
	}

	combinedAssumedID, err := c.setSingleOrAnd(assumedIDs...)
	if err != nil {
		return "", err
	}

	return c.model.SetImply(combinedPeriodsID, combinedAssumedID)
}

func (c *RuleSetCreator) setSingleOrOR(ids ...string) (string, error) {
	if len(ids) == 0 {
		return "", errors.New("at least one id is required")
	}

	if len(ids) == 1 {
		return ids[0], nil
	}

	return c.model.SetOr(ids...)
}

func (c *RuleSetCreator) setSingleOrAnd(ids ...string) (string, error) {
	if len(ids) == 0 {
		return "", errors.New("at least one id is required")
	}

	if len(ids) == 1 {
		return ids[0], nil
	}

	return c.model.SetAnd(ids...)
}

func (c *RuleSetCreator) createAssumeConstraints() error {
	if len(c.assumedVariables) == 0 {
		return nil
	}

	root, err := c.setSingleOrAnd(c.assumedVariables...)
	if err != nil {
		return err
	}

	return c.model.Assume(root)
}
