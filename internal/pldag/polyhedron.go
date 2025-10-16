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

func (p *Polyhedron) A() [][]int {
	return p.aMatrix
}

func (p *Polyhedron) B() []int {
	return p.bVector
}

func (p *Polyhedron) IsEmpty() bool {
	return len(p.aMatrix) == 0
}

func (p *Polyhedron) AddEmptyColumn() {
	for i := range p.aMatrix {
		p.aMatrix[i] = append(p.aMatrix[i], 0)
	}
}

func (p *Polyhedron) Extend(row []int, bias Bias) {
	p.aMatrix = append(p.aMatrix, row)
	p.bVector = append(p.bVector, int(bias))
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

func (p *Polyhedron) shape() Shape {
	if len(p.aMatrix) == 0 {
		return Shape{}
	}

	nrOfRows := len(p.aMatrix)
	nrOfColumns := len(p.aMatrix[0])

	return NewShape(nrOfRows, nrOfColumns)
}
