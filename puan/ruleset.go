package puan

import (
	"time"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
)

type RuleSet struct {
	polyhedron           *pldag.Polyhedron
	selectableVariables  []string
	dependantVariables   []string
	independentVariables []string
	preferredVariables   []string
	periodVariables      timeBoundVariables
}

// For when creating a rule set from a serialized representation
// When setting up new rule sets, use RuleSetCreator instead
func HydrateRuleSet(
	aMatrix [][]int,
	bVector []int,
	variables []string,
	selectableVariables []string,
	preferredVariables []string,
) *RuleSet {
	polyhedron := pldag.NewPolyhedron(aMatrix, bVector)
	return &RuleSet{
		polyhedron:          polyhedron,
		selectableVariables: selectableVariables,
		dependantVariables:  variables,
		preferredVariables:  preferredVariables,
	}
}

func (r *RuleSet) Polyhedron() *pldag.Polyhedron {
	return r.polyhedron
}

func (r *RuleSet) SelectableVariables() []string {
	return r.selectableVariables
}

func (r *RuleSet) Variables() []string {
	return r.dependantVariables
}

func (r *RuleSet) FreeVariables() []string {
	return r.independentVariables
}

func (r *RuleSet) PreferredVariables() []string {
	return r.preferredVariables
}

func (r *RuleSet) RemoveSupportVariables(solution Solution) (Solution, error) {
	nonSupportVariables := []string{}
	nonSupportVariables = append(nonSupportVariables, r.periodVariables.ids()...)
	nonSupportVariables = append(nonSupportVariables, r.selectableVariables...)

	return solution.Extract(nonSupportVariables...)
}

func (r *RuleSet) RemoveAndAddStuff(solution Solution, selections IndependentSelections) (Solution, error) {

	return solution, nil
}

func (r *RuleSet) FindPeriodInSolution(solution Solution) (Period, error) {
	var period *Period
	for _, periodVariable := range r.periodVariables {
		if isSet := solution[periodVariable.variable]; isSet == 1 {
			if period != nil {
				return Period{},
					errors.Errorf(
						"multiple periods found: %v and %v",
						period,
						periodVariable.period,
					)
			}
			period = &periodVariable.period
		}
	}

	if period == nil {
		return Period{}, errors.New("period not found for solution")
	}

	return *period, nil
}

type QueryInput struct {
	Selections Selections
	From       *time.Time
}

func (r *RuleSet) NewQuery(input QueryInput) (*Query, error) {
	specification, err := r.newQuerySpecification(input.Selections, input.From)
	if err != nil {
		return nil, err
	}

	weights := calculateWeights(
		specification.ruleSet.selectableVariables,
		specification.querySelections,
		specification.ruleSet.preferredVariables,
		specification.ruleSet.periodVariables.ids(),
	)

	query := NewQuery(
		specification.ruleSet.polyhedron,
		specification.ruleSet.dependantVariables,
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

	variableIDs := make([]string, len(r.dependantVariables))
	copy(variableIDs, r.dependantVariables)

	freeVariablesIDs := make([]string, len(r.independentVariables))
	copy(freeVariablesIDs, r.independentVariables)

	selectableVariables := make([]string, len(r.selectableVariables))
	copy(selectableVariables, r.selectableVariables)

	preferredIDs := make([]string, len(r.preferredVariables))
	copy(preferredIDs, r.preferredVariables)

	periodVariables := make([]timeBoundVariable, len(r.periodVariables))
	copy(periodVariables, r.periodVariables)

	return &RuleSet{
		polyhedron:           polyhedron,
		selectableVariables:  selectableVariables,
		dependantVariables:   variableIDs,
		independentVariables: freeVariablesIDs,
		preferredVariables:   preferredIDs,
		periodVariables:      periodVariables,
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

	if from != nil {
		err = ruleSet.forbidPassedPeriods(*from)
		if err != nil {
			return nil, err
		}
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
	return utils.Contains(r.dependantVariables, constraint.ID())
}

func (r *RuleSet) setConstraint(constraint pldag.Constraint) error {
	r.polyhedron.AddEmptyColumn()

	r.dependantVariables = append(r.dependantVariables, constraint.ID())

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
	row := make([]int, len(r.dependantVariables))

	for id, value := range coefficients {
		idIndex, err := utils.IndexOf(r.dependantVariables, id)
		if err != nil {
			return nil, err
		}

		row[idIndex] = value
	}

	return row, nil
}

func (r *RuleSet) forbidPassedPeriods(from time.Time) error {
	passedPeriods := r.periodVariables.passed(from)
	passedPeriodIDs := passedPeriods.ids()

	if len(passedPeriodIDs) == 0 {
		return nil
	}

	constraint, err := pldag.NewAtMostConstraint(passedPeriodIDs, 0)
	if err != nil {
		return err
	}

	if err := r.setConstraint(constraint); err != nil {
		return err
	}

	return r.assume(constraint.ID())
}

func (r *RuleSet) assume(id string) error {
	constraints := pldag.NewAssumedConstraints(id)
	for _, constraint := range constraints {
		err := r.setAuxiliaryConstraint(constraint)
		if err != nil {
			return err
		}
	}

	return nil
}
