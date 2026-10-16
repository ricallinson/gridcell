package gridcell

import (
	"math"
)

type CoordinateSystem struct {
	Modules []*GridModule
}

func NewCoordinateSystem(baseScale, scaleMultiplier float64, numModules, cellsPerModuleSide int) *CoordinateSystem {
	cs := &CoordinateSystem{Modules: make([]*GridModule, numModules)}
	currentScale := baseScale
	for i := 0; i < numModules; i++ {
		orientation := float64(i) * (math.Pi / 12.0)
		cs.Modules[i] = NewGridModule(currentScale, orientation, cellsPerModuleSide)
		currentScale *= scaleMultiplier
	}
	return cs
}

func (cs *CoordinateSystem) Encode(x, y float64) [][]float64 {
	code := make([][]float64, len(cs.Modules))
	for i, mod := range cs.Modules {
		code[i] = mod.Encode(x, y)
	}
	return code
}

// Decode tracks down coordinates inside given boundaries using a strict cascading frequency sieve.
// It applies a 1.5x safety scale-padding window during descending transitions to eliminate boundary clipping.
func (cs *CoordinateSystem) Decode(targetCode [][]float64, minX, maxX, minY, maxY float64, searchResolution int) (float64, float64) {
	numModules := len(cs.Modules)
	curMinX, curMaxX := minX, maxX
	curMinY, curMaxY := minY, maxY

	bestX, bestY := 0.0, 0.0

	// Sieve down the module layers from most macroscopic to most microscopic
	for stage := numModules - 1; stage >= 0; stage-- {
		bestX, bestY = 0.0, 0.0
		minErr := math.MaxFloat64

		stepX := (curMaxX - curMinX) / float64(searchResolution)
		stepY := (curMaxY - curMinY) / float64(searchResolution)

		for i := 0; i <= searchResolution; i++ {
			cx := curMinX + float64(i)*stepX
			for j := 0; j <= searchResolution; j++ {
				cy := curMinY + float64(j)*stepY

				totalErr := 0.0
				// Evaluate ONLY from the current staging tier up to the largest module.
				// This isolates high frequencies out of early global iterations.
				for mIdx := stage; mIdx < numModules; mIdx++ {
					mod := cs.Modules[mIdx]
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

		// Tighten the boundary focus window with a 1.5x safety multiplier
		// to gracefully capture phase shifts near boundary corners.
		if stage > 0 {
			nextModuleScale := cs.Modules[stage-1].Scale
			padding := nextModuleScale * 1.5
			curMinX, curMaxX = bestX-padding, bestX+padding
			curMinY, curMaxY = bestY-padding, bestY+padding
		}
	}

	return bestX, bestY
}
