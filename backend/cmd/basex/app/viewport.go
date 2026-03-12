package app

import (
	"github.com/bas-x/basex/assets"
	"github.com/bas-x/basex/services"
)

type Viewport struct {
	MinX   float64 `json:"minX"`
	MinY   float64 `json:"minY"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

func DefaultViewport() Viewport {
	b := assets.BoundsData.Overall
	return Viewport{MinX: b.Min.X, MinY: b.Min.Y, Width: b.Width, Height: b.Height}
}

func (v Viewport) FocusAirbase(base services.Airbase) Viewport {
	width := maxf(v.Width*0.35, 120)
	height := maxf(v.Height*0.35, 120)
	return Viewport{
		MinX:   base.Location.X - width/2,
		MinY:   base.Location.Y - height/2,
		Width:  width,
		Height: height,
	}
}

func (v Viewport) Pan(dx, dy float64) Viewport {
	v.MinX += dx
	v.MinY += dy
	return v
}

func (v Viewport) Zoom(scale float64) Viewport {
	if scale <= 0 {
		return v
	}
	centerX := v.MinX + v.Width/2
	centerY := v.MinY + v.Height/2
	v.Width /= scale
	v.Height /= scale
	v.MinX = centerX - v.Width/2
	v.MinY = centerY - v.Height/2
	return v
}

func (v Viewport) ZoomAt(scale, relX, relY float64) Viewport {
	if scale <= 0 {
		return v
	}
	if relX < 0 {
		relX = 0
	} else if relX > 1 {
		relX = 1
	}
	if relY < 0 {
		relY = 0
	} else if relY > 1 {
		relY = 1
	}
	anchorX := v.MinX + relX*v.Width
	anchorY := v.MinY + relY*v.Height
	v.Width /= scale
	v.Height /= scale
	v.MinX = anchorX - relX*v.Width
	v.MinY = anchorY - relY*v.Height
	return v
}

func maxf(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
