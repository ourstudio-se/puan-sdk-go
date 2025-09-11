package puan

import (
	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/domain/pldag"
	"github.com/ourstudio-se/puan-sdk-go/utils"
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
	err := c.validatePreferredIDs(id)
	if err != nil {
		return err
	}

	c.preferredVariables = append(c.preferredVariables, id...)

	return nil
}

func (c *RuleSetCreator) validatePreferredIDs(ids []string) error {
	if utils.ContainsDuplicates(ids) {
		return errors.New("duplicated preferred variables")
	}

	if utils.ContainsAny(c.preferredVariables, ids) {
		return errors.New("preferred variable already added")
	}

	missingIDs := !utils.ContainsAll(c.pldag.Variables(), ids)
	if missingIDs {
		return errors.New("preferred variable not in model")
	}

	return nil
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
	querySelections querySelections
}

func (r *RuleSet) NewQuery(selections Selections) (*Query, error) {
	// TODO: Refactor this
	moddedSelections := Selections{}
	for _, selection := range selections {
		if selection.isComposite() {
			newS := newSelection(selection.action, selection.id, nil)
			moddedSelections = append(moddedSelections, newS)
		}

		moddedSelections = append(moddedSelections, selection)
	}

	impactingSelections := getImpactingSelections(moddedSelections)
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

func (r *RuleSet) newQuerySelections(selections Selections) (querySelections, error) {
	querySelections := make(querySelections, len(selections))
	for i, selection := range selections {
		querySelection, err := r.newQuerySelection(selection)
		if err != nil {
			return nil, err
		}

		querySelections[i] = querySelection
	}

	return querySelections, nil
}

func (r *RuleSet) newQuerySelection(selection Selection) (querySelection, error) {
	id, err := r.obtainQuerySelectionID(selection)
	if err != nil {
		return querySelection{}, err
	}

	querySelection := querySelection{
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
	return pldag.NewAtLeastConstraint(ids, len(ids))
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
