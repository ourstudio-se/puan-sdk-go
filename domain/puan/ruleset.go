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
	polyhedron := c.pldag.GeneratePolyhedron()
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
	ruleSet           *RuleSet
	selectionIDLookUp map[string]string
}

func (r *RuleSet) NewQuery(selections Selections) (*Query, error) {
	impactingSelections := getImpactingSelections(selections)
	specification, err := r.newQuerySpecification(impactingSelections)
	if err != nil {
		return nil, err
	}

	s, err := getSelections2(impactingSelections, specification.selectionIDLookUp)
	if err != nil {
		return nil, err
	}

	objective := calculateObjective(
		specification.ruleSet.primitiveVariables,
		s,
		specification.ruleSet.preferredVariables,
	)

	query := NewQuery(
		specification.ruleSet.polyhedron,
		specification.ruleSet.variables,
		objective,
	)

	return query, nil
}

func getSelections2(selections Selections, idLookUp map[string]string) (selections2, error) {
	s := make(selections2, len(selections))
	for i, selection := range selections {
		id, ok := idLookUp[selection.Key()]
		if !ok {
			return nil, errors.New("selection not found")
		}

		s[i] = selection2{
			id:     id,
			action: selection.action,
		}
	}

	return s, nil
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

	selectionIDLookUp, err := ruleSet.newIDLookup(selections)
	if err != nil {
		return nil, err
	}

	return &querySpecification{
		ruleSet:           ruleSet,
		selectionIDLookUp: selectionIDLookUp,
	}, nil
}

func (r *RuleSet) newIDLookup(selections Selections) (map[string]string, error) {
	lookup := make(map[string]string)
	for _, selection := range selections {
		id, err := r.obtainSelectionID(selection)
		if err != nil {
			return nil, err
		}

		lookup[selection.Key()] = id
	}

	return lookup, nil
}

func (r *RuleSet) obtainSelectionID(selection Selection) (string, error) {
	id := selection.id
	var err error
	if selection.isComposite() {
		id, err = r.setCompositeSelectionConstraint(selection)
		if err != nil {
			return "", err
		}

		return id, nil
	}

	return id, nil
}

func (r *RuleSet) setCompositeSelectionConstraint(selection Selection) (string, error) {
	return r.setCompositeSelectionAddConstraint(selection.IDs())
}

func (r *RuleSet) setCompositeSelectionRemoveConstraint(ids []string) (string, error) {
	constraint, err := newCompositeSelectionRemoveConstraint(ids)
	if err != nil {
		return "", err
	}

	err = r.setConstraintIfNotExist(constraint)
	if err != nil {
		return "", err
	}

	return constraint.ID(), nil
}

func (r *RuleSet) setCompositeSelectionAddConstraint(ids []string) (string, error) {
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

func newCompositeSelectionRemoveConstraint(ids []string) (pldag.Constraint, error) {
	return pldag.NewAtMostConstraint(ids, 0)
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

func (r *RuleSet) newRow(coefficients pldag.CoefficientValues) ([]int, error) {
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

func (r *RuleSet) setRemoveSelectionConstraint(id string) (string, error) {
	constraint, err := newRemoveSelectionConstraint(id)
	if err != nil {
		return "", err
	}

	err = r.setConstraintIfNotExist(constraint)
	if err != nil {
		return "", err
	}

	return constraint.ID(), nil
}

func newRemoveSelectionConstraint(id string) (pldag.Constraint, error) {
	return pldag.NewAtMostConstraint([]string{id}, 0)
}
