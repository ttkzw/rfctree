package rfctree

import (
	"cmp"
	"image/color"
	"slices"
	"strconv"
	"time"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
)

var cellSize = canvas.Size{W: 2.0, H: 20.0}
var labelSize = canvas.Size{W: 24.0, H: 14.0}

const (
	axisFontSize  = 12.0
	labelFontSize = 5.0
)

const (
	labelPadding = 1.0
	canvasMargin = 20.0
)

var (
	updatedLineColor   = color.RGBA{0, 0, 223, 255}
	obsoletedLineColor = color.RGBA{87, 50, 14, 255}
)

func CreateDiagram(rfcsDir string, targets []*Target, keywords []string, follow bool, excludes []string, outputFilename string) error {
	rfcIndex, err := NewRfcIndex(rfcsDir, targets, keywords, follow, excludes)
	if err != nil {
		return err
	}

	// for debug
	// for _, rfcIdA := range slices.Sorted(maps.Keys(rfcIndex.relationMap)) {
	// 	fmt.Printf("%s\n", rfcIdA)
	// 	for _, rfcIdB := range slices.Sorted(maps.Keys(rfcIndex.relationMap[rfcIdA])) {
	// 		fmt.Printf("    %s: %f\n", rfcIdB, rfcIndex.relationMap[rfcIdA][rfcIdB])
	// 	}
	// }

	targetRfcs := rfcIndex.GetTargets()

	yearMin, yearMax := getYearRange(targetRfcs)
	xmax := (yearMax - yearMin + 2) * 12
	ymax := getArrangementCapacity(targetRfcs, yearMin, yearMax)
	rfcCanvas := canvas.New(float64(xmax)*cellSize.W+canvasMargin*2, float64(ymax)*cellSize.H+canvasMargin*2)
	ctx := canvas.NewContext(rfcCanvas)
	drawBackground(rfcCanvas, ctx)

	fontFamily := canvas.NewFontFamily("Arial")
	if err := fontFamily.LoadSystemFont("Arial", canvas.FontRegular); err != nil {
		panic(err)
	}
	axisFace := fontFamily.Face(axisFontSize, canvas.Black, canvas.FontNormal)
	drawAxisLine(ctx, axisFace, yearMin, xmax, ymax)
	drawGrid(ctx, yearMin, xmax, ymax)

	arrangePosition(rfcIndex, targetRfcs, yearMin, xmax, ymax)

	for _, rfc := range targetRfcs {
		drawRelationLines(ctx, rfcIndex, rfc)
	}

	whiteLabelFace := fontFamily.Face(labelFontSize, canvas.White, canvas.FontNormal)
	blackLabelFace := fontFamily.Face(labelFontSize, canvas.Black, canvas.FontNormal)
	for _, rfc := range targetRfcs {
		if !rfc.IsTarget {
			continue
		}
		face := whiteLabelFace
		if rfc.Status == UNKNOWN {
			face = blackLabelFace
		}
		drawRfcLabel(ctx, face, rfc)
	}

	if err := renderers.Write(outputFilename, rfcCanvas, canvas.DPMM(10.0)); err != nil {
		panic(err)
	}

	return nil
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

func getArrangementCapacity(rfcs []*RfcLabel, minYear, maxYear int) int {
	const (
		arrangementRange = 2
		ratio            = 2
	)

	numOfyear := make([]int, maxYear-minYear+arrangementRange+1)
	for _, rfc := range rfcs {
		for i := range arrangementRange {
			numOfyear[rfc.PubYear-minYear+i]++
		}
	}
	return slices.Max(numOfyear) * ratio
}

func arrangePosition(rfcIndex *RfcIndex, rfcs []*RfcLabel, yearMin, xmax, ymax int) {
	cell := make([][]string, xmax)
	for x := range xmax {
		cell[x] = make([]string, ymax)
	}

	rfcIdsOnX := make(map[int][]string, xmax)
	originMonth := (yearMin - yearOfOrigin) * 12
	for _, rfc := range rfcs {
		rfc.Position.X = rfc.PubDateInMonth - originMonth
		rfcIdsOnX[rfc.Position.X] = append(rfcIdsOnX[rfc.Position.X], rfc.DocId)
	}

	slices.SortFunc(rfcs, func(a, b *RfcLabel) int {
		return cmp.Compare(a.Position.X, b.Position.X)
	})

	arrangement := NewArrangement(xmax, ymax)
	for x, rfcIds := range rfcIdsOnX {
		s := 0
		for _, rfcId := range rfcIds {
			s = s + rfcIndex.Find(rfcId).DescendantYRange
		}
		y := ymax/2 + s/2
		for _, rfcId := range rfcIds {
			rfc := rfcIndex.Find(rfcId)
			y = y - rfc.DescendantYRange/2
			if !arrangement.IsEmpty(x, y) {
				y = y - 1
			}
			rfc.Position.Y = y
			arrangement.Set(rfc.Position, rfc.DocId)
		}
	}
}

type Coodinate struct {
	X float64
	Y float64
}

func positionToCoordinate(p Position) Coodinate {
	x := float64(p.X)*cellSize.W + canvasMargin
	y := (float64(p.Y)+0.5)*cellSize.H + canvasMargin
	return Coodinate{X: x, Y: y}
}

func drawBackground(c *canvas.Canvas, ctx *canvas.Context) {
	ctx.SetFillColor(color.White)
	ctx.SetStrokeColor(color.White)
	ctx.SetStrokeWidth(0)
	ctx.DrawPath(0, 0, canvas.Rectangle(c.W, c.H))
}

func drawAxisLine(ctx *canvas.Context, face *canvas.FontFace, yearMin, xmax, ymax int) {
	ctx.SetStrokeColor(canvas.Gray)
	ctx.SetStrokeWidth(0.5)

	// bottom line
	{
		left := positionToCoordinate(Position{X: 0, Y: 0})
		right := positionToCoordinate(Position{X: xmax, Y: 0})
		ctx.MoveTo(left.X, left.Y-cellSize.H/2)
		ctx.LineTo(right.X, right.Y-cellSize.H/2)
		ctx.Stroke()
	}

	// top line
	{
		left := positionToCoordinate(Position{X: 0, Y: ymax - 1})
		right := positionToCoordinate(Position{X: xmax, Y: ymax - 1})
		ctx.MoveTo(left.X, left.Y+cellSize.H/2)
		ctx.LineTo(right.X, right.Y+cellSize.H/2)
		ctx.Stroke()
	}

	// y-axis year label
	for i := range int(xmax / 12) {
		bottom := positionToCoordinate(Position{X: i * 12, Y: 0})
		top := positionToCoordinate(Position{X: i * 12, Y: ymax - 1})
		yearLabel := canvas.NewTextLine(face, strconv.Itoa(yearMin+i), canvas.Middle)
		ctx.DrawText(bottom.X+cellSize.W*6, bottom.Y-cellSize.H/2-axisFontSize/2, yearLabel)
		ctx.DrawText(top.X+cellSize.W*6, top.Y+cellSize.H/2+axisFontSize/2, yearLabel)
	}

}

func drawGrid(ctx *canvas.Context, yearMin, xmax, ymax int) {
	fontFamily := canvas.NewFontFamily("Arial")
	if err := fontFamily.LoadSystemFont("Arial", canvas.FontRegular); err != nil {
		panic(err)
	}
	if err := fontFamily.LoadSystemFont("Arial", canvas.FontBold); err != nil {
		panic(err)
	}
	face := fontFamily.Face(axisFontSize, canvas.Grey, canvas.FontBold, canvas.FontNormal)

	ctx.SetStrokeColor(canvas.Gray)
	ctx.SetStrokeWidth(0.5)

	// x grid line
	for i := range int(xmax/12) + 1 {
		bottom := positionToCoordinate(Position{X: i * 12, Y: 0})
		top := positionToCoordinate(Position{X: i * 12, Y: ymax - 1})
		ctx.MoveTo(bottom.X, bottom.Y-cellSize.H/2)
		ctx.LineTo(top.X, top.Y+cellSize.H/2)
		ctx.Stroke()
	}

	ctx.SetStrokeColor(canvas.Grey)
	ctx.SetStrokeWidth(0.1)

	// y grid line (for debug)
	for i := range ymax {
		left := positionToCoordinate(Position{X: 0, Y: i})
		right := positionToCoordinate(Position{X: xmax, Y: i})
		ctx.MoveTo(left.X, left.Y)
		ctx.LineTo(right.X, right.Y)
		ctx.Stroke()

		yLabel := canvas.NewTextLine(face, strconv.Itoa(i), canvas.Left)
		ctx.DrawText(left.X, left.Y, yLabel)
	}
}

func drawRfcLabel(ctx *canvas.Context, face *canvas.FontFace, rfc *RfcLabel) {
	ctx.SetFillColor(rfc.Status.Color())
	ctx.SetStrokeColor(color.Gray{Y: 127})
	ctx.SetStrokeWidth(0.1)

	c := positionToCoordinate(rfc.Position)

	ctx.DrawPath(c.X, c.Y-labelSize.H*0.5, canvas.RoundedRectangle(labelSize.W, labelSize.H, 1))

	textBox := canvas.NewTextBox(face, rfc.String(), labelSize.W-labelPadding*2, labelSize.H-labelPadding*2, canvas.Left, canvas.Top, 0, 0)
	ctx.DrawText(c.X+labelPadding, c.Y+labelSize.H/2-labelPadding, textBox)
}

func drawRelationLines(ctx *canvas.Context, rfcIndex *RfcIndex, rfc *RfcLabel) {
	for _, sourceDocId := range rfc.Updates {
		sourceRfc := rfcIndex.Find(sourceDocId)
		if sourceRfc == nil || !sourceRfc.IsTarget {
			continue
		}
		drawArrowLine(ctx, sourceRfc, rfc, updatedLineColor)
	}

	for _, sourceDocId := range rfc.Obsoletes {
		sourceRfc := rfcIndex.Find(sourceDocId)
		if sourceRfc == nil || !sourceRfc.IsTarget {
			continue
		}
		drawArrowLine(ctx, sourceRfc, rfc, obsoletedLineColor)
	}
}

func drawArrowLine(ctx *canvas.Context, sourceRfc, destRfc *RfcLabel, lineColor color.Color) {
	ctx.SetFillColor(lineColor)
	ctx.SetStrokeColor(lineColor)
	ctx.SetStrokeWidth(1.0)

	source := positionToCoordinate(sourceRfc.Position)
	dest := positionToCoordinate(destRfc.Position)

	polyline := canvas.Polyline{}
	endingMarker := polyline.Add(0, 0).Add(-1.5, -1.0).Add(-1.5, 1.0).Close().ToPath()
	line := canvas.Line(dest.X-(source.X+labelSize.W), dest.Y-source.Y)
	markers := line.Markers(nil, nil, endingMarker, true)
	if len(markers) == 0 {
		time.Sleep(time.Millisecond * 10)
		// Workaround for bug that causes markers to be empty
	}
	ctx.DrawPath(source.X+labelSize.W, source.Y, line, markers[0])
}
