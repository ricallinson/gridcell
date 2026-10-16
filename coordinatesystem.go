package gridcell

import (
	"math"
)

// CoordinateSystem combines multiple distinct grid modules to establish a multi-scale spatial system.
type CoordinateSystem struct {
	Modules []*GridModule
}

// NewCoordinateSystem sets up a system using a base scale, geometric step multiplier, and module count.
func NewCoordinateSystem(baseScale, scaleMultiplier float64, numModules, cellsPerModuleSide int) *CoordinateSystem {
	cs := &CoordinateSystem{
		Modules: make([]*GridModule, numModules),
	}

	currentScale := baseScale
	for i := 0; i < numModules; i++ {
		// Vary the structural orientation slightly across modules to reduce global aliasing
		orientation := float64(i) * (math.Pi / 12.0)
		cs.Modules[i] = NewGridModule(currentScale, orientation, cellsPerModuleSide)
		currentScale *= scaleMultiplier
	}

	return cs
}

// Encode converts a 2D location into a multi-module neural ensemble code.
func (cs *CoordinateSystem) Encode(x, y float64) [][]float64 {
	code := make([][]float64, len(cs.Modules))
	for i, mod := range cs.Modules {
		code[i] = mod.Encode(x, y)
	}
	return code
}

// Decode translates a multi-module grid code back into Cartesian coordinates.
// It uses a Mean Squared Error (MSE) minimization walk across a specified bounded search space.
func (cs *CoordinateSystem) Decode(targetCode [][]float64, minX, maxX, minY, maxY float64, searchResolution int) (bestX, bestY float64) {
	minErr := math.MaxFloat64

	stepX := (maxX - minX) / float64(searchResolution)
	stepY := (maxY - minY) / float64(searchResolution)

	for i := 0; i <= searchResolution; i++ {
		cx := minX + float64(i)*stepX
		for j := 0; j <= searchResolution; j++ {
			cy := minY + float64(j)*stepY

			totalErr := 0.0
			for mIdx, mod := range cs.Modules {
				currentRates := mod.Encode(cx, cy)
				for cIdx := range currentRates {
					diff := currentRates[cIdx] - targetCode[mIdx][cIdx]
					totalErr += diff * diff
				}
			}

			if totalErr < minErr {
				minErr = totalErr
				bestX = cx
				bestY = cy
			}
		}
	}
	return bestX, bestY
}
