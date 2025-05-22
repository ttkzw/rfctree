package rfctree

import (
	"image/color"
	"strconv"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
)

var cellSize = canvas.Size{W: 2.0, H: 20.0}
var labelSize = canvas.Size{W: 23.0, H: 14.0}

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
	// for _, docIdA := range slices.Sorted(maps.Keys(rfcIndex.relationScoreMap)) {
	// 	fmt.Printf("%s\n", docIdA)
	// 	for _, docIdB := range slices.Sorted(maps.Keys(rfcIndex.relationScoreMap[docIdA])) {
	// 		fmt.Printf("    %s: %d\n", docIdB, rfcIndex.relationScoreMap[docIdA][docIdB])
	// 	}
	// }

	targetRfcs := rfcIndex.GetTargets()

	arrangement := NewArrangement()
	arrangement.Arrange(rfcIndex, targetRfcs)

	rfcCanvas := canvas.New(float64(arrangement.W)*cellSize.W+canvasMargin*2, float64(arrangement.H)*cellSize.H+canvasMargin*2)
	ctx := canvas.NewContext(rfcCanvas)
	drawBackground(rfcCanvas, ctx)

	fontFamily := canvas.NewFontFamily("Arial")
	if err := fontFamily.LoadSystemFont("Arial", canvas.FontRegular); err != nil {
		panic(err)
	}
	axisFace := fontFamily.Face(axisFontSize, canvas.Black, canvas.FontNormal)
	drawAxisLine(ctx, axisFace, arrangement)
	drawGrid(ctx, arrangement)

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

func drawAxisLine(ctx *canvas.Context, face *canvas.FontFace, arrangement Arrangement) {
	ctx.SetStrokeColor(canvas.Gray)
	ctx.SetStrokeWidth(0.5)

	// bottom line
	{
		left := positionToCoordinate(Position{X: 0, Y: 0})
		right := positionToCoordinate(Position{X: arrangement.W, Y: 0})
		ctx.MoveTo(left.X, left.Y-cellSize.H/2)
		ctx.LineTo(right.X, right.Y-cellSize.H/2)
		ctx.Stroke()
	}

	// top line
	{
		left := positionToCoordinate(Position{X: 0, Y: arrangement.H - 1})
		right := positionToCoordinate(Position{X: arrangement.W, Y: arrangement.H - 1})
		ctx.MoveTo(left.X, left.Y+cellSize.H/2)
		ctx.LineTo(right.X, right.Y+cellSize.H/2)
		ctx.Stroke()
	}

	// y-axis year label
	for i := range int(arrangement.W / 12) {
		bottom := positionToCoordinate(Position{X: i * 12, Y: 0})
		top := positionToCoordinate(Position{X: i * 12, Y: arrangement.H - 1})
		yearLabel := canvas.NewTextLine(face, strconv.Itoa(arrangement.YearOfOrigin+i), canvas.Middle)
		ctx.DrawText(bottom.X+cellSize.W*6, bottom.Y-cellSize.H/2-axisFontSize/2, yearLabel)
		ctx.DrawText(top.X+cellSize.W*6, top.Y+cellSize.H/2+axisFontSize/2, yearLabel)
	}

}

func drawGrid(ctx *canvas.Context, arrangement Arrangement) {
	fontFamily := canvas.NewFontFamily("Arial")
	if err := fontFamily.LoadSystemFont("Arial", canvas.FontRegular); err != nil {
		panic(err)
	}
	if err := fontFamily.LoadSystemFont("Arial", canvas.FontBold); err != nil {
		panic(err)
	}
	face := fontFamily.Face(axisFontSize, canvas.Grey, canvas.FontBold, canvas.FontNormal)

	ctx.SetStrokeColor(canvas.Gray)
	ctx.SetStrokeWidth(0.4)

	// x grid line
	for i := range int(arrangement.W/12) + 1 {
		bottom := positionToCoordinate(Position{X: i * 12, Y: 0})
		top := positionToCoordinate(Position{X: i * 12, Y: arrangement.H - 1})
		ctx.MoveTo(bottom.X, bottom.Y-cellSize.H/2)
		ctx.LineTo(top.X, top.Y+cellSize.H/2)
		ctx.Stroke()
	}

	// y grid line (for debug)
	ctx.SetStrokeColor(canvas.Lightgray)
	ctx.SetStrokeWidth(0.05)
	face = fontFamily.Face(axisFontSize, canvas.Lightgray, canvas.FontBold, canvas.FontNormal)

	for i := range arrangement.H {
		left := positionToCoordinate(Position{X: 0, Y: i})
		right := positionToCoordinate(Position{X: arrangement.W, Y: i})
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
	ctx.SetStrokeWidth(0.8)

	source := positionToCoordinate(sourceRfc.Position)
	dest := positionToCoordinate(destRfc.Position)

	polyline := canvas.Polyline{}
	endingMarker := polyline.Add(0, 0).Add(-1.0, -0.5).Add(-1.0, 0.5).Close().ToPath()
	line := canvas.Line(dest.X-(source.X+labelSize.W), dest.Y-source.Y)
	markers := line.Markers(nil, nil, endingMarker, true)
	ctx.DrawPath(source.X+labelSize.W, source.Y, line, markers[0])
}
