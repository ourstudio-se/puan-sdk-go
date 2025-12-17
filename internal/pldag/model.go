package pldag

import (
	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
	"github.com/ourstudio-se/puan-sdk-go/puanerror"
)

type Model struct {
	variables         []string
	constraints       Constraints
	assumeConstraints AuxiliaryConstraints
}

func New() *Model {
	return &Model{
		variables:         []string{},
		constraints:       Constraints{},
		assumeConstraints: AuxiliaryConstraints{},
	}
}

func (m *Model) AddPrimitives(primitives ...string) error {
	if utils.ContainsDuplicates(primitives) {
		return errors.Errorf(
			"%w: primitives contains duplicates",
			puanerror.InvalidArgument,
		)
	}

	for _, primitive := range primitives {
		if primitive == "" {
			return errors.Errorf(
				"%w: primitive cannot be empty",
				puanerror.InvalidArgument,
			)
		}

		if m.idAlreadyExists(primitive) {
			return errors.Errorf(
				"%w: primitive %s already exists in model",
				puanerror.InvalidOperation,
				primitive,
			)
		}

		m.variables = append(m.variables, primitive)
	}

	return nil
}

func (m *Model) SetAnd(variables ...string) (string, error) {
	deduped := utils.Dedupe(variables)

	if len(deduped) < 2 {
		return "", errors.Errorf(
			"%w: AND requires at least two variables, got %v",
			puanerror.InvalidArgument,
			deduped,
		)
	}

	id, err := m.setAtLeast(deduped, len(deduped))
	if err != nil {
		return "", errors.WrapPrefix(
			err,
			"AND",
			0,
		)
	}

	return id, nil
}

func (m *Model) SetOr(variables ...string) (string, error) {
	deduped := utils.Dedupe(variables)

	if len(deduped) < 2 {
		return "", errors.Errorf(
			"%w: OR requires at least two variables, got %v",
			puanerror.InvalidArgument,
			deduped,
		)
	}

	id, err := m.setAtLeast(deduped, 1)
	if err != nil {
		return "", errors.WrapPrefix(
			err,
			"OR",
			0,
		)
	}

	return id, nil
}

func (m *Model) SetNot(variables ...string) (string, error) {
	deduped := utils.Dedupe(variables)
	id, err := m.setAtMost(deduped, 0)
	if err != nil {
		return "", errors.WrapPrefix(
			err,
			"NOT",
			0,
		)
	}

	return id, nil
}

func (m *Model) SetImply(condition, consequence string) (string, error) {
	notID, err := m.SetNot(condition)
	if err != nil {
		return "", errors.WrapPrefix(
			err,
			"IMPLY",
			0,
		)
	}

	id, err := m.SetOr([]string{notID, consequence}...)
	if err != nil {
		return "", errors.WrapPrefix(
			err,
			"IMPLY",
			0,
		)
	}

	return id, nil
}

func (m *Model) SetXor(variables ...string) (string, error) {
	deduped := utils.Dedupe(variables)

	if len(deduped) < 2 {
		return "", errors.Errorf(
			"%w: XOR requires at least two variables, got %v",
			puanerror.InvalidArgument,
			deduped,
		)
	}

	atLeastID, err := m.setAtLeast(deduped, 1)
	if err != nil {
		return "", errors.WrapPrefix(
			err,
			"XOR",
			0,
		)
	}

	atMostID, err := m.setAtMost(deduped, 1)
	if err != nil {
		return "", errors.WrapPrefix(
			err,
			"XOR",
			0,
		)
	}

	id, err := m.SetAnd([]string{atLeastID, atMostID}...)
	if err != nil {
		return "", errors.WrapPrefix(
			err,
			"XOR",
			0,
		)
	}

	return id, nil
}

func (m *Model) SetOneOrNone(variables ...string) (string, error) {
	deduped := utils.Dedupe(variables)

	if len(deduped) < 2 {
		return "", errors.Errorf(
			"%w: ONE OR NONE requires at least two variables, got %v",
			puanerror.InvalidArgument,
			deduped,
		)
	}

	id, err := m.setAtMost(deduped, 1)
	if err != nil {
		return "", errors.WrapPrefix(
			err,
			"ONE OR NONE",
			0,
		)
	}

	return id, nil
}

func (m *Model) SetEquivalent(variableOne, variableTwo string) (string, error) {
	andID, err := m.SetAnd(variableOne, variableTwo)
	if err != nil {
		return "", errors.WrapPrefix(
			err,
			"EQUIVALENT",
			0,
		)
	}

	notID, err := m.SetNot(variableOne, variableTwo)
	if err != nil {
		return "", errors.WrapPrefix(
			err,
			"EQUIVALENT",
			0,
		)
	}

	id, err := m.SetOr(andID, notID)
	if err != nil {
		return "", errors.WrapPrefix(
			err,
			"EQUIVALENT",
			0,
		)
	}

	return id, nil
}

func (m *Model) Assume(variables ...string) error {
	deduped := utils.Dedupe(variables)
	err := m.ValidateVariables(deduped...)
	if err != nil {
		return errors.WrapPrefix(
			err,
			"ASSUME",
			0,
		)
	}

	newAssumed := utils.Without(variables, m.assumeConstraints.coefficientIDs())

	constraint := NewAssumedConstraint(newAssumed...)
	m.assumeConstraints = append(m.assumeConstraints, constraint)

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
	for _, variable := range variables {
		if !utils.Contains(m.variables, variable) {
			return errors.Errorf(
				"%w: %s not in model",
				puanerror.InvalidArgument,
				variable,
			)
		}
	}

	return nil
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
	if !utils.ContainsAll(m.variables, variables) {
		return "", errors.Errorf(
			"%w: variables do not exist in model: %v",
			puanerror.InvalidArgument,
			variables,
		)
	}

	constraint, err := NewAtLeastConstraint(variables, amount)
	if err != nil {
		return "", err
	}

	m.setConstraint(constraint)

	return constraint.id, nil
}

func (m *Model) setAtMost(variables []string, amount int) (string, error) {
	if !utils.ContainsAll(m.variables, variables) {
		return "", errors.Errorf(
			"%w: variables do not exist in model: %v",
			puanerror.InvalidArgument,
			variables,
		)
	}

	constraint, err := NewAtMostConstraint(variables, amount)
	if err != nil {
		return "", err
	}

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

func (m *Model) Constraints() Constraints {
	return m.constraints
}

func (m *Model) idAlreadyExists(id string) bool {
	return utils.Contains(m.variables, id)
}
