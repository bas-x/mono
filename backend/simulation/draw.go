package simulation

import (
	"hash/fnv"
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/bas-x/basex/assets"
)

const (
	canvasScale   = 2.5
	canvasMargin  = 20
	airbaseRadius = 6
)

var (
	backgroundColor = color.RGBA{R: 15, G: 17, B: 26, A: 255}
	airbaseFill     = color.RGBA{R: 226, G: 115, B: 59, A: 255}
	airbaseStroke   = color.RGBA{R: 255, G: 205, B: 178, A: 255}
)

// Draw renders a snapshot of the simulation state highlighting airbase locations.
// The returned image has dimensions derived from the embedded Sweden bounds data.
func Draw(sim *Simulation) image.Image {
	if sim == nil {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	sim.AssertInvariants()

	bounds := assets.BoundsData.Overall
	width := int(math.Ceil(bounds.Width*canvasScale)) + canvasMargin*2
	height := int(math.Ceil(bounds.Height*canvasScale)) + canvasMargin*2
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	fillRect(img, img.Bounds(), backgroundColor)

	for _, region := range assets.Regions {
		regionCol := regionColor(region.Name)
		for _, area := range region.Areas {
			if len(area) < 3 {
				continue
			}
			points := make([]image.Point, 0, len(area))
			for _, pt := range area {
				x := int(math.Round((pt.X-bounds.Min.X)*canvasScale)) + canvasMargin
				y := int(math.Round((pt.Y-bounds.Min.Y)*canvasScale)) + canvasMargin
				points = append(points, image.Pt(x, y))
			}
			fillPolygon(img, points, regionCol)
		}
	}

	for _, base := range sim.Airbases() {
		x := (base.Location.X - bounds.Min.X) * canvasScale
		y := (base.Location.Y - bounds.Min.Y) * canvasScale
		cx := int(math.Round(x)) + canvasMargin
		cy := int(math.Round(y)) + canvasMargin

		drawFilledCircle(img, cx, cy, airbaseRadius, airbaseFill)
		drawCircle(img, cx, cy, airbaseRadius, airbaseStroke)
	}

	return img
}

func fillRect(img *image.RGBA, rect image.Rectangle, col color.Color) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			img.Set(x, y, col)
		}
	}
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
		for i := 0; i < len(points); i++ {
			j := (i + 1) % len(points)
			p1 := points[i]
			p2 := points[j]
			if p1.Y == p2.Y {
				continue
			}
			if (p1.Y <= y && p2.Y > y) || (p2.Y <= y && p1.Y > y) {
				rx := float64(p2.X-p1.X) * float64(y-p1.Y) / float64(p2.Y-p1.Y)
				x := int(math.Round(float64(p1.X) + rx))
				intersections = append(intersections, x)
			}
		}
		if len(intersections) < 2 {
			continue
		}
		sort.Ints(intersections)
		for i := 0; i+1 < len(intersections); i += 2 {
			x1 := intersections[i]
			x2 := intersections[i+1]
			if x1 > x2 {
				x1, x2 = x2, x1
			}
			if x2 < img.Bounds().Min.X || x1 >= img.Bounds().Max.X {
				continue
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

func drawFilledCircle(img *image.RGBA, cx, cy, radius int, col color.Color) {
	r2 := radius * radius
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= r2 {
				x := cx + dx
				y := cy + dy
				if image.Pt(x, y).In(img.Bounds()) {
					img.Set(x, y, col)
				}
			}
		}
	}
}

func drawCircle(img *image.RGBA, cx, cy, radius int, col color.Color) {
	if radius <= 0 {
		return
	}
	f := 1 - radius
	dx := 1
	dy := -2 * radius
	x := 0
	y := radius

	setCirclePoints(img, cx, cy, x, y, col)

	for x < y {
		if f >= 0 {
			y--
			dy += 2
			f += dy
		}
		x++
		dx += 2
		f += dx
		setCirclePoints(img, cx, cy, x, y, col)
	}
}

func setCirclePoints(img *image.RGBA, cx, cy, x, y int, col color.Color) {
	points := [8]image.Point{
		{X: cx + x, Y: cy + y},
		{X: cx - x, Y: cy + y},
		{X: cx + x, Y: cy - y},
		{X: cx - x, Y: cy - y},
		{X: cx + y, Y: cy + x},
		{X: cx - y, Y: cy + x},
		{X: cx + y, Y: cy - x},
		{X: cx - y, Y: cy - x},
	}

	for _, pt := range points {
		if pt.In(img.Bounds()) {
			img.Set(pt.X, pt.Y, col)
		}
	}
}

func regionColor(name string) color.Color {
	h := fnv.New32a()
	_, _ = h.Write([]byte(name))
	val := h.Sum32()
	r := uint8(40 + (val>>16)&0x7F)
	g := uint8(60 + (val>>8)&0x7F)
	b := uint8(80 + val&0x7F)
	return color.RGBA{R: r, G: g, B: b, A: 120}
}
