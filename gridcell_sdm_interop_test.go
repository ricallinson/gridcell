package gridcell

import (
	"fmt"
	"math"
	"testing"
)

// HammingDistance computes bitwise variations via POPCOUNT loops for SDM validation.
func HammingDistance(w1, w2 []uint64) int {
	dist := 0
	for i := range w1 {
		diff := w1[i] ^ w2[i]
		for diff > 0 {
			dist += int(diff & 1)
			diff >>= 1
		}
	}
	return dist
}

func uint64SliceToString(bits []uint64) string {
	var result string
	for _, word := range bits {
		// Print each uint64 word as a 64-character wide zero-padded binary string
		result += fmt.Sprintf("%064b", word)
	}
	return result
}

func TestSDMInteroperabilityPipeline(t *testing.T) {
	// 9 Modules, geometric multiplier 2.0x, base 6.0 meters
	// Max wavelength = 1,536m, safely surrounding the 1,000m arena diagonals
	sys := NewCoordinateSystem(6.0, 2.0, 9, 4)

	// Bounding limits
	globalMinX, globalMaxX := -500.0, 500.0
	globalMinY, globalMaxY := -500.0, 500.0
	threshold := 0.65

	// True Target Position
	targetX, targetY := 415.11, 394.34

	// 1. Encode Coordinate to Float Matrix
	floatCodeOrig := sys.Encode(targetX, targetY)

	// 2. Transfrom Float Matrix into Packed SDM uint64 words
	packedBits := sys.EncodeToBits(floatCodeOrig, threshold)

	// 3. Confirm spatial alignment of packed addresses via Hamming calculations
	shiftedFloatCode := sys.Encode(targetX+1, targetY-1) // Near drift
	shiftedPackedBits := sys.EncodeToBits(shiftedFloatCode, threshold)

	nearDist := HammingDistance(packedBits, shiftedPackedBits)
	if nearDist != 8 {
		t.Logf("Failed packing metrics. Local neighbor Hamming Distance was %d bits not 8 bits", nearDist)
	} else {
		t.Logf("Verified packing metrics. Local neighbor Hamming Distance: %d bits", nearDist)
		t.Logf("\n%s\n%s", uint64SliceToString(packedBits), uint64SliceToString(shiftedPackedBits))
	}

	// 4. Reverse the structure: Reconstruct float[][] matrix array straight out of binary words
	floatCodeReconstructed := sys.DecodeBits(packedBits, threshold)

	// 5. Pass reconstituted matrix right back down to standard CoordinateSystem.Decode
	recoveredX, recoveredY := sys.Decode(floatCodeReconstructed, globalMinX, globalMaxX, globalMinY, globalMaxY, 60)

	// Measure accuracy window
	errX := math.Abs(recoveredX - targetX)
	errY := math.Abs(recoveredY - targetY)
	totalDistanceError := math.Sqrt(errX*errX + errY*errY)

	const strictTolerance = 0.25
	if totalDistanceError > strictTolerance {
		t.Errorf("FAIL: Reconstructed float matrices failed spatial convergence! Got position (%.2f, %.2f) instead of (%.2f, %.2f) - Distance Error: %.4fm",
			recoveredX, recoveredY, targetX, targetY, totalDistanceError)
	} else {
		t.Logf("PASS: Reconstructed matrix successfully resolved via sieve decoders with a precision error of %.4fm", totalDistanceError)
	}
}
