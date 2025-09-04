package pldag

type Polyhedron struct {
	aMatrix [][]int
	bVector []int
}

func NewPolyhedron(aMatrix [][]int, bVector []int) *Polyhedron {
	return &Polyhedron{
		aMatrix: aMatrix,
		bVector: bVector,
	}
}

func (p *Polyhedron) IsEmpty() bool {
	return len(p.aMatrix) == 0
}

func (p *Polyhedron) SparseMatrix() SparseMatrix {
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

func (p *Polyhedron) shape() Shape {
	if len(p.aMatrix) == 0 {
		return Shape{}
	}

	nrOfRows := len(p.aMatrix)
	nrOfColumns := len(p.aMatrix[0])

	return NewShape(nrOfRows, nrOfColumns)
}

func (p *Polyhedron) A() [][]int {
	return p.aMatrix
}

func (p *Polyhedron) B() []int {
	return p.bVector
}

func (p *Polyhedron) IncrementMatrixRows() {
	for i := range p.aMatrix {
		p.aMatrix[i] = append(p.aMatrix[i], 0)
	}
}
func (p *Polyhedron) Append(row []int, bias Bias) {
	p.aMatrix = append(p.aMatrix, row)
	p.bVector = append(p.bVector, int(bias))
}
