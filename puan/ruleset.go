package puan

import (
	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
)

type RuleSetCreator struct {
	pldag              *pldag.Model
	preferredVariables []string
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

func (c *RuleSetCreator) PLDAG() *pldag.Model {
	return c.pldag
}

func (c *RuleSetCreator) SetPreferreds(id ...string) error {
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

func (c *RuleSetCreator) Create() *RuleSet {
	polyhedron := c.pldag.NewPolyhedron()
	variables := c.pldag.Variables()
	primitiveVariables := c.PLDAG().PrimitiveVariables()

	return &RuleSet{
		polyhedron:         polyhedron,
		primitiveVariables: primitiveVariables,
		variables:          variables,
		preferredVariables: c.preferredVariables,
	}
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
