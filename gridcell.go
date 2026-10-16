package gridcell

import (
	"math"
)

type GridCell struct {
	Scale       float64
	Orientation float64
	PhaseX      float64
	PhaseY      float64
}

func NewGridCell(scale, orientation, phaseX, phaseY float64) GridCell {
	return GridCell{Scale: scale, Orientation: orientation, PhaseX: phaseX, PhaseY: phaseY}
}

// FiringRate projects spatial coordinates onto interlocking 60-degree plane waves.
func (gc GridCell) FiringRate(x, y float64) float64 {
	dx := x - gc.PhaseX
	dy := y - gc.PhaseY
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
