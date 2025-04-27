package rfctree

type ArrangementValue struct {
	Candidate map[string]int
	DocId     string
}

type Arrangement struct {
	cell [][]ArrangementValue
	w    int
	h    int
}

func NewArrangement(xmax, ymax int) Arrangement {
	cell := make([][]ArrangementValue, xmax)
	for x := range xmax {
		cell[x] = make([]ArrangementValue, ymax)
	}

	return Arrangement{
		cell: cell,
		w:    xmax,
		h:    ymax,
	}
}

func (a *Arrangement) Cell(p Position) (docId string) {
	return a.cell[p.X][p.Y].DocId
}

func (a *Arrangement) Set(p Position, docId string) {
	a.cell[p.X][p.Y].DocId = docId
}

func (a *Arrangement) IsEmpty(x, y int) bool {
	for px := max(0, x-int(labelSize.W/cellSize.W)); px <= x; px++ {
		if a.cell[px][y].DocId != "" {
			return false
		}
	}
	return true
}

func (a *Arrangement) FindCandidatePositions(x int) []int {
	candidates := make([]int, 0, a.h)

	// Middle positions are high priority.
	for j := range a.h / 2 {
		if a.IsEmpty(x, a.h/2-j-1) {
			candidates = append(candidates, a.h/2-j-1)
		}
		if a.IsEmpty(x, a.h/2+j) {
			candidates = append(candidates, a.h/2+j)
		}
	}
	return candidates
}
