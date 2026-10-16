package gridcell

import (
	"math"
)

// Landmark represents a recognizable environmental anchor point.
// It stores its fixed multi-scale spatial grid signature but allows its heading to change.
type Landmark struct {
	// The original continuous coordinate code matrix cached from CoordinateSystem.Encode()
	GridCode [][]float64
	// Current viewing orientation angle of the landmark in radians (can be mutated dynamically)
	Heading float64
	// Reference pointer to the parent coordinate layout to parse cell tuning properties
	sys *CoordinateSystem
}

// NewLandmark builds a localized landmark instance bound to a stable grid network.
func (cs *CoordinateSystem) NewLandmark(gridCode [][]float64, initialHeadingRad float64) *Landmark {
	return &Landmark{
		GridCode: gridCode,
		Heading:  initialHeadingRad,
		sys:      cs,
	}
}

// SetHeading updates the landmark's active view orientation alignment.
func (l *Landmark) SetHeading(newHeadingRad float64) {
	l.Heading = newHeadingRad
}

// EncodeToBits packs the landmark's combined spatial position and active heading angle
// into a flat, high-dimensional binary address vector stored in uint64 words for SDM.
func (l *Landmark) EncodeToBits(activationThreshold float64) []uint64 {
	// 1. Calculate total cells across the system matrix to size our packed bit stream
	totalCells := 0
	for _, layer := range l.GridCode {
		totalCells += len(layer)
	}

	// We append 64 extra heading bits to the end of the address word slice
	// to cleanly bind space and orientation into a unified high-dimensional SDM key.
	totalAddressBits := totalCells + 64
	numWords := (totalAddressBits + 63) / 64
	packedWords := make([]uint64, numWords)

	// 2. Map the spatial grid matrix bits into the primary sector words
	bitIdx := 0
	for _, layer := range l.GridCode {
		for _, rate := range layer {
			if rate >= activationThreshold {
				packedWords[bitIdx/64] |= (1 << (uint(bitIdx) % 64))
			}
			bitIdx++
		}
	}

	// 3. Map the continuous mutable heading angle into a 64-bit binary thermometer code segment.
	// Thermometer encoding preserves spatial smoothness: small angle changes cause small bit flips,
	// keeping Hamming distance perfectly linear with heading changes.
	normalizedHeading := math.Mod(l.Heading, 2.0*math.Pi)
	if normalizedHeading < 0 {
		normalizedHeading += 2.0 * math.Pi
	}

	// Determine how many of the 64 heading bits should flip to 1 based on the angle (0 to 2*Pi)
	activeHeadingBits := int((normalizedHeading / (2.0 * math.Pi)) * 64.0)
	if activeHeadingBits > 64 {
		activeHeadingBits = 64
	}

	for h := 0; h < activeHeadingBits; h++ {
		globalHeadingBitIdx := totalCells + h
		packedWords[globalHeadingBitIdx/64] |= (1 << (uint(globalHeadingBitIdx) % 64))
	}

	return packedWords
}

// ExtractHeading FromBits isolated the trailing 64-bit thermometer sector of an SDM address
// and converts the active bit count back into a continuous heading angle in radians [0, 2*Pi).
func (cs *CoordinateSystem) ExtractHeadingFromBits(packedWords []uint64) float64 {
	// 1. Calculate how many total cells precede the heading block to know where to start reading
	totalSpatialCells := cs.TotalCells()

	// 2. Count the active thermometer bits in the final 64-bit block
	activeHeadingBits := 0
	for h := 0; h < 64; h++ {
		globalHeadingBitIdx := totalSpatialCells + h

		// Safely extract the bit from the packed storage words
		isFired := (packedWords[globalHeadingBitIdx/64] & (1 << (uint(globalHeadingBitIdx) % 64))) != 0
		if isFired {
			activeHeadingBits++
		}
	}

	// 3. Convert the active bit count back into a continuous heading angle in radians
	// We map the discrete bucket step (0 to 64) back onto the 0 to 2*Pi radian scale
	return (float64(activeHeadingBits) / 64.0) * (2.0 * math.Pi)
}
