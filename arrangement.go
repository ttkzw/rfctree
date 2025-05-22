package rfctree

import (
	"cmp"
	"slices"
)

type ArrangementValue struct {
	Candidate map[string]int
	DocId     string
}

type Arrangement struct {
	cell         [][]ArrangementValue
	W            int
	H            int
	YearOfOrigin int
}

func NewArrangement() Arrangement {
	w := 1
	h := 1
	cell := make([][]ArrangementValue, w)
	for x := range w {
		cell[x] = make([]ArrangementValue, h)
	}

	return Arrangement{
		cell: cell,
		W:    w,
		H:    h,
	}
}

func (a *Arrangement) Arrange(rfcIndex *RfcIndex, rfcs []*RfcLabel) {
	a.initializeToTargets(rfcs)

	docIdsOnX := make(map[int][]string, a.W)
	originInMonth := (a.YearOfOrigin - yearOfOrigin) * 12
	for _, rfc := range rfcs {
		rfc.Position.X = rfc.PubDateInMonth - originInMonth
		docIdsOnX[rfc.Position.X] = append(docIdsOnX[rfc.Position.X], rfc.DocId)
	}

	for x, docIds := range docIdsOnX {
		s := 0
		for _, docId := range docIds {
			s = s + rfcIndex.Find(docId).DescendantYRange
		}
		y := a.H/2 + s/2
		for _, docId := range docIds {
			rfc := rfcIndex.Find(docId)
			y = y - rfc.DescendantYRange/2
			if !a.IsEmpty(x, y) {
				y = y - 1
			}
			rfc.Position.Y = y
			a.Set(rfc.Position, rfc.DocId)
		}
	}

	newOrigin, w, h := a.getResizeInfo()
	a.reallocate(rfcIndex, newOrigin, w, h)
}

func (a *Arrangement) initializeToTargets(rfcs []*RfcLabel) {
	minYearRfc := slices.MinFunc(rfcs, func(a, b *RfcLabel) int {
		return cmp.Compare(a.PubYear, b.PubYear)
	})
	minYear := minYearRfc.PubYear

	maxYearRfc := slices.MaxFunc(rfcs, func(a, b *RfcLabel) int {
		return cmp.Compare(a.PubYear, b.PubYear)
	})
	maxYear := maxYearRfc.PubYear

	a.YearOfOrigin = minYear
	a.W = (maxYear - minYear + 2) * 12
	a.H = len(rfcs)
	a.cell = make([][]ArrangementValue, a.W)
	for x := range a.W {
		a.cell[x] = make([]ArrangementValue, a.H)
	}
}

func (a *Arrangement) getResizeInfo() (newOrigin Position, w, h int) {
	ymins := make([]int, 0, a.H)
	ymaxs := make([]int, 0, a.H)
	var x1, x2 int
	for x, arrangementValues := range a.cell {
		var y1, y2 int
		for y := 0; y < a.H; y++ {
			if arrangementValues[y].DocId != "" {
				if y1 == 0 {
					y1 = y
				}
				y2 = y
			}
		}
		if y1 > 0 {
			ymins = append(ymins, y1)
		}
		ymaxs = append(ymaxs, y2)
		if y1 > 0 || y2 > 0 {
			if x1 == 0 {
				x1 = x
			}
			x2 = x
		}
	}

	xmin := (x1 / 12) * 12
	xmax := (x2 / 12) * 12
	ymin := 0
	ymax := 0
	if len(ymins) > 0 {
		ymin = slices.Min(ymins)
	}
	if len(ymaxs) > 0 {
		ymax = slices.Max(ymaxs)
	}

	w = (xmax/12 - xmin/12 + 2) * 12
	h = ymax - ymin + 1
	newOrigin = Position{X: xmin, Y: ymin}
	return newOrigin, w, h
}

func (a *Arrangement) reallocate(rfcIndex *RfcIndex, newOrigin Position, w, h int) {
	cell := make([][]ArrangementValue, w)
	for x := range w {
		cell[x] = make([]ArrangementValue, h)
	}

	for x := range w {
		if a.W <= newOrigin.X+x {
			break
		}
		for y := range h {
			if a.H <= newOrigin.Y+y {
				break
			}
			v := a.cell[newOrigin.X+x][newOrigin.Y+y]
			if v.DocId != "" {
				rfc := rfcIndex.Find(v.DocId)
				rfc.Position.X = x
				rfc.Position.Y = y
				cell[x][y] = v
			}
		}
	}

	a.cell = cell
	a.W = w
	a.H = h
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
