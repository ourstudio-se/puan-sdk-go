package glpk

type SolveRequest struct {
	Polyhedron Polyhedron  `json:"polyhedron"`
	Objectives []Objective `json:"objectives"`
	Direction  string      `json:"direction"`
}

type Polyhedron struct {
	A         SparseMatrix `json:"A"`
	B         []int        `json:"b"`
	Variables []Variable   `json:"variables"`
}
type SparseMatrix struct {
	Rows  []int `json:"rows"`
	Cols  []int `json:"cols"`
	Vals  []int `json:"vals"`
	Shape Shape `json:"shape"`
}

type Shape struct {
	Nrows int `json:"nrows"`
	Ncols int `json:"ncols"`
}

type Variable struct {
	ID    string `json:"id"`
	Bound [2]int `json:"bound"`
}

type Objective map[string]int

type SolveResponse struct {
	Solutions []Solution `json:"solutions"`
}

type Solution struct {
	Error     *string        `json:"error"`
	Objective float64        `json:"objective"`
	Solution  SolutionValues `json:"solution"`
	Status    string         `json:"status"`
}

type SolutionValues map[string]int
