package gridcell

import (
	"math"
)

// GridModule represents a population of grid cells that share the same scale and orientation
// but vary across different spatial phases to uniformly cover a localized topological tile.
type GridModule struct {
	Scale       float64
	Orientation float64
	Cells       []GridCell
}

// NewGridModule creates a module containing an NxN population matrix with distributed phases.
func NewGridModule(scale, orientation float64, matrixSize int) *GridModule {
	cells := make([]GridCell, 0, matrixSize*matrixSize)

	// Uniformly distribute phase offsets across the baseline periodic spatial unit
	for i := 0; i < matrixSize; i++ {
		for j := 0; j < matrixSize; j++ {
			px := (float64(i) / float64(matrixSize)) * scale
			py := (float64(j) / float64(matrixSize)) * scale * (math.Sqrt(3.0) / 2.0)
			cells = append(cells, NewGridCell(scale, orientation, px, py))
		}
	}

	return &GridModule{
		Scale:       scale,
		Orientation: orientation,
		Cells:       cells,
	}
}

// Encode converts a 2D coordinate into a flat vector of firing rates for this specific module.
func (gm *GridModule) Encode(x, y float64) []float64 {
	rates := make([]float64, len(gm.Cells))
	for i, cell := range gm.Cells {
		rates[i] = cell.FiringRate(x, y)
	}
	return rates
}
