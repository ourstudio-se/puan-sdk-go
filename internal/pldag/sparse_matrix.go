package pldag

type SparseMatrix struct {
	rows    []int
	columns []int
	values  []int
	shape   Shape
}

type Shape struct {
	nrOfRows, nrOfColumns int
}

func NewSparseMatrix(rows, columns, values []int, shape Shape) SparseMatrix {
	return SparseMatrix{
		rows:    rows,
		columns: columns,
		values:  values,
		shape:   shape,
	}
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
