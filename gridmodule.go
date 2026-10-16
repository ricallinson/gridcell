package gridcell

import (
	"math"
)

type GridModule struct {
	Scale       float64
	Orientation float64
	Cells       []GridCell
}

func NewGridModule(scale, orientation float64, matrixSize int) *GridModule {
	cells := make([]GridCell, 0, matrixSize*matrixSize)
	for i := 0; i < matrixSize; i++ {
		for j := 0; j < matrixSize; j++ {
			px := (float64(i) / float64(matrixSize)) * scale
			py := (float64(j) / float64(matrixSize)) * scale * (math.Sqrt(3.0) / 2.0)
			cells = append(cells, NewGridCell(scale, orientation, px, py))
		}
	}
	return &GridModule{Scale: scale, Orientation: orientation, Cells: cells}
}

func (gm *GridModule) Encode(x, y float64) []float64 {
	rates := make([]float64, len(gm.Cells))
	for i, cell := range gm.Cells {
		rates[i] = cell.FiringRate(x, y)
	}
	return rates
}
