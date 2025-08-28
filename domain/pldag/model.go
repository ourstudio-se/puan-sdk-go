package pldag

import (
	"crypto/sha1"
	"fmt"
	"maps"
	"math"
	"slices"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/utils"
)

type Polyhedron struct {
	aMatrix [][]int
	bVector []int
}

type SparseMatrix struct {
	rows    []int
	columns []int
	values  []int
	shape   Shape
}

func (s SparseMatrix) Rows() []int {
	return s.rows
}

func (s SparseMatrix) Columns() []int {
	return s.columns
}

func (s SparseMatrix) Values() []int {
	return s.values
}

func (s SparseMatrix) Shape() Shape {
	return s.shape
}

func (p Polyhedron) SparseMatrix() SparseMatrix {
	var row []int
	var column []int
	var value []int

	for rowIndex := range p.aMatrix {
		for columIndex := range p.aMatrix[rowIndex] {
			if p.aMatrix[rowIndex][columIndex] != 0 {
				row = append(row, rowIndex)
				column = append(column, columIndex)
				value = append(value, p.aMatrix[rowIndex][columIndex])
			}
		}
	}

	return NewSparseMatrix(row, column, value, p.shape())
}

func NewSparseMatrix(rows, columns, values []int, shape Shape) SparseMatrix {
	return SparseMatrix{
		rows:    rows,
		columns: columns,
		values:  values,
		shape:   shape,
	}
}

type Shape struct {
	nrOfRows, nrOfColumns int
}

func NewShape(rows, columns int) Shape {
	return Shape{
		nrOfRows:    rows,
		nrOfColumns: columns,
	}
}

func (s Shape) NrOfRows() int {
	return s.nrOfRows
}

func (s Shape) NrOfColumns() int {
	return s.nrOfColumns
}

func (p Polyhedron) shape() Shape {
	if len(p.aMatrix) == 0 {
		return Shape{}
	}

	nrOfRows := len(p.aMatrix)
	nrOfColumns := len(p.aMatrix[0])

	return NewShape(nrOfRows, nrOfColumns)
}

func (p Polyhedron) A() [][]int {
	return p.aMatrix
}

func (p Polyhedron) B() []int {
	return p.bVector
}

type Bias int

func (b Bias) negate() Bias {
	return -b - 1
}

type coefficientValues map[string]int

func (c coefficientValues) negate() coefficientValues {
	negated := make(map[string]int, len(c))
	for key, value := range c {
		negated[key] = -value
	}

	return negated
}

func (c coefficientValues) calculateMaxAbsInnerBound() int {
	sumNegatives, sumPositives := 0, 0
	for _, value := range c {
		if value < 0 {
			sumNegatives += value
		}
		if value > 0 {
			sumPositives += value
		}
	}

	absSumNegatives := math.Abs(float64(sumNegatives))
	absSumPositives := math.Abs(float64(sumPositives))
	maxValue := math.Max(absSumNegatives, absSumPositives)

	return int(maxValue)
}

type (
	Model struct {
		variables         []string
		constraints       Constraints
		assumeConstraints AuxiliaryConstraints
	}

	Constraint struct {
		id           string
		coefficients coefficientValues
		bias         Bias
	}

	AuxiliaryConstraint struct {
		coefficients coefficientValues
		bias         Bias
	}

	Constraints          []Constraint
	AuxiliaryConstraints []AuxiliaryConstraint
)

func (m *Model) PrimitiveVariables() []string {
	constraintIDs := make([]string, len(m.constraints))
	for i := range m.constraints {
		constraintIDs[i] = m.constraints[i].id
	}

	primitiveIDs := utils.Without(m.variables, constraintIDs)

	return primitiveIDs
}

func (c AuxiliaryConstraints) coefficientIDs() []string {
	idMap := make(map[string]any)
	for _, constraint := range c {
		for coefficientID := range constraint.coefficients {
			idMap[coefficientID] = nil
		}
	}

	ids := make([]string, len(idMap))
	for id := range idMap {
		ids = append(ids, id)
	}

	return ids
}

func (m *Model) Variables() []string {
	return m.variables
}

func New() *Model {
	return &Model{
		variables:         []string{},
		constraints:       Constraints{},
		assumeConstraints: AuxiliaryConstraints{},
	}
}

func (m *Model) SetPrimitives(primitives ...string) {
	m.variables = append(m.variables, primitives...)
}

func (m *Model) SetAnd(variables ...string) (string, error) {
	id, err := m.setAtLeast(variables, len(variables))
	if err != nil {
		return "", err
	}

	return id, nil
}

func (m *Model) SetOr(variables ...string) (string, error) {
	id, err := m.setAtLeast(variables, 1)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (m *Model) SetNot(variables ...string) (string, error) {
	id, err := m.setAtMost(variables, 0)
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
	atLeastID, err := m.setAtLeast(variables, 1)
	if err != nil {
		return "", err
	}
	atMostID, err := m.setAtMost(variables, 1)
	if err != nil {
		return "", err
	}

	return m.SetAnd([]string{atLeastID, atMostID}...)
}

func (m *Model) SetOneOrNone(variables ...string) (string, error) {
	return m.setAtMost(variables, 1)
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
	err := m.validateAssumedVariables(variables...)
	if err != nil {
		return err
	}

	constraint := m.newAssumedConstraint(variables...)
	m.assumeConstraints = append(m.assumeConstraints, constraint)

	return nil
}

func (m *Model) newAssumedConstraint(variables ...string) AuxiliaryConstraint {
	coefficients := make(coefficientValues, len(variables))
	for _, id := range variables {
		coefficients[id] = -1
	}

	bias := Bias(-len(variables))

	return newAuxiliaryConstraint(coefficients, bias)
}

func (m *Model) validateAssumedVariables(assumedVariables ...string) error {
	existingAssumedVariables := m.assumeConstraints.coefficientIDs()
	seen := make(map[string]any)
	for _, v := range assumedVariables {
		if _, ok := seen[v]; ok {
			return errors.New("duplicated variable")
		}
		seen[v] = nil

		if slices.Contains(existingAssumedVariables, v) {
			return errors.New("variable already assumed")
		}
		if !slices.Contains(m.variables, v) {
			return errors.New("variable not in model")
		}
	}

	return nil
}

func (m *Model) GeneratePolyhedron() Polyhedron {
	var aMatrix [][]int
	var bVector []int

	constraintsWithSupport := m.toAuxiliaryConstraintsWithSupport()
	var constraintsInMatrix AuxiliaryConstraints
	constraintsInMatrix = append(constraintsInMatrix, constraintsWithSupport...)
	constraintsInMatrix = append(constraintsInMatrix, m.assumeConstraints...)
	for _, c := range constraintsInMatrix {
		row := c.asMatrixRow(m.variables)
		bias := int(c.bias)
		aMatrix = append(aMatrix, row)
		bVector = append(bVector, bias)
	}

	return Polyhedron{aMatrix, bVector}
}

func (m *Model) toAuxiliaryConstraintsWithSupport() AuxiliaryConstraints {
	var constraints AuxiliaryConstraints
	for _, c := range m.constraints {
		supportImpliesConstraint, constraintImpliesSupport := c.toAuxiliaryConstraintsWithSupport()
		constraints = append(constraints, supportImpliesConstraint)
		constraints = append(constraints, constraintImpliesSupport)
	}

	return constraints
}

func (c AuxiliaryConstraint) asMatrixRow(variables []string) []int {
	row := make([]int, len(variables))
	for i, id := range variables {
		if value, ok := c.coefficients[id]; ok {
			row[i] = value
		} else {
			row[i] = 0
		}
	}

	return row
}

func (c Constraint) toAuxiliaryConstraintsWithSupport() (AuxiliaryConstraint, AuxiliaryConstraint) {
	supportImpliesConstraint := c.newSupportImpliesConstraint()
	constraintImpliesSupport := c.newConstraintImpliesSupport()

	return supportImpliesConstraint, constraintImpliesSupport
}

func (c Constraint) newConstraintImpliesSupport() AuxiliaryConstraint {
	negatedCoefficients := c.coefficients.negate()
	innerBound := negatedCoefficients.calculateMaxAbsInnerBound()
	negatedBias := c.bias.negate()

	newCoefficients := make(coefficientValues, len(c.coefficients)+1)
	for coefficientID, value := range negatedCoefficients {
		newCoefficients[coefficientID] = value
	}

	newCoefficients[c.id] = int(negatedBias) - innerBound

	return AuxiliaryConstraint{
		coefficients: newCoefficients,
		bias:         negatedBias,
	}
}

func (c Constraint) newSupportImpliesConstraint() AuxiliaryConstraint {
	innerBound := c.coefficients.calculateMaxAbsInnerBound()
	bias := Bias(int(c.bias) + innerBound)

	newCoefficients := make(coefficientValues, len(c.coefficients)+1)
	for coefficientID, value := range c.coefficients {
		newCoefficients[coefficientID] = value
	}

	newCoefficients[c.id] = innerBound

	return AuxiliaryConstraint{
		coefficients: newCoefficients,
		bias:         bias,
	}
}

func (m *Model) setAtLeast(variables []string, amount int) (string, error) {
	constraint, err := newAtLeastConstraint(variables, amount)
	if err != nil {
		return "", err
	}

	m.setConstraint(constraint)

	return constraint.id, nil
}

func (m *Model) setAtMost(variables []string, amount int) (string, error) {
	constraint, err := newAtMostConstraint(variables, amount)
	if err != nil {
		return "", err
	}
	m.setConstraint(constraint)

	return constraint.id, nil
}

func newAtMostConstraint(variables []string, amount int) (Constraint, error) {
	if containsDuplicates(variables) {
		return Constraint{}, errors.New("duplicated variables")
	}

	if amount > len(variables) {
		return Constraint{}, errors.New("amount cannot be greater than number of variables")
	}

	if amount < 0 {
		return Constraint{}, errors.New("amount cannot be negative")
	}

	coefficients := make(coefficientValues)
	for _, v := range variables {
		coefficients[v] = 1
	}

	bias := Bias(amount)

	constraint := newConstraint(coefficients, bias)

	return constraint, nil
}

func newAtLeastConstraint(variables []string, amount int) (Constraint, error) {
	if containsDuplicates(variables) {
		return Constraint{}, errors.New("duplicated variables")
	}

	if amount > len(variables) {
		return Constraint{}, errors.New("amount cannot be greater than number of variables")
	}

	if amount < 0 {
		return Constraint{}, errors.New("amount cannot be negative")
	}

	coefficients := make(coefficientValues)
	for _, v := range variables {
		coefficients[v] = -1
	}

	bias := Bias(-amount)

	constraint := newConstraint(coefficients, bias)

	return constraint, nil
}

func containsDuplicates[T comparable](elements []T) bool {
	seen := make(map[T]any)
	for _, e := range elements {
		if _, ok := seen[e]; ok {
			return true
		}
		seen[e] = nil
	}

	return false
}

func newConstraint(coefficients coefficientValues, bias Bias) Constraint {
	id := newConstraintID(coefficients, bias)
	constraint := Constraint{
		id:           id,
		coefficients: coefficients,
		bias:         bias,
	}
	return constraint
}

func newAuxiliaryConstraint(coefficients coefficientValues, bias Bias) AuxiliaryConstraint {
	constraint := AuxiliaryConstraint{
		coefficients: coefficients,
		bias:         bias,
	}
	return constraint
}

func (m *Model) setConstraint(c Constraint) {
	if m.idAlreadyExists(c.id) {
		return
	}

	m.variables = append(m.variables, c.id)
	m.constraints = append(m.constraints, c)
}

func (m *Model) idAlreadyExists(id string) bool {
	return slices.Contains(m.variables, id)
}

func newConstraintID(coefficients coefficientValues, bias Bias) string {
	keys := slices.Sorted(maps.Keys(coefficients))

	h := sha1.New()
	for _, key := range keys {
		h.Write([]byte(key))
		fmt.Fprintf(h, "%d", coefficients[key])
	}
	fmt.Fprintf(h, "%d", bias)

	return fmt.Sprintf("%x", h.Sum(nil))
}
