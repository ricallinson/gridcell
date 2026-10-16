package gridcell

import (
	"math"
)

// GridCell represents a single grid cell with a specific scale, orientation, and phase.
type GridCell struct {
	Scale       float64 // Spatial wavelength (distance between firing peaks)
	Orientation float64 // Grid rotation angle in radians
	PhaseX      float64 // Spatial phase offset X
	PhaseY      float64 // Spatial phase offset Y
}

// NewGridCell initializes a cell with its spatial tuning properties.
func NewGridCell(scale, orientation, phaseX, phaseY float64) GridCell {
	return GridCell{
		Scale:       scale,
		Orientation: orientation,
		PhaseX:      phaseX,
		PhaseY:      phaseY,
	}
}

// FiringRate calculates the cell's activation level [0.0, 1.0] at a given (x, y) position.
// It uses modulo wrapping to guarantee numerical stability at extreme out-of-bounds coordinates.
func (gc GridCell) FiringRate(x, y float64) float64 {
	dx := x - gc.PhaseX
	dy := y - gc.PhaseY

	// Failsafe: Wrap coordinates using Modulo to protect against float precision
	// degradation at extreme spatial scales (e.g., values > 1e12)
	dx = math.Mod(dx, gc.Scale)
	dy = math.Mod(dy, gc.Scale*math.Sqrt(3.0))

	// Convert spatial scale to wave frequency components
	freq := (4.0 * math.Pi) / (gc.Scale * math.Sqrt(3.0))

	var sum float64
	for i := 0; i < 3; i++ {
		angle := gc.Orientation + float64(i)*math.Pi/3.0
		kx := freq * math.Cos(angle)
		ky := freq * math.Sin(angle)

		sum += math.Cos(kx*dx + ky*dy)
	}

	return (sum + 3.0) / 6.0
}
