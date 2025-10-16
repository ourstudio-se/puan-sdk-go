package pldag

import (
	"slices"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
)

type Model struct {
	variables            []string
	constraints          Constraints
	assumeConstraints    AuxiliaryConstraints
	independentVariables map[string]bool
}

func New() *Model {
	return &Model{
		variables:            []string{},
		constraints:          Constraints{},
		assumeConstraints:    AuxiliaryConstraints{},
		independentVariables: map[string]bool{},
	}
}

func (m *Model) AddPrimitives(primitives ...string) error {
	if utils.ContainsDuplicates(primitives) {
		return errors.New("primitives contain duplicates")
	}

	for _, p := range primitives {
		if p == "" {
			return errors.New("primitive cannot be empty")
		}
		if m.idAlreadyExists(p) {
			return errors.Errorf("primitive %s already exists in model", p)
		}

		// By default, primitives are independent variables
		// until they are used in a constraint
		m.independentVariables[p] = true
		m.variables = append(m.variables, p)
	}

	return nil
}

func (m *Model) SetAnd(variables ...string) (string, error) {
	deduped := utils.Dedupe(variables)

	if len(deduped) < 2 {
		return "", errors.Errorf("at least two variables are required, got %v", deduped)
	}

	id, err := m.setAtLeast(deduped, len(deduped))
	if err != nil {
		return "", err
	}

	return id, nil
}

func (m *Model) SetOr(variables ...string) (string, error) {
	deduped := utils.Dedupe(variables)

	if len(deduped) < 2 {
		return "", errors.Errorf("at least two variables are required, got %v", deduped)
	}

	id, err := m.setAtLeast(deduped, 1)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (m *Model) SetNot(variables ...string) (string, error) {
	deduped := utils.Dedupe(variables)
	id, err := m.setAtMost(deduped, 0)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (m *Model) SetImply(condition, consequence string) (string, error) {
	notID, err := m.SetNot(condition)
	if err != nil {
		return "", err
	}

	return m.SetOr([]string{notID, consequence}...)
}

func (m *Model) SetXor(variables ...string) (string, error) {
	deduped := utils.Dedupe(variables)

	if len(deduped) < 2 {
		return "", errors.Errorf("at least two variables are required, got %v", deduped)
	}

	atLeastID, err := m.setAtLeast(deduped, 1)
	if err != nil {
		return "", err
	}

	atMostID, err := m.setAtMost(deduped, 1)
	if err != nil {
		return "", err
	}

	return m.SetAnd([]string{atLeastID, atMostID}...)
}

func (m *Model) SetOneOrNone(variables ...string) (string, error) {
	deduped := utils.Dedupe(variables)

	if len(deduped) < 2 {
		return "", errors.Errorf("at least two variables are required, got %v", deduped)
	}

	return m.setAtMost(deduped, 1)
}

func (m *Model) SetEquivalent(variableOne, variableTwo string) (string, error) {
	andID, err := m.SetAnd(variableOne, variableTwo)
	if err != nil {
		return "", err
	}
	notID, err := m.SetNot(variableOne, variableTwo)
	if err != nil {
		return "", err
	}

	return m.SetOr(andID, notID)
}

func (m *Model) Assume(variables ...string) error {
	deduped := utils.Dedupe(variables)
	err := m.ValidateVariables(deduped...)
	if err != nil {
		return err
	}

	newAssumed := utils.Without(variables, m.assumeConstraints.coefficientIDs())

	constraints := NewAssumedConstraints(newAssumed...)
	m.assumeConstraints = append(m.assumeConstraints, constraints...)

	return nil
}

func (m *Model) AssumedConstraints() AuxiliaryConstraints {
	return m.assumeConstraints
}

func CreatePolyhedron(
	variables []string,
	constraints Constraints,
	assumeConstraints AuxiliaryConstraints,
) *Polyhedron {
	var aMatrix [][]int
	var bVector []int

	constraintsWithSupport := toAuxiliaryConstraintsWithSupport(constraints)
	var constraintsInMatrix AuxiliaryConstraints
	constraintsInMatrix = append(constraintsInMatrix, constraintsWithSupport...)
	constraintsInMatrix = append(constraintsInMatrix, assumeConstraints...)
	for _, c := range constraintsInMatrix {
		row := c.asMatrixRow(variables)
		bias := int(c.bias)
		aMatrix = append(aMatrix, row)
		bVector = append(bVector, bias)
	}

	return NewPolyhedron(aMatrix, bVector)
}

func (m *Model) PrimitiveVariables() []string {
	constraintIDs := make([]string, len(m.constraints))
	for i := range m.constraints {
		constraintIDs[i] = m.constraints[i].id
	}

	primitiveIDs := utils.Without(m.variables, constraintIDs)

	return primitiveIDs
}

func (m *Model) Variables() []string {
	return m.variables
}

func (m *Model) ValidateVariables(variables ...string) error {
	for _, v := range variables {
		if !utils.Contains(m.variables, v) {
			return errors.Errorf("variable %s not in model", v)
		}
	}

	return nil
}

func (m *Model) IndependentVariables() []string {
	var independentVariables []string
	for _, variable := range m.variables {
		if m.independentVariables[variable] {
			independentVariables = append(independentVariables, variable)
		}
	}

	return independentVariables
}

func toAuxiliaryConstraintsWithSupport(constraints Constraints) AuxiliaryConstraints {
	var auxiliaryConstraints AuxiliaryConstraints
	for _, c := range constraints {
		supportImpliesConstraint, constraintImpliesSupport := c.ToAuxiliaryConstraintsWithSupport()
		auxiliaryConstraints = append(auxiliaryConstraints, supportImpliesConstraint)
		auxiliaryConstraints = append(auxiliaryConstraints, constraintImpliesSupport)
	}

	return auxiliaryConstraints
}

func (m *Model) setAtLeast(variables []string, amount int) (string, error) {
	constraint, err := NewAtLeastConstraint(variables, amount)
	if err != nil {
		return "", err
	}

	m.setDependentVariables(variables)
	m.setConstraint(constraint)

	return constraint.id, nil
}

func (m *Model) setAtMost(variables []string, amount int) (string, error) {
	constraint, err := NewAtMostConstraint(variables, amount)
	if err != nil {
		return "", err
	}

	m.setDependentVariables(variables)
	m.setConstraint(constraint)

	return constraint.id, nil
}

func (m *Model) setConstraint(c Constraint) {
	if m.idAlreadyExists(c.id) {
		return
	}

	m.variables = append(m.variables, c.id)
	m.constraints = append(m.constraints, c)
}

func (m *Model) setDependentVariables(variables []string) {
	for _, variable := range variables {
		m.independentVariables[variable] = false
	}
}

func (m *Model) Constraints() Constraints {
	return m.constraints
}

func (m *Model) idAlreadyExists(id string) bool {
	return slices.Contains(m.variables, id)
}
