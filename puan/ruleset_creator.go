package puan

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
	"github.com/ourstudio-se/puan-sdk-go/puanerror"
)

type RulesetCreator struct {
	model              *pldag.Model
	preferredVariables []string
	assumedVariables   []string

	period                      *Period
	forbiddenPeriods            []Period
	timeBoundAssumedVariables   TimeBoundVariables
	timeBoundPreferredVariables TimeBoundVariables
}

func NewRulesetCreator() *RulesetCreator {
	return &RulesetCreator{
		model: pldag.New(),
	}
}

func (c *RulesetCreator) AddPrimitives(primitives ...string) error {
	for _, primitive := range primitives {
		// Prefix 'period_' is reserved for internal use to handle time support.
		if strings.HasPrefix(primitive, "period_") {
			return errors.Errorf(
				"%w: primitive %s cannot start with reserved prefix 'period_'",
				puanerror.InvalidArgument,
				primitive,
			)
		}
	}

	return c.model.AddPrimitives(primitives...)
}

func (c *RulesetCreator) SetAnd(variables ...string) (string, error) {
	return c.model.SetAnd(variables...)
}

func (c *RulesetCreator) SetOr(variables ...string) (string, error) {
	return c.model.SetOr(variables...)
}

func (c *RulesetCreator) SetNot(variables ...string) (string, error) {
	return c.model.SetNot(variables...)
}

func (c *RulesetCreator) SetImply(condition, consequence string) (string, error) {
	return c.model.SetImply(condition, consequence)
}

func (c *RulesetCreator) SetXor(variables ...string) (string, error) {
	return c.model.SetXor(variables...)
}

func (c *RulesetCreator) SetOneOrNone(variables ...string) (string, error) {
	return c.model.SetOneOrNone(variables...)
}

func (c *RulesetCreator) SetEquivalent(variableOne, variableTwo string) (string, error) {
	return c.model.SetEquivalent(variableOne, variableTwo)
}

func (c *RulesetCreator) Prefer(ids ...string) error {
	negatedIDs, err := c.negatePreferreds(ids)
	if err != nil {
		return err
	}

	c.preferredVariables = append(c.preferredVariables, negatedIDs...)

	return nil
}

func (c *RulesetCreator) negatePreferreds(ids []string) ([]string, error) {
	negatedIDs := make([]string, len(ids))
	for i, id := range ids {
		negatedID, err := c.SetNot(id)
		if err != nil {
			return nil, err
		}

		negatedIDs[i] = negatedID
	}

	return negatedIDs, nil
}

func (c *RulesetCreator) PreferInPeriod(
	id string,
	from, to time.Time,
) error {
	variable, err := c.newTimeBoundVariable(id, from, to)
	if err != nil {
		return err
	}

	// If the variable period is equal to the ruleset period,
	// i.e., is preferred during the entire ruleset period,
	// it can be preferred directly without time bounding.
	preferredDuringRulesetPeriod := c.period.isEqual(variable.period)
	if preferredDuringRulesetPeriod {
		return c.Prefer(id)
	}

	c.timeBoundPreferredVariables = append(c.timeBoundPreferredVariables, variable)

	return nil
}

func (c *RulesetCreator) Assume(ids ...string) error {
	dedupedIDs := utils.Dedupe(ids)
	unassumedIDs := utils.Without(dedupedIDs, c.assumedVariables)

	err := c.model.ValidateVariables(unassumedIDs...)
	if err != nil {
		return err
	}

	c.assumedVariables = append(c.assumedVariables, unassumedIDs...)

	return nil
}

func (c *RulesetCreator) AssumeInPeriod(
	id string,
	from, to time.Time,
) error {
	variable, err := c.newTimeBoundVariable(id, from, to)
	if err != nil {
		return err
	}

	// If the variable period is equal to the ruleset period,
	// i.e., is available during the entire ruleset period,
	// it can be assumed directly without time bounding.
	if c.period.isEqual(variable.period) {
		return c.Assume(variable.variable)
	}

	c.timeBoundAssumedVariables = append(c.timeBoundAssumedVariables, variable)

	return nil
}

func (c *RulesetCreator) newTimeBoundVariable(
	id string,
	from, to time.Time,
) (TimeBoundVariable, error) {
	if c.timeDisabled() {
		return TimeBoundVariable{}, errors.Errorf(
			"%w: time support not enabled. Call EnableTime() first",
			puanerror.InvalidOperation,
		)
	}

	period, err := NewPeriod(from, to)
	if err != nil {
		return TimeBoundVariable{}, err
	}

	if !c.period.contains(period) {
		return TimeBoundVariable{},
			errors.Errorf(
				"%w: period %v is outside of enabled period %v",
				puanerror.InvalidArgument,
				period,
				*c.period,
			)
	}

	return TimeBoundVariable{
		variable: id,
		period:   period,
	}, nil
}

func (c *RulesetCreator) EnableTime(
	from, to time.Time,
) error {
	period, err := NewPeriod(from, to)
	if err != nil {
		return err
	}

	c.period = &period

	return nil
}

func (c *RulesetCreator) ForbidPeriod(
	from, to time.Time,
) error {
	if c.timeDisabled() {
		return errors.Errorf(
			"%w: time support not enabled. Call EnableTime() first",
			puanerror.InvalidOperation,
		)
	}

	period, err := NewPeriod(from, to)
	if err != nil {
		return err
	}

	if !c.period.contains(period) {
		return errors.Errorf(
			"%w: period %v is outside of enabled period %v",
			puanerror.InvalidArgument,
			period,
			*c.period,
		)
	}

	for _, existingForbiddenPeriod := range c.forbiddenPeriods {
		if existingForbiddenPeriod.overlaps(period) {
			return errors.Errorf(
				"%w: period %v overlaps with existing forbidden period %v",
				puanerror.InvalidArgument,
				period,
				existingForbiddenPeriod,
			)
		}
	}

	c.forbiddenPeriods = append(c.forbiddenPeriods, period)

	return nil
}

func (c *RulesetCreator) Create() (Ruleset, error) {
	periodVariables, err := c.createPeriodVariables()
	if err != nil {
		return Ruleset{}, err
	}

	err = c.createPeriodConstraints(periodVariables)
	if err != nil {
		return Ruleset{}, err
	}

	err = c.createPeriodPreferreds(periodVariables)
	if err != nil {
		return Ruleset{}, err
	}

	err = c.createAssumeConstraints()
	if err != nil {
		return Ruleset{}, err
	}

	dependentVariables := c.findDependantVariables()
	independentVariables := utils.Without(c.model.PrimitiveVariables(), dependentVariables)
	selectableVariables := utils.Without(c.model.PrimitiveVariables(), periodVariables.ids())
	preferredVariables := utils.Dedupe(c.preferredVariables)

	// Sort dependentVariables and constraints to ensure
	// consistent order in the polyhedron,
	// this to facilitate testing
	sortedDependentVariables := utils.Sorted(dependentVariables)
	sortedConstraints := utils.SortedBy(
		c.model.Constraints(),
		func(c pldag.Constraint) string {
			return c.ID()
		},
	)

	polyhedron := pldag.CreatePolyhedron(
		sortedDependentVariables,
		sortedConstraints,
		c.model.AssumedConstraints(),
	)

	return newRuleset(
		polyhedron,
		selectableVariables,
		sortedDependentVariables,
		independentVariables,
		preferredVariables,
		periodVariables,
	)
}

func (c *RulesetCreator) findDependantVariables() []string {
	constraintVariables := c.model.Constraints().Variables()
	assumedVariables := c.model.AssumedConstraints().Variables()

	return utils.Union(constraintVariables, assumedVariables)
}

func (c *RulesetCreator) createPeriodVariables() (TimeBoundVariables, error) {
	if c.timeDisabled() {
		return nil, nil
	}

	partitionedPeriods := c.calculateAllowedPartitionedPeriods()

	periodVariables, err := c.newPeriodVariables(partitionedPeriods)
	if err != nil {
		return nil, err
	}

	err = c.model.AddPrimitives(periodVariables.ids()...)
	if err != nil {
		return nil, err
	}

	return periodVariables, nil
}

func (c *RulesetCreator) calculateAllowedPartitionedPeriods() []Period {
	periods := c.periods()

	partitionedPeriods := calculatePartitionedPeriods(periods)

	allowedPartitionedPeriods := filterOutForbiddenPeriods(
		partitionedPeriods,
		c.forbiddenPeriods,
	)

	return allowedPartitionedPeriods
}

func (c *RulesetCreator) timeDisabled() bool {
	return c.period == nil
}

func (c *RulesetCreator) newPeriodVariables(
	orderedPeriods []Period,
) (TimeBoundVariables, error) {
	for i := range len(orderedPeriods) - 1 {
		previous := orderedPeriods[i]
		current := orderedPeriods[i+1]
		invalidOrder := current.from.Before(previous.to)
		if invalidOrder {
			return nil, errors.Errorf(
				"periods %v and %v does not have expected order or overlap",
				orderedPeriods[i],
				orderedPeriods[i+1],
			)
		}
	}

	periodVariables := make(TimeBoundVariables, len(orderedPeriods))
	for i, period := range orderedPeriods {
		periodVariable := TimeBoundVariable{
			variable: fmt.Sprintf("period_%d", i),
			period:   period,
		}
		periodVariables[i] = periodVariable
	}

	return periodVariables, nil
}

func (c *RulesetCreator) periods() []Period {
	periods := []Period{}
	periods = append(periods, *c.period)
	periods = append(periods, c.forbiddenPeriods...)
	periods = append(periods, c.timeBoundAssumedVariables.periods()...)
	periods = append(periods, c.timeBoundPreferredVariables.periods()...)
	return periods
}

func (c *RulesetCreator) createPeriodConstraints(
	periodVariables TimeBoundVariables,
) error {
	if c.timeDisabled() {
		return nil
	}

	if err := c.createTimeBoundAssumeConstraints(periodVariables); err != nil {
		return err
	}

	if err := c.createExactlyOnePeriodConstraint(periodVariables); err != nil {
		return err
	}

	return nil
}

func (c *RulesetCreator) createTimeBoundAssumeConstraints(
	periodVariables TimeBoundVariables,
) error {
	assumedVariables := c.getTimeBoundAssumedVariablesInPeriods(periodVariables.periods())

	groupedByPeriods, err := groupByPeriods(periodVariables, assumedVariables)
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

	return c.Assume(constraintIDs...)
}

func (c *RulesetCreator) getTimeBoundAssumedVariablesInPeriods(
	periods []Period,
) TimeBoundVariables {
	return c.timeBoundAssumedVariables.containing(
		periods,
	)
}

func (c *RulesetCreator) createExactlyOnePeriodConstraint(
	periodVariables TimeBoundVariables,
) error {
	exactlyOnePeriod, err := c.setSingleOrXOR(periodVariables.ids()...)
	if err != nil {
		return err
	}
	return c.Assume(exactlyOnePeriod)
}

func (c *RulesetCreator) createPeriodPreferreds(
	periodVariables TimeBoundVariables,
) error {
	if c.timeDisabled() {
		return nil
	}

	preferredVariables := c.getTimeBoundPreferredVariablesInPeriods(
		periodVariables.periods(),
	)

	groupedByPeriods, err := groupByPeriods(periodVariables, preferredVariables)
	if err != nil {
		return err
	}

	for serializedPeriodIDs, preferredIDs := range groupedByPeriods {
		periodIDs := serializedPeriodIDs.ids()
		anyPeriodID, err := c.setSingleOrOR(periodIDs...)
		if err != nil {
			return err
		}

		err = c.createPreferredsInPeriod(anyPeriodID, preferredIDs...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *RulesetCreator) getTimeBoundPreferredVariablesInPeriods(
	periods []Period,
) TimeBoundVariables {
	return c.timeBoundPreferredVariables.containing(
		periods,
	)
}

func (c *RulesetCreator) createPreferredsInPeriod(periodsID string, preferredIDs ...string) error {
	for _, preferredID := range preferredIDs {
		negatedID, err := c.SetNot(preferredID)
		if err != nil {
			return err
		}

		id, err := c.SetAnd(periodsID, negatedID)
		if err != nil {
			return err
		}

		c.preferredVariables = append(c.preferredVariables, id)
	}

	return nil
}

func (c *RulesetCreator) setTimeBoundConstraint(
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

	return c.SetImply(combinedPeriodsID, combinedAssumedID)
}

func (c *RulesetCreator) setSingleOrOR(ids ...string) (string, error) {
	if len(ids) == 0 {
		return "", errors.Errorf(
			"%w: at least one id is required",
			puanerror.InvalidArgument,
		)
	}

	deduped := utils.Dedupe(ids)
	if len(deduped) == 1 {
		return deduped[0], nil
	}

	return c.SetOr(deduped...)
}

func (c *RulesetCreator) setSingleOrXOR(ids ...string) (string, error) {
	if len(ids) == 0 {
		return "", errors.Errorf(
			"%w: at least one id is required",
			puanerror.InvalidArgument,
		)
	}

	deduped := utils.Dedupe(ids)
	if len(deduped) == 1 {
		return deduped[0], nil
	}

	return c.SetXor(deduped...)
}

func (c *RulesetCreator) setSingleOrAnd(ids ...string) (string, error) {
	if len(ids) == 0 {
		return "", errors.Errorf(
			"%w: at least one id is required",
			puanerror.InvalidArgument,
		)
	}

	deduped := utils.Dedupe(ids)
	if len(deduped) == 1 {
		return deduped[0], nil
	}

	return c.SetAnd(deduped...)
}

func (c *RulesetCreator) createAssumeConstraints() error {
	if len(c.assumedVariables) == 0 {
		return nil
	}

	root, err := c.setSingleOrAnd(c.assumedVariables...)
	if err != nil {
		return err
	}

	return c.model.Assume(root)
}
