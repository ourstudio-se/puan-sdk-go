package puan

import (
	"time"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
)

type RuleSet struct {
	polyhedron         *pldag.Polyhedron
	primitiveVariables []string
	variables          []string
	preferredVariables []string
	periodVariables    timeBoundVariables
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

type querySpecification struct {
	ruleSet         *RuleSet
	querySelections QuerySelections
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
