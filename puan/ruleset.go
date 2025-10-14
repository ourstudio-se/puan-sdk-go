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
	variables            []string
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
		variables:           variables,
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
	return r.variables
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

func (r *RuleSet) NewQuery(input QueryInput) (*Query, IndependentSelections, error) {
	selections := input.Selections

	err := r.validateSelectionIDs(selections.ids())
	if err != nil {
		return nil, nil, err
	}

	dependantSelections, independentSelections := r.categorizeSelections(selections)

	extendedSelections := dependantSelections.modifySelections()
	impactingSelections := getImpactingSelections(extendedSelections)

	specification, err := r.newQuerySpecification(impactingSelections, input.From)
	if err != nil {
		return nil, nil, err
	}

	weights := calculateWeights(
		specification.ruleSet.selectableVariables,
		specification.querySelections,
		specification.ruleSet.preferredVariables,
		specification.ruleSet.periodVariables.ids(),
	)

	query := NewQuery(
		specification.ruleSet.polyhedron,
		specification.ruleSet.variables,
		weights,
	)

	return query, independentSelections, nil
}

func (r *RuleSet) categorizeSelections(selections Selections) (Selections, IndependentSelections, error) {
	err := r.validateSelectionIDs(selections.ids())
	if err != nil {
		return nil, nil, err
	}
	dependantSelections := extractDependantSelections(selections, r.independentVariables)
	independentSelections := extractIndependentSelections(selections, r.independentVariables)

	return dependantSelections, independentSelections, nil
}

func extractDependantSelections(selections Selections, freeVariables []string) Selections {
	var newSelections Selections
	for _, selection := range selections {
		s := extractDependantSelection(selection, freeVariables)
		if s != nil {
			newSelections = append(newSelections, *s)
		}
	}

	return newSelections
}

func extractDependantSelection(selection Selection, freeVariables []string) (*Selection, error) {
	if utils.ContainsAny(selection.subSelectionIDs, freeVariables) {
		return nil, errors.Errorf(
			"cannot have independent variables in composite selection: %v",
			selection,
		)
	}

	if utils.Contains(freeVariables, selection.id) {
		return nil, nil
	}

	return &selection, nil
}

func extractIndependentSelections(selections Selections, independentVariables []string) IndependentSelections {
	var independentSelections IndependentSelections
	for _, variable := range independentVariables {
		selection := extractIndependentSelection(selections, variable)
		if selection != nil {
			independentSelections = append(independentSelections, *selection)
		}
	}

	return independentSelections
}

func extractIndependentSelection(selections Selections, freeVariable string) *IndependentSelection {
	// reverse loop for prioritizing the latest selection action
	for i := len(selections) - 1; i >= 0; i-- {
		selection := selections[i]
		if utils.Contains(selection.ids(), freeVariable) {
			independentSelection := selection.toIndependentSelection()
			return &independentSelection
		}
	}

	return nil
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
		variables:            variableIDs,
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
