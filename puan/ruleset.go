package puan

import (
	"time"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
)

type timeBoundAssumedVariable struct {
	variable string
	period   period
}

type period struct {
	from time.Time
	to   time.Time
}

type RuleSetCreator struct {
	pldag                     *pldag.Model
	preferredVariables        []string
	assumedVariables          []string
	timeBoundAssumedVariables []timeBoundAssumedVariable
}

type RuleSet struct {
	polyhedron         *pldag.Polyhedron
	primitiveVariables []string
	variables          []string
	preferredVariables []string
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
	validityPeriod := timeBoundAssumedVariable{
		variable: variable,
		period: period{
			from: from,
			to:   to,
		},
	}
	c.timeBoundAssumedVariables = append(c.timeBoundAssumedVariables, validityPeriod)
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
	err := c.createAssumedConstraints()
	if err != nil {
		return nil, err
	}

	polyhedron := c.pldag.NewPolyhedron()
	variables := c.pldag.Variables()
	primitiveVariables := c.pldag.PrimitiveVariables()

	return &RuleSet{
		polyhedron:         polyhedron,
		primitiveVariables: primitiveVariables,
		variables:          variables,
		preferredVariables: c.preferredVariables,
	}, nil
}

func (c *RuleSetCreator) createValidityPeriodConstraints() error {
	if len(c.timeBoundAssumedVariables) == 0 {
		return nil
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
	// For each validity period, group them with their period variables
	// Create implies constraint between the period variables (OR) and the variables
	// Create XOR constraint between the period variables

	return nil
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

type querySpecification struct {
	ruleSet         *RuleSet
	querySelections QuerySelections
}

func (r *RuleSet) NewQuery(selections Selections) (*Query, error) {
	err := r.validateSelectionIDs(selections.ids())
	if err != nil {
		return nil, err
	}

	extendedSelections := selections.modifySelections()
	impactingSelections := getImpactingSelections(extendedSelections)
	specification, err := r.newQuerySpecification(impactingSelections)
	if err != nil {
		return nil, err
	}

	weights := calculateWeights(
		specification.ruleSet.primitiveVariables,
		specification.querySelections,
		specification.ruleSet.preferredVariables,
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

	return &RuleSet{
		polyhedron:         polyhedron,
		primitiveVariables: primitiveVariables,
		variables:          variableIDs,
		preferredVariables: preferredIDs,
	}
}

func (r *RuleSet) newQuerySpecification(selections Selections) (*querySpecification, error) {
	ruleSet := r.copy()

	querySelections, err := ruleSet.newQuerySelections(selections)
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
