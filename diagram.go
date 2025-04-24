package rfctree

import (
	"cmp"
	"fmt"
	"image/color"
	"slices"
	"strconv"
	"strings"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
)

var cellSize = canvas.Size{W: 2.0, H: 20.0}
var labelSize = canvas.Size{W: 24.0, H: 14.0}

const (
	axisFontSize  = 10.0
	labelFontSize = 5.0
)

const (
	labelPadding = 1.0
	canvasMargin = 20.0
)

func CreateDiagram(rfcsDir, outputFilename string, keywords []string) error {
	rfcMap, err := ReadRfcJsonFiles(rfcsDir)
	if err != nil {
		return err
	}

	targetDocIds := getTargetDocIds(rfcMap, keywords)

	minYear, maxYear := getYearRange(rfcMap, targetDocIds)
	years := maxYear - minYear + 2
	x := years * 12
	y := getArrangementCapacity(rfcMap, targetDocIds, minYear, maxYear)
	rfcCanvas := canvas.New(float64(x)*cellSize.W+canvasMargin*2, float64(y)*cellSize.H+canvasMargin*2)
	ctx := canvas.NewContext(rfcCanvas)
	drawBackground(rfcCanvas, ctx)

	fmt.Println(rfcCanvas.Size())

	fontFamily := canvas.NewFontFamily("Arial")
	if err := fontFamily.LoadSystemFont("Arial", canvas.FontRegular); err != nil {
		panic(err)
	}
	if err := fontFamily.LoadSystemFont("Arial", canvas.FontBold); err != nil {
		panic(err)
	}

	axisFace := fontFamily.Face(axisFontSize, canvas.Black, canvas.FontBold, canvas.FontNormal)
	drawXAxisLine(ctx, axisFace, years, y)
	drawYAxisLine(ctx, axisFace, minYear, years, y)
	originMonth := (minYear - yearOfOrigin) * 12
	arrangePosition(rfcMap, targetDocIds, originMonth)

	for _, docId := range targetDocIds {
		rfc := rfcMap[docId]
		for _, sourceDocId := range rfc.Updates {
			sourceRfc, ok := rfcMap[sourceDocId]
			if !ok {
				continue
			}
			if !sourceRfc.Metadata.IsTarget {
				continue
			}
			drawLine(ctx, sourceRfc, rfc, true, false)
		}

		for _, sourceDocId := range rfc.Obsoletes {
			sourceRfc, ok := rfcMap[sourceDocId]
			if !ok {
				continue
			}
			if !sourceRfc.Metadata.IsTarget {
				continue
			}
			drawLine(ctx, sourceRfc, rfc, false, true)
		}
	}

	labelFace := fontFamily.Face(labelFontSize, canvas.White, canvas.FontBold, canvas.FontNormal)
	for _, docId := range targetDocIds {
		rfc := rfcMap[docId]
		if !rfc.Metadata.IsTarget {
			continue
		}

		drawLabel(ctx, labelFace, rfc)
	}
	ctx.Close()

	if err := renderers.Write(outputFilename, rfcCanvas, canvas.DPMM(10.0)); err != nil {
		panic(err)
	}

	return nil
}

func getYearRange(rfcMap map[string]*Rfc, rfcDocIds []string) (minYear, maxYear int) {
	minYearDocId := slices.MinFunc(rfcDocIds, func(a, b string) int {
		return cmp.Compare(rfcMap[a].Metadata.PubYear, rfcMap[b].Metadata.PubYear)
	})
	maxYearDocId := slices.MaxFunc(rfcDocIds, func(a, b string) int {
		return cmp.Compare(rfcMap[a].Metadata.PubYear, rfcMap[b].Metadata.PubYear)
	})
	minYear = rfcMap[minYearDocId].Metadata.PubYear
	maxYear = rfcMap[maxYearDocId].Metadata.PubYear
	return minYear, maxYear
}

func getArrangementCapacity(rfcMap map[string]*Rfc, docIds []string, minYear, maxYear int) int {
	const (
		arrangementRange = 5
		ratio            = 2
	)

	numOfyear := make([]int, maxYear-minYear+arrangementRange+1)
	for _, docId := range docIds {
		for i := range arrangementRange {
			numOfyear[rfcMap[docId].Metadata.PubYear-minYear+i]++
		}
	}
	return slices.Max(numOfyear) * ratio
}

func arrangePosition(rfcMap map[string]*Rfc, rfcDocIds []string, originMonth int) {
	y := 1
	for _, docId := range rfcDocIds {
		rfc := rfcMap[docId]
		rfc.Metadata.Position.X = rfc.Metadata.PubDateInMonth - originMonth
		rfc.Metadata.Position.Y = y
		y = y + 1
	}
}

type Coodinate struct {
	X float64
	Y float64
}

func positionToCoordinate(p Position) Coodinate {
	x := float64(p.X)*cellSize.W + canvasMargin
	y := float64(p.Y)*cellSize.H + canvasMargin
	return Coodinate{X: x, Y: y}
}

func drawBackground(c *canvas.Canvas, ctx *canvas.Context) {
	ctx.SetFillColor(color.White)
	ctx.SetStrokeColor(color.White)
	ctx.SetStrokeWidth(0)
	ctx.DrawPath(0, 0, canvas.Rectangle(c.W, c.H))
}

func drawXAxisLine(ctx *canvas.Context, face *canvas.FontFace, elapsedYear, y int) {
	ctx.SetStrokeColor(canvas.Gray)
	ctx.SetStrokeWidth(0.1)

	for i := range y {
		left := positionToCoordinate(Position{X: 0, Y: i})
		right := positionToCoordinate(Position{X: elapsedYear * 12, Y: i})
		ctx.MoveTo(left.X, left.Y)
		ctx.LineTo(right.X, right.Y)
		ctx.Stroke()
	}
}

func drawYAxisLine(ctx *canvas.Context, face *canvas.FontFace, minYear, elapsedYear, y int) {
	ctx.SetStrokeColor(canvas.Gray)
	ctx.SetStrokeWidth(0.5)

	for i := range elapsedYear {
		bottom := positionToCoordinate(Position{X: i * 12, Y: 0})
		top := positionToCoordinate(Position{X: i * 12, Y: y})
		ctx.MoveTo(bottom.X, bottom.Y)
		ctx.LineTo(top.X, top.Y)
		ctx.Stroke()

		ctx.DrawText(bottom.X+2.0, bottom.Y+2.0, canvas.NewTextLine(face, strconv.Itoa(minYear+i), canvas.Left))
		ctx.DrawText(top.X+2.0, top.Y-axisFontSize, canvas.NewTextLine(face, strconv.Itoa(minYear+i), canvas.Left))
	}
}

func drawLabel(ctx *canvas.Context, face *canvas.FontFace, rfc *Rfc) {
	ctx.SetFillColor(rfc.Metadata.Status.Color())
	ctx.SetStrokeColor(color.Gray{Y: 127})
	ctx.SetStrokeWidth(0.1)

	c := positionToCoordinate(rfc.Metadata.Position)
	ctx.DrawPath(c.X, c.Y-labelSize.H*0.5, canvas.RoundedRectangle(labelSize.W, labelSize.H, 1))

	var b strings.Builder
	b.WriteString(rfc.Metadata.DocId)
	b.WriteString(" / ")
	if rfc.Metadata.SubSeries != "" {
		b.WriteString(rfc.Metadata.SubSeries)
	} else {
		b.WriteString(rfc.Metadata.Status.Display())
	}
	b.WriteString("\n")
	b.WriteString(rfc.Title)
	textBox := canvas.NewTextBox(face, b.String(), labelSize.W-labelPadding*2, labelSize.H-labelPadding*2, canvas.Left, canvas.Top, 0, 0)
	ctx.DrawText(c.X+labelPadding, c.Y+labelSize.H/2-labelPadding, textBox)
}

func drawLine(ctx *canvas.Context, sourceRfc, destRfc *Rfc, isUpdate, isObsolete bool) {
	var lineColor color.Color
	if isUpdate {
		lineColor = color.RGBA{0, 0, 223, 255}
	}
	if isObsolete {
		lineColor = color.RGBA{87, 50, 14, 255}
	}
	ctx.SetFillColor(lineColor)
	ctx.SetStrokeColor(lineColor)
	ctx.SetStrokeWidth(1.0)

	source := positionToCoordinate(sourceRfc.Metadata.Position)
	dest := positionToCoordinate(destRfc.Metadata.Position)

	startingMarker := canvas.Circle(0.1)
	polyline := canvas.Polyline{}
	endingMarker := polyline.Add(0, 0).Add(-1.5, -1.0).Add(-1.5, 1.0).Close().ToPath()
	line := canvas.Line(dest.X-(source.X+labelSize.W), dest.Y-source.Y)
	markers := line.Markers(startingMarker, nil, endingMarker, true)
	ctx.DrawPath(source.X+labelSize.W, source.Y, markers[0], line, markers[1])
}
