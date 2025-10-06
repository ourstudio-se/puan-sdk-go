package puan

import (
	"fmt"
	"time"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
)

type timeBoundVariables []timeBoundVariable

type timeBoundVariable struct {
	variable string
	period   Period
}

func (p timeBoundVariables) periods() []Period {
	periods := make([]Period, len(p))
	for i, periodVariable := range p {
		periods[i] = periodVariable.period
	}
	return periods
}

func (p timeBoundVariables) ids() []string {
	ids := make([]string, len(p))
	for i, periodVariable := range p {
		ids[i] = periodVariable.variable
	}
	return ids
}

type RuleSetCreator struct {
	pldag              *pldag.Model
	preferredVariables []string
	assumedVariables   []string

	period                    *Period
	timeBoundAssumedVariables timeBoundVariables
}

type RuleSet struct {
	polyhedron         *pldag.Polyhedron
	primitiveVariables []string
	variables          []string
	preferredVariables []string
	periodVariables    timeBoundVariables
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

func (c *RuleSetCreator) Prefer(id ...string) error {
	dedupedIDs := utils.Dedupe(id)
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

func (c *RuleSetCreator) Assume(id ...string) error {
	dedupedIDs := utils.Dedupe(id)
	unassumedIDs := utils.Without(dedupedIDs, c.assumedVariables)

	err := c.pldag.ValidateVariables(unassumedIDs...)
	if err != nil {
		return err
	}

	c.assumedVariables = append(c.assumedVariables, unassumedIDs...)

	return nil
}

func (c *RuleSetCreator) AssumeInPeriod(
	variable string,
	from, to time.Time,
) error {
	period, err := NewPeriod(from, to)
	if err != nil {
		return err
	}

	validityPeriod := timeBoundVariable{
		variable: variable,
		period:   period,
	}
	c.timeBoundAssumedVariables = append(c.timeBoundAssumedVariables, validityPeriod)
	return nil
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

func (c *RuleSetCreator) createAssumedConstraints() error {
	if len(c.assumedVariables) == 0 {
		return nil
	}

	root, err := c.pldag.SetAnd(c.assumedVariables...)
	if err != nil {
		return err
	}

	return c.pldag.Assume(root)
}

func (c *RuleSetCreator) Create() (*RuleSet, error) {
	periodVariables, err := c.createValidityPeriodConstraints()
	if err != nil {
		return nil, err
	}

	err = c.createAssumedConstraints()
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

func (c *RuleSetCreator) createValidityPeriodConstraints() (timeBoundVariables, error) {
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

func (r *RuleSet) Polyhedron() *pldag.Polyhedron {
	return r.polyhedron
}

func (r *RuleSet) PrimitiveVariables() []string {
	return r.primitiveVariables
}

func (r *RuleSet) Variables() []string {
	return r.variables
}

func (r *RuleSet) PreferredVariables() []string {
	return r.preferredVariables
}

func (r *RuleSet) RemoveSupportVariables(solution Solution) (Solution, error) {
	nonSupportVariables := []string{}
	nonSupportVariables = append(nonSupportVariables, r.periodVariables.ids()...)
	nonSupportVariables = append(nonSupportVariables, r.primitiveVariables...)

	return solution.Extract(nonSupportVariables...)
}

type QueryInput struct {
	Selections Selections
	From       *time.Time
}

type querySpecification struct {
	ruleSet         *RuleSet
	querySelections QuerySelections
}

func (r *RuleSet) NewQuery(input QueryInput) (*Query, error) {
	selections := input.Selections

	err := r.validateSelectionIDs(selections.ids())
	if err != nil {
		return nil, err
	}

	extendedSelections := selections.modifySelections()
	impactingSelections := getImpactingSelections(extendedSelections)
	specification, err := r.newQuerySpecification(impactingSelections, input.From)
	if err != nil {
		return nil, err
	}

	weights := calculateWeights(
		specification.ruleSet.primitiveVariables,
		specification.querySelections,
		specification.ruleSet.preferredVariables,
		specification.ruleSet.periodVariables.ids(),
	)

	query := NewQuery(
		specification.ruleSet.polyhedron,
		specification.ruleSet.variables,
		weights,
	)

	return query, nil
}

func (r *RuleSet) validateSelectionIDs(ids []string) error {
	for _, id := range ids {
		if utils.Contains(r.variables, id) {
			continue
		}

		return errors.Errorf("invalid selection id: %s", id)
	}

	return nil
}

func (r *RuleSet) copy() *RuleSet {
	aMatrix := make([][]int, len(r.polyhedron.A()))
	copy(aMatrix, r.polyhedron.A())

	bVector := make([]int, len(r.polyhedron.B()))
	copy(bVector, r.polyhedron.B())

	polyhedron := pldag.NewPolyhedron(aMatrix, bVector)

	variableIDs := make([]string, len(r.variables))
	copy(variableIDs, r.variables)

	primitiveVariables := make([]string, len(r.primitiveVariables))
	copy(primitiveVariables, r.primitiveVariables)

	preferredIDs := make([]string, len(r.preferredVariables))
	copy(preferredIDs, r.preferredVariables)

	periodVariables := make([]timeBoundVariable, len(r.periodVariables))
	copy(periodVariables, r.periodVariables)

	return &RuleSet{
		polyhedron:         polyhedron,
		primitiveVariables: primitiveVariables,
		variables:          variableIDs,
		preferredVariables: preferredIDs,
		periodVariables:    periodVariables,
	}
}

func (r *RuleSet) newQuerySpecification(
	selections Selections,
	from *time.Time,
) (*querySpecification, error) {
	ruleSet := r.copy()

	querySelections, err := ruleSet.newQuerySelections(selections)
	if err != nil {
		return nil, err
	}

	err = ruleSet.forbidPassedPeriods(from)
	if err != nil {
		return nil, err
	}

	return &querySpecification{
		ruleSet:         ruleSet,
		querySelections: querySelections,
	}, nil
}

func (r *RuleSet) newQuerySelections(selections Selections) (QuerySelections, error) {
	querySelections := make(QuerySelections, len(selections))
	for i, selection := range selections {
		querySelection, err := r.newQuerySelection(selection)
		if err != nil {
			return nil, err
		}

		querySelections[i] = querySelection
	}

	return querySelections, nil
}

func (r *RuleSet) newQuerySelection(selection Selection) (QuerySelection, error) {
	id, err := r.obtainQuerySelectionID(selection)
	if err != nil {
		return QuerySelection{}, err
	}

	querySelection := QuerySelection{
		id:     id,
		action: selection.action,
	}
	return querySelection, nil
}

func (r *RuleSet) obtainQuerySelectionID(selection Selection) (string, error) {
	if selection.isComposite() {
		return r.setCompositeSelectionConstraint(selection.ids())
	}

	return selection.id, nil
}

func (r *RuleSet) setCompositeSelectionConstraint(ids []string) (string, error) {
	constraint, err := newCompositeSelectionConstraint(ids)
	if err != nil {
		return "", err
	}

	err = r.setConstraintIfNotExist(constraint)
	if err != nil {
		return "", err
	}

	return constraint.ID(), nil
}

func newCompositeSelectionConstraint(ids []string) (pldag.Constraint, error) {
	dedupedIDs := utils.Dedupe(ids)
	return pldag.NewAtLeastConstraint(dedupedIDs, len(dedupedIDs))
}

func (r *RuleSet) setConstraintIfNotExist(constraint pldag.Constraint) error {
	if r.constraintExists(constraint) {
		return nil
	}

	return r.setConstraint(constraint)
}

func (r *RuleSet) constraintExists(constraint pldag.Constraint) bool {
	return utils.Contains(r.variables, constraint.ID())
}

func (r *RuleSet) setConstraint(constraint pldag.Constraint) error {
	r.polyhedron.AddEmptyColumn()

	r.variables = append(r.variables, constraint.ID())

	supportImpliesConstraint, constraintImpliesSupport :=
		constraint.ToAuxiliaryConstraintsWithSupport()

	err := r.setAuxiliaryConstraint(supportImpliesConstraint)
	if err != nil {
		return err
	}

	err = r.setAuxiliaryConstraint(constraintImpliesSupport)
	if err != nil {
		return err
	}

	return nil
}

func (r *RuleSet) setAuxiliaryConstraint(constraint pldag.AuxiliaryConstraint) error {
	row, err := r.newRow(constraint.Coefficients())
	if err != nil {
		return err
	}

	r.polyhedron.Extend(row, constraint.Bias())

	return nil
}

func (r *RuleSet) newRow(coefficients pldag.Coefficients) ([]int, error) {
	row := make([]int, len(r.variables))

	for id, value := range coefficients {
		idIndex, err := utils.IndexOf(r.variables, id)
		if err != nil {
			return nil, err
		}

		row[idIndex] = value
	}

	return row, nil
}

func (r *RuleSet) forbidPassedPeriods(from *time.Time) error {
	if from == nil {
		return nil
	}

	var passedPeriodIDs []string
	for _, periodVariable := range r.periodVariables {
		if periodVariable.period.to.Before(*from) {
			passedPeriodIDs = append(passedPeriodIDs, periodVariable.variable)
		}
	}

	if len(passedPeriodIDs) == 0 {
		return nil
	}

	passedPeriodsConstraint, err := pldag.NewAtMostConstraint(passedPeriodIDs, 0)
	if err != nil {
		return err
	}
	if err := r.setConstraint(passedPeriodsConstraint); err != nil {
		return err
	}

	if err := r.assumeAuxiliaryConstraint(passedPeriodsConstraint.ID()); err != nil {
		return err
	}

	return nil
}

func (r *RuleSet) assumeAuxiliaryConstraint(id string) error {
	row, err := r.newRow(map[string]int{id: 1})
	if err != nil {
		return err
	}
	r.polyhedron.Extend(row, pldag.Bias(1))

	row2, err := r.newRow(map[string]int{id: -1})
	if err != nil {
		return err
	}

	r.polyhedron.Extend(row2, pldag.Bias(-1))

	return nil
}

// For when creating a rule set from a serialized representation
// When setting up new rule sets, use RuleSetCreator instead
func HydrateRuleSet(
	aMatrix [][]int,
	bVector []int,
	variables []string,
	primitiveVariables []string,
	preferredVariables []string,
) *RuleSet {
	polyhedron := pldag.NewPolyhedron(aMatrix, bVector)
	return &RuleSet{
		polyhedron:         polyhedron,
		primitiveVariables: primitiveVariables,
		variables:          variables,
		preferredVariables: preferredVariables,
	}
}
