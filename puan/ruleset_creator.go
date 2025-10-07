package puan

import (
	"fmt"
	"time"

	"github.com/go-errors/errors"
	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
)

type RuleSetCreator struct {
	pldag              *pldag.Model
	preferredVariables []string
	assumedVariables   []string

	period                    *Period
	timeBoundAssumedVariables timeBoundVariables
}

func NewRuleSetCreator() *RuleSetCreator {
	return &RuleSetCreator{
		pldag: pldag.New(),
	}
}

func (c *RuleSetCreator) AddPrimitives(primitives ...string) error {
	return c.pldag.SetPrimitives(primitives...)
}

func (c *RuleSetCreator) SetAnd(variables ...string) (string, error) {
	return c.pldag.SetAnd(variables...)
}

func (c *RuleSetCreator) SetOr(variables ...string) (string, error) {
	return c.pldag.SetOr(variables...)
}

func (c *RuleSetCreator) SetNot(variable ...string) (string, error) {
	return c.pldag.SetNot(variable...)
}

func (c *RuleSetCreator) SetImply(condition, consequence string) (string, error) {
	return c.pldag.SetImply(condition, consequence)
}

func (c *RuleSetCreator) SetXor(variables ...string) (string, error) {
	return c.pldag.SetXor(variables...)
}

func (c *RuleSetCreator) SetOneOrNone(variables ...string) (string, error) {
	return c.pldag.SetOneOrNone(variables...)
}

func (c *RuleSetCreator) SetEquivalent(variableOne, variableTwo string) (string, error) {
	return c.pldag.SetEquivalent(variableOne, variableTwo)
}

func (c *RuleSetCreator) Prefer(ids ...string) error {
	dedupedIDs := utils.Dedupe(ids)
	unpreferredIDs := utils.Without(dedupedIDs, c.preferredVariables)

	err := c.pldag.ValidateVariables(unpreferredIDs...)
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
		negatedID, err := c.pldag.SetNot(id)
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

	err := c.pldag.ValidateVariables(unassumedIDs...)
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

	if !c.period.Contains(period) {
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
	periodVariables, err := c.newValidityPeriodConstraints()
	if err != nil {
		return nil, err
	}

	err = c.createAssumeConstraints()
	if err != nil {
		return nil, err
	}

	polyhedron := c.pldag.NewPolyhedron()
	variables := c.pldag.Variables()
	primitiveVariables := utils.Without(c.pldag.PrimitiveVariables(), periodVariables.ids())

	return &RuleSet{
		polyhedron:         polyhedron,
		primitiveVariables: primitiveVariables,
		variables:          variables,
		preferredVariables: c.preferredVariables,
		periodVariables:    periodVariables,
	}, nil
}

func (c *RuleSetCreator) newValidityPeriodConstraints() (timeBoundVariables, error) {
	if len(c.timeBoundAssumedVariables) == 0 {
		return nil, nil
	}

	// find all true periods
	// Input:
	// ....|-------|........
	// ........|-------|....
	// Output:
	// ....|---|---|---|....
	// Create variable for each
	// We also need 2 extra period variables:
	// {start-of-time}-{start-of-first-period}
	// {end-of-last-period}-{end-of-time}
	periods := []Period{}
	periods = append(periods, *c.period)
	periods = append(periods, c.timeBoundAssumedVariables.periods()...)
	nonOverlappingPeriods := calculateCompletePeriods(
		periods,
	)

	// Create variable for each period
	periodVariables := make(timeBoundVariables, len(nonOverlappingPeriods))
	for i, period := range nonOverlappingPeriods {
		period := timeBoundVariable{
			variable: fmt.Sprintf("period_%d", i),
			period:   period,
		}
		periodVariables[i] = period
		if err := c.pldag.SetPrimitives(period.variable); err != nil {
			return nil, err
		}
	}

	groupedByPeriods := groupByPeriods(periodVariables, c.timeBoundAssumedVariables)

	var constraintIDs []string
	for periodVariables, assumedVariables := range groupedByPeriods {
		constraintID, err := c.createTimeBoundConstraint(periodVariables, assumedVariables)
		if err != nil {
			return nil, err
		}
		constraintIDs = append(constraintIDs, constraintID)
	}

	// Create XOR constraint between the period variables
	exactlyOnePeriod, err := c.pldag.SetXor(periodVariables.ids()...)
	if err != nil {
		return nil, err
	}
	if err := c.pldag.Assume(exactlyOnePeriod); err != nil {
		return nil, err
	}

	for _, constraintID := range constraintIDs {
		if err := c.pldag.Assume(constraintID); err != nil {
			return nil, err
		}
	}

	return periodVariables, nil
}

func (c *RuleSetCreator) createTimeBoundConstraint(
	periodVariables periodVariables,
	assumedVariables []string,
) (string, error) {
	periodIDs := periodVariables.variables()

	var combinedPeriodsID string
	var err error
	if len(periodIDs) == 1 {
		combinedPeriodsID = periodIDs[0]
	} else {
		combinedPeriodsID, err = c.pldag.SetOr(periodIDs...)
		if err != nil {
			return "", err
		}
	}

	var combinedAssumedID string
	if len(assumedVariables) == 1 {
		combinedAssumedID = assumedVariables[0]
	} else {
		combinedAssumedID, err = c.pldag.SetAnd(assumedVariables...)
		if err != nil {
			return "", err
		}
	}

	return c.pldag.SetImply(combinedPeriodsID, combinedAssumedID)
}

func (c *RuleSetCreator) createAssumeConstraints() error {
	if len(c.assumedVariables) == 0 {
		return nil
	}

	root, err := c.pldag.SetAnd(c.assumedVariables...)
	if err != nil {
		return err
	}

	return c.pldag.Assume(root)
}
