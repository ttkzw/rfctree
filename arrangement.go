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
	cell       [][]ArrangementValue
	w          int
	h          int
	originYear int
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
		w:    w,
		h:    h,
	}
}

func (a *Arrangement) Arrange(rfcIndex *RfcIndex, rfcs []*RfcLabel) {
	yearMin, yearMax := getYearRange(rfcs)
	a.originYear = yearMin
	xmax := (yearMax - yearMin + 2) * 12
	ymax := len(rfcs)
	a.Reallocate(rfcIndex, xmax, ymax, false)

	rfcIdsOnX := make(map[int][]string, xmax)
	originMonth := (yearMin - yearOfOrigin) * 12
	for _, rfc := range rfcs {
		rfc.Position.X = rfc.PubDateInMonth - originMonth
		rfcIdsOnX[rfc.Position.X] = append(rfcIdsOnX[rfc.Position.X], rfc.DocId)
	}

	slices.SortFunc(rfcs, func(a, b *RfcLabel) int {
		return cmp.Compare(a.Position.X, b.Position.X)
	})

	for x, rfcIds := range rfcIdsOnX {
		s := 0
		for _, rfcId := range rfcIds {
			s = s + rfcIndex.Find(rfcId).DescendantYRange
		}
		y := ymax/2 + s/2
		for _, rfcId := range rfcIds {
			rfc := rfcIndex.Find(rfcId)
			y = y - rfc.DescendantYRange/2
			if !a.IsEmpty(x, y) {
				y = y - 1
			}
			rfc.Position.Y = y
			a.Set(rfc.Position, rfc.DocId)
		}
	}
	a.Reallocate(rfcIndex, xmax, ymax, true)
}

func (a *Arrangement) Reallocate(rfcIndex *RfcIndex, w, h int, fit bool) {
	ymins := make([]int, 0, a.h)
	ymaxs := make([]int, 0, a.h)
	var x1, x2 int
	for x, arrangementValues := range a.cell {
		var y1, y2 int
		for y := 0; y < a.h; y++ {
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

	if fit {
		w = (xmax/12 - xmin/12 + 2) * 12
		h = ymax - ymin + 1
	}

	cell := make([][]ArrangementValue, w)
	for x := range w {
		cell[x] = make([]ArrangementValue, h)
	}

	for x := range w {
		if a.w <= xmin+x {
			break
		}
		for y := range h {
			if a.h <= ymin+y {
				break
			}
			v := a.cell[xmin+x][ymin+y]
			if v.DocId != "" {
				rfc := rfcIndex.Find(v.DocId)
				rfc.Position.X = x
				rfc.Position.Y = y
				cell[x][y] = v
			}
		}
	}

	a.cell = cell
	a.w = w
	a.h = h
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

func getYearRange(rfcs []*RfcLabel) (minYear, maxYear int) {
	minYearRfc := slices.MinFunc(rfcs, func(a, b *RfcLabel) int {
		return cmp.Compare(a.PubYear, b.PubYear)
	})
	maxYearRfc := slices.MaxFunc(rfcs, func(a, b *RfcLabel) int {
		return cmp.Compare(a.PubYear, b.PubYear)
	})
	return minYearRfc.PubYear, maxYearRfc.PubYear
}
