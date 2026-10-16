package gridcell

import (
	"math"
	"testing"
)

func TestLandmarkMutableHeadingSDM(t *testing.T) {
	// Initialize standard 1,000m system context: 9 modules, 4x4 matrix
	sys := NewCoordinateSystem(6.0, 2.0, 9, 4)
	threshold := 0.65

	// Capture grid coordinates at an arbitrary location
	targetX, targetY := 200.50, -150.75
	baseGridCode := sys.Encode(targetX, targetY)

	// 1. Instantiate the landmark pointing directly East (0 Radians)
	landmark := sys.NewLandmark(baseGridCode, 0.0)

	// Generate baseline address representation
	bitsEast := landmark.EncodeToBits(threshold)

	// 2. Shift heading slightly to the North-East (15 degrees / ~0.26 radians)
	landmark.SetHeading(15.0 * (math.Pi / 180.0))
	bitsNortheast := landmark.EncodeToBits(threshold)

	// 3. Twist heading completely around to the West (180 degrees / Pi radians)
	landmark.SetHeading(math.Pi)
	bitsWest := landmark.EncodeToBits(threshold)

	// Calculate Hamming Distances between the different heading states
	nearAngleDist := HammingDistance(bitsEast, bitsNortheast)
	farAngleDist := HammingDistance(bitsEast, bitsWest)

	t.Logf("Total Packed Address Length: %d uint64 words", len(bitsEast))
	t.Logf("Hamming Distance for a minor 15° heading adjustment: %d bits", nearAngleDist)
	t.Logf("Hamming Distance for a major 180° heading shift:    %d bits", farAngleDist)

	// Validation Assertions: Bit changes must scale proportionally with heading changes
	if nearAngleDist == 0 {
		t.Errorf("FAIL: Modifying the landmark heading failed to register any changes in the packed SDM bit array.")
	}

	if nearAngleDist >= farAngleDist {
		t.Errorf("FAIL: SDM orientation topology failed. Minor turn distance (%d) should be lower than complete turnaround distance (%d)",
			nearAngleDist, farAngleDist)
	} else {
		t.Log("PASS: Landmark address vectors handle mutable orientations cleanly across high-dimensional bit spaces.")
	}
}

func TestLandmarkHeadingExtraction(t *testing.T) {
	// 9 Modules, 4x4 matrix configuration matching previous architectures
	sys := NewCoordinateSystem(6.0, 2.0, 9, 4)
	threshold := 0.65

	// Capture baseline coordinates
	baseGridCode := sys.Encode(150.0, -85.0)

	// Define an arbitrary target heading (e.g., 120 degrees converted to radians)
	originalTargetHeading := 120.0 * (math.Pi / 180.0)

	// 1. Create landmark and set the specific target direction
	landmark := sys.NewLandmark(baseGridCode, originalTargetHeading)

	// 2. Compile spatial code and orientation into packed binary words
	packedBits := landmark.EncodeToBits(threshold)

	// 3. Extract just the continuous heading back out from the raw bit arrays
	extractedHeading := sys.ExtractHeadingFromBits(packedBits)

	// Convert back to degrees for clear, readable validation logs
	origDegrees := originalTargetHeading * (180.0 / math.Pi)
	extractedDegrees := extractedHeading * (180.0 / math.Pi)

	t.Logf("Original Input Heading:  %.2f° (%.4f rad)", origDegrees, originalTargetHeading)
	t.Logf("Extracted Bit Heading:   %.2f° (%.4f rad)", extractedDegrees, extractedHeading)

	// Because a 64-bit thermometer encoding divides 360 degrees into 64 discrete steps,
	// the maximum possible rounding step variance is 360 / 64 = 5.625 degrees.
	// We allow a safe rounding tolerance window of 6.0 degrees.
	const resolutionToleranceDegrees = 6.0
	degreeDelta := math.Abs(extractedDegrees - origDegrees)

	if degreeDelta > resolutionToleranceDegrees {
		t.Errorf("FAIL: Heading extractor drifted outside acceptable quantization limits! Delta: %.2f°", degreeDelta)
	} else {
		t.Logf("PASS: Successfully extracted mutable heading from packed bits within resolution step precision.")
	}
}
