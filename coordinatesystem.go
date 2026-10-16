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
// It has been broken down into modular steps for clean structural readability.
func (cs *CoordinateSystem) Decode(targetCode [][]float64, minX, maxX, minY, maxY float64, searchResolution int) (float64, float64) {
	// Step 1: Detect binary thresholds vs continuous analog configurations
	isBinarized, threshold := cs.detectBinarization(targetCode)

	// Step 2: Run the cascading module sieve to lock onto the correct macro-neighborhood
	bestX, bestY := cs.runCascadingSieve(targetCode, minX, maxX, minY, maxY, searchResolution, isBinarized, threshold)

	// Step 3: Run local sub-meter refinement to eliminate step quantization errors
	return cs.refineLocalPrecision(targetCode, bestX, bestY, isBinarized, threshold)
}

// detectBinarization checks if the matrix contains exclusively zeros and a uniform activation threshold.
func (cs *CoordinateSystem) detectBinarization(targetCode [][]float64) (isBinarized bool, threshold float64) {
	for _, layer := range targetCode {
		for _, rate := range layer {
			if rate > threshold {
				threshold = rate
			}
		}
	}

	if threshold == 0.0 {
		return false, 0.0
	}

	for _, layer := range targetCode {
		for _, rate := range layer {
			if rate != 0.0 && rate != threshold {
				return false, threshold // Found intermediate analog variances
			}
		}
	}
	return true, threshold
}

// computeMatchError calculates the mean squared mismatch between candidate continuous values
// and targets using a continuous soft-margin probabilistic sigmoid curve for binarized data.
func (cs *CoordinateSystem) computeMatchError(targetCode [][]float64, cx, cy float64, startStage int, isBinarized bool, threshold float64) float64 {
	totalErr := 0.0
	numModules := len(cs.Modules)

	for mIdx := startStage; mIdx < numModules; mIdx++ {
		currentRates := cs.Modules[mIdx].Encode(cx, cy)
		for cIdx, rate := range currentRates {
			var targetVal float64
			var candidateVal float64

			if isBinarized {
				if targetCode[mIdx][cIdx] == threshold {
					targetVal = 1.0
				} else {
					targetVal = 0.0
				}
				// Soft-Margin Sigmoid Function preserves sub-meter continuous contours
				candidateVal = 1.0 / (1.0 + math.Exp(-25.0*(rate-threshold)))
			} else {
				targetVal = targetCode[mIdx][cIdx]
				candidateVal = rate
			}

			diff := candidateVal - targetVal
			totalErr += diff * diff
		}
	}
	return totalErr
}

// runCascadingSieve steps down through module frequencies to locate the base target region.
func (cs *CoordinateSystem) runCascadingSieve(targetCode [][]float64, minX, maxX, minY, maxY float64, res int, isBinarized bool, threshold float64) (float64, float64) {
	numModules := len(cs.Modules)
	curMinX, curMaxX := minX, maxX
	curMinY, curMaxY := minY, maxY
	bestX, bestY := 0.0, 0.0

	for stage := numModules - 1; stage >= 0; stage-- {
		bestX, bestY = 0.0, 0.0
		minErr := math.MaxFloat64

		stepX := (curMaxX - curMinX) / float64(res)
		stepY := (curMaxY - curMinY) / float64(res)

		for i := 0; i <= res; i++ {
			cx := curMinX + float64(i)*stepX
			for j := 0; j <= res; j++ {
				cy := curMinY + float64(j)*stepY

				totalErr := cs.computeMatchError(targetCode, cx, cy, stage, isBinarized, threshold)
				if totalErr < minErr {
					minErr = totalErr
					bestX = cx
					bestY = cy
				}
			}
		}

		if stage > 0 {
			nextModuleScale := cs.Modules[stage-1].Scale
			padding := nextModuleScale * 1.5
			curMinX, curMaxX = bestX-padding, bestX+padding
			curMinY, curMaxY = bestY-padding, bestY+padding
		}
	}
	return bestX, bestY
}

// refineLocalPrecision executes a high-density micro-grid sweep centered around the sieve's locked coordinate.
func (cs *CoordinateSystem) refineLocalPrecision(targetCode [][]float64, bestX, bestY float64, isBinarized bool, threshold float64) (float64, float64) {
	refineMinX, refineMaxX := bestX-1.0, bestX+1.0
	refineMinY, refineMaxY := bestY-1.0, bestY+1.0
	refineRes := 120

	finalX, finalY := bestX, bestY
	minErr := math.MaxFloat64
	rx := (refineMaxX - refineMinX) / float64(refineRes)
	ry := (refineMaxY - refineMinY) / float64(refineRes)

	for i := 0; i <= refineRes; i++ {
		cx := refineMinX + float64(i)*rx
		for j := 0; j <= refineRes; j++ {
			cy := refineMinY + float64(j)*ry

			// Start from Stage 0 to evaluate global spatial alignment across ALL modules
			totalErr := cs.computeMatchError(targetCode, cx, cy, 0, isBinarized, threshold)
			if totalErr < minErr {
				minErr = totalErr
				finalX = cx
				finalY = cy
			}
		}
	}
	return finalX, finalY
}

// EncodeToBits accepts a continuous matrix output from CoordinateSystem.Encode()
// and packages it into an SDM-ready flat binary bit array packed in uint64 words.
func (cs *CoordinateSystem) EncodeToBits(floatCode [][]float64, activationThreshold float64) []uint64 {
	// Count total element length to build exact word storage limits
	totalCells := 0
	for _, layer := range floatCode {
		totalCells += len(layer)
	}

	numWords := (totalCells + 63) / 64
	packedWords := make([]uint64, numWords)

	bitIdx := 0
	for _, layer := range floatCode {
		for _, rate := range layer {
			if rate >= activationThreshold {
				packedWords[bitIdx/64] |= (1 << (uint(bitIdx) % 64))
			}
			bitIdx++
		}
	}

	return packedWords
}

// DecodeBits converts packed uint64 words back into a continuous multi-layer[][]float64 grid matrix
// that can be cleanly evaluated by your existing geometric CoordinateSystem.Decode loop.
func (cs *CoordinateSystem) DecodeBits(packedWords []uint64, activationThreshold float64) [][]float64 {
	decodedFloatCode := make([][]float64, len(cs.Modules))

	bitIdx := 0
	for mIdx, mod := range cs.Modules {
		decodedFloatCode[mIdx] = make([]float64, len(mod.Cells))
		for cIdx := range mod.Cells {
			// Extract individual boolean bit from packed word indices
			isFired := (packedWords[bitIdx/64] & (1 << (uint(bitIdx) % 64))) != 0

			if isFired {
				decodedFloatCode[mIdx][cIdx] = activationThreshold
			} else {
				decodedFloatCode[mIdx][cIdx] = 0.0
			}
			bitIdx++
		}
	}

	return decodedFloatCode
}
