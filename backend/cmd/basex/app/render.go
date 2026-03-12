package app

import (
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"image/draw"
	"sort"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/bas-x/basex/assets"
	"github.com/bas-x/basex/services"
)

var (
	uiBackground      = color.RGBA{R: 18, G: 21, B: 31, A: 255}
	panelBackground   = color.RGBA{R: 32, G: 38, B: 52, A: 255}
	panelMuted        = color.RGBA{R: 48, G: 57, B: 77, A: 255}
	panelHighlight    = color.RGBA{R: 67, G: 97, B: 145, A: 255}
	textPrimary       = color.RGBA{R: 235, G: 240, B: 250, A: 255}
	textSecondary     = color.RGBA{R: 172, G: 182, B: 205, A: 255}
	airbaseMarker     = color.RGBA{R: 230, G: 130, B: 66, A: 255}
	airbaseSelected   = color.RGBA{R: 255, G: 215, B: 64, A: 255}
	aircraftMarker    = color.RGBA{R: 123, G: 198, B: 255, A: 255}
	eventBackground   = color.RGBA{R: 24, G: 29, B: 41, A: 255}
	regionStroke      = color.RGBA{R: 17, G: 21, B: 30, A: 255}
	selectedListColor = color.RGBA{R: 84, G: 108, B: 153, A: 255}
)

func (r *Runtime) renderImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: uiBackground}, image.Point{}, draw.Src)

	headerRect := image.Rect(0, 0, width, 42)
	mapRect := mapPanelRect(width, height)
	sidebarRect := image.Rect(width-344, 12, width-12, height-12)

	drawPanel(img, headerRect, panelBackground)
	drawPanel(img, mapRect, panelBackground)
	drawPanel(img, sidebarRect, panelBackground)

	drawText(img, 16, 24, textPrimary, "BASEX LOCAL TESTER")
	drawText(img, 16, 38, textSecondary, "A/V=cycle  arrows=pan  +/-=zoom  P=pause  F=refresh")

	r.drawMap(img, mapRect)

	return img
}

func (r *Runtime) drawMap(img *image.RGBA, rect image.Rectangle) {
	inner := mapInnerRectFromPanel(rect)
	mapSurface := image.NewRGBA(image.Rect(0, 0, max(inner.Dx(), 1), max(inner.Dy(), 1)))
	draw.Draw(mapSurface, mapSurface.Bounds(), &image.Uniform{C: color.RGBA{R: 24, G: 35, B: 50, A: 255}}, image.Point{}, draw.Src)
	proj := newProjection(mapSurface.Bounds(), r.viewport)

	for _, region := range assets.Regions {
		col := regionFillColor(region.Name)
		for _, area := range region.Areas {
			if len(area) < 3 {
				continue
			}
			points := make([]image.Point, 0, len(area))
			for _, pt := range area {
				points = append(points, proj.projectAssetPoint(pt))
			}
			fillPolygon(mapSurface, points, col)
			drawPolygonStroke(mapSurface, points, regionStroke)
		}
	}

	for _, base := range r.state.Airbases {
		pt := proj.projectServicePoint(base.Location)
		col := airbaseMarker
		radius := 5
		if base.ID == r.state.SelectedAirbaseID {
			col = airbaseSelected
			radius = 7
		}
		drawCircle(mapSurface, pt.X, pt.Y, radius, col)
	}

	for _, threat := range r.state.Threats {
		pt, ok := projectedThreatPoint(r.state, proj, threat)
		if !ok {
			continue
		}
		drawCircle(mapSurface, pt.X, pt.Y, 6, color.RGBA{R: 255, G: 80, B: 80, A: 255})
		drawLine(mapSurface, image.Pt(pt.X-6, pt.Y-6), image.Pt(pt.X+6, pt.Y+6), textPrimary)
		drawLine(mapSurface, image.Pt(pt.X-6, pt.Y+6), image.Pt(pt.X+6, pt.Y-6), textPrimary)
		drawThreatLabel(mapSurface, pt, threat.Region)
	}

	for _, aircraft := range r.state.Aircraft {
		if aircraft.Position.X == 0 && aircraft.Position.Y == 0 {
			continue
		}
		pt := proj.projectServicePoint(aircraft.Position)
		if aircraft.TailNumber == r.state.SelectedAircraft {
			drawCircle(mapSurface, pt.X, pt.Y, 5, airbaseSelected)
			drawText(mapSurface, pt.X+7, pt.Y+4, textSecondary, aircraft.TailNumber[:8])
		} else {
			drawCircle(mapSurface, pt.X, pt.Y, 4, aircraftMarker)
		}
	}

	r.drawSelectionOverlay(mapSurface)
	r.drawRecentEventsOverlay(mapSurface)
	drawText(mapSurface, 8, mapSurface.Bounds().Dy()-8, textSecondary, fmt.Sprintf("viewport %.0fx%.0f", r.viewport.Width, r.viewport.Height))
	draw.Draw(img, inner, mapSurface, image.Point{}, draw.Src)
}

func (r *Runtime) drawSelectionOverlay(img *image.RGBA) {
	lines := selectedOverlayLines(r.state)
	if len(lines) == 0 {
		return
	}
	lineHeight := 16
	panelWidth := 260
	panelHeight := 12 + len(lines)*lineHeight
	panel := image.Rect(10, 10, 10+panelWidth, 10+panelHeight)
	drawPanel(img, panel, color.RGBA{R: 20, G: 24, B: 34, A: 220})
	for i, line := range lines {
		col := textSecondary
		if i == 0 {
			col = textPrimary
		}
		drawText(img, panel.Min.X+10, panel.Min.Y+16+i*lineHeight, col, line)
	}
}

func (r *Runtime) drawRecentEventsOverlay(img *image.RGBA) {
	if r == nil || len(r.state.RecentEvents) == 0 {
		return
	}
	lines := min(len(r.state.RecentEvents), 8)
	lineHeight := 16
	panelWidth := 280
	panelHeight := 18 + lines*lineHeight
	panel := image.Rect(10, img.Bounds().Dy()-panelHeight-24, 10+panelWidth, img.Bounds().Dy()-24)
	drawText(img, panel.Min.X+8, panel.Min.Y+14, textPrimary, "Recent Events")
	for i := 0; i < lines; i++ {
		line := "• " + truncateText(r.state.RecentEvents[i], 36)
		drawText(img, panel.Min.X+8, panel.Min.Y+30+i*lineHeight, textSecondary, line)
	}
}

func selectedOverlayLines(state *State) []string {
	if state == nil {
		return nil
	}
	switch state.SelectedKind {
	case "airbase":
		base := selectedAirbase(state)
		if base == nil {
			return nil
		}
		return []string{
			"Selected Airbase",
			"ID: " + base.ID,
			"Region: " + base.Region,
			fmt.Sprintf("Location: %.1f, %.1f", base.Location.X, base.Location.Y),
		}
	case "aircraft":
		aircraft := selectedAircraft(state)
		if aircraft == nil {
			return nil
		}
		lines := []string{
			"Selected Aircraft",
			"Tail: " + aircraft.TailNumber,
			"State: " + aircraft.State,
		}
		if aircraft.AssignedTo != nil {
			lines = append(lines, "Assigned: "+*aircraft.AssignedTo)
		}
		for _, need := range aircraft.Needs {
			lines = append(lines, fmt.Sprintf("Need: %s (%d)", need.Type, need.Severity))
			if len(lines) >= 7 {
				break
			}
		}
		return lines
	default:
		return nil
	}
}

func projectedThreatPoint(state *State, proj projection, threat services.Threat) (image.Point, bool) {
	for _, base := range state.Airbases {
		if base.RegionID == threat.RegionID {
			return proj.projectServicePoint(base.Location), true
		}
	}
	if bounds, ok := assets.GetRegionBounds(threat.Region); ok {
		center := assets.Point{X: bounds.Min.X + bounds.Width/2, Y: bounds.Min.Y + bounds.Height/2}
		return proj.projectAssetPoint(center), true
	}
	return image.Point{}, false
}

func drawThreatLabel(img *image.RGBA, anchor image.Point, region string) {
	label := truncateText("Threat: "+region, 18)
	width := 8 + len(label)*7
	rect := image.Rect(anchor.X+10, anchor.Y-18, anchor.X+10+width, anchor.Y+2)
	if rect.Max.X >= img.Bounds().Max.X {
		shift := rect.Max.X - img.Bounds().Max.X + 1
		rect = rect.Add(image.Pt(-shift, 0))
	}
	if rect.Min.Y < 0 {
		rect = rect.Add(image.Pt(0, -rect.Min.Y))
	}
	drawPanel(img, rect, color.RGBA{R: 40, G: 20, B: 24, A: 220})
	drawText(img, rect.Min.X+4, rect.Min.Y+14, textPrimary, label)
}

type projection struct {
	viewport Viewport
	width    int
	height   int
	scale    float64
	offsetX  float64
	offsetY  float64
}

func newProjection(surface image.Rectangle, viewport Viewport) projection {
	view := viewport
	if view.Width <= 0 || view.Height <= 0 {
		view = DefaultViewport()
	}
	width := max(surface.Dx(), 1)
	height := max(surface.Dy(), 1)
	scale := minf(float64(width)/maxf(view.Width, 1), float64(height)/maxf(view.Height, 1))
	renderedWidth := view.Width * scale
	renderedHeight := view.Height * scale
	return projection{
		viewport: view,
		width:    width,
		height:   height,
		scale:    scale,
		offsetX:  (float64(width) - renderedWidth) / 2,
		offsetY:  (float64(height) - renderedHeight) / 2,
	}
}

func (p projection) relativePoint(point image.Point) (float64, float64) {
	if p.scale <= 0 || p.viewport.Width <= 0 || p.viewport.Height <= 0 {
		return 0.5, 0.5
	}
	relX := (float64(point.X) - p.offsetX) / (p.scale * p.viewport.Width)
	relY := (float64(point.Y) - p.offsetY) / (p.scale * p.viewport.Height)
	return relX, relY
}

func (p projection) projectAssetPoint(point assets.Point) image.Point {
	return p.project(point.X, point.Y)
}

func (p projection) projectServicePoint(point services.Point) image.Point {
	return p.project(point.X, point.Y)
}

func (p projection) project(x, y float64) image.Point {
	px := p.offsetX + (x-p.viewport.MinX)*p.scale
	py := p.offsetY + (y-p.viewport.MinY)*p.scale
	ix := clampInt(int(px), 0, p.width-1)
	iy := clampInt(int(py), 0, p.height-1)
	return image.Pt(ix, iy)
}

func drawPanel(img *image.RGBA, rect image.Rectangle, col color.Color) {
	if rect.Empty() {
		return
	}
	draw.Draw(img, rect, &image.Uniform{C: col}, image.Point{}, draw.Src)
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func insetRect(rect image.Rectangle, inset int) image.Rectangle {
	return image.Rect(rect.Min.X+inset, rect.Min.Y+inset, rect.Max.X-inset, rect.Max.Y-inset)
}

func mapPanelRect(width, height int) image.Rectangle {
	return image.Rect(12, 54, width-356, height-12)
}

func mapInnerRect(width, height int) image.Rectangle {
	return mapInnerRectFromPanel(mapPanelRect(width, height))
}

func mapInnerRectFromPanel(panel image.Rectangle) image.Rectangle {
	return insetRect(panel, 10)
}

func drawText(img *image.RGBA, x, y int, col color.Color, text string) {
	if text == "" {
		return
	}
	d := font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(text)
}

func fillPolygon(img *image.RGBA, points []image.Point, col color.Color) {
	if len(points) < 3 {
		return
	}
	minY := points[0].Y
	maxY := points[0].Y
	for _, p := range points[1:] {
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	if minY > img.Bounds().Max.Y || maxY < img.Bounds().Min.Y {
		return
	}
	if minY < img.Bounds().Min.Y {
		minY = img.Bounds().Min.Y
	}
	if maxY >= img.Bounds().Max.Y {
		maxY = img.Bounds().Max.Y - 1
	}
	for y := minY; y <= maxY; y++ {
		var intersections []int
		for i := range points {
			j := (i + 1) % len(points)
			p1 := points[i]
			p2 := points[j]
			if p1.Y == p2.Y {
				continue
			}
			if (p1.Y <= y && p2.Y > y) || (p2.Y <= y && p1.Y > y) {
				rx := float64(p2.X-p1.X) * float64(y-p1.Y) / float64(p2.Y-p1.Y)
				x := int(float64(p1.X) + rx)
				intersections = append(intersections, x)
			}
		}
		if len(intersections) < 2 {
			continue
		}
		sort.Ints(intersections)
		for i := 0; i+1 < len(intersections); i += 2 {
			x1, x2 := intersections[i], intersections[i+1]
			if x1 > x2 {
				x1, x2 = x2, x1
			}
			if x1 < img.Bounds().Min.X {
				x1 = img.Bounds().Min.X
			}
			if x2 >= img.Bounds().Max.X {
				x2 = img.Bounds().Max.X - 1
			}
			for x := x1; x <= x2; x++ {
				img.Set(x, y, col)
			}
		}
	}
}

func drawPolygonStroke(img *image.RGBA, points []image.Point, col color.Color) {
	for i := range points {
		j := (i + 1) % len(points)
		drawLine(img, points[i], points[j], col)
	}
}

func drawLine(img *image.RGBA, from, to image.Point, col color.Color) {
	dx := to.X - from.X
	dy := to.Y - from.Y
	steps := max(abs(dx), abs(dy), 1)
	for i := 0; i <= steps; i++ {
		x := from.X + dx*i/steps
		y := from.Y + dy*i/steps
		if image.Pt(x, y).In(img.Bounds()) {
			img.Set(x, y, col)
		}
	}
}

func drawCircle(img *image.RGBA, cx, cy, radius int, col color.Color) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy > radius*radius {
				continue
			}
			x := cx + dx
			y := cy + dy
			if image.Pt(x, y).In(img.Bounds()) {
				img.Set(x, y, col)
			}
		}
	}
}

func distanceSquared(a, b image.Point) int {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return dx*dx + dy*dy
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func minf(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

func regionFillColor(name string) color.Color {
	h := fnv.New32a()
	_, _ = h.Write([]byte(name))
	v := h.Sum32()
	return color.RGBA{
		R: uint8(45 + (v>>16)&0x5F),
		G: uint8(60 + (v>>8)&0x5F),
		B: uint8(80 + v&0x5F),
		A: 220,
	}
}
