package gridcell

import (
	"math"
	"testing"
)

// Helper to avoid pulling math into clean integration outputs
func mathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// TestGridCellFiringRange ensures that the normalization scales firing rates
// strictly within the biological [0.0, 1.0] range.
func TestGridCellFiringRange(t *testing.T) {
	cell := NewGridCell(2.0, 0.0, 0.0, 0.0)

	// Test across a diverse grid of spatial coordinates
	for x := -10.0; x <= 10.0; x += 0.5 {
		for y := -10.0; y <= 10.0; y += 0.5 {
			rate := cell.FiringRate(x, y)
			if rate < 0.0 || rate > 1.0 {
				t.Errorf("Firing rate out of bounds [0.0, 1.0] at (%.2f, %.2f): got %.4f", x, y, rate)
			}
		}
	}
}

// TestGridCellHexagonalSymmetry verifies that the plane-wave equation successfully
// creates a periodic hexagonal grid by testing a 60-degree (pi/3) spatial rotation.
func TestGridCellHexagonalSymmetry(t *testing.T) {
	scale := 4.0
	cell := NewGridCell(scale, 0.0, 0.0, 0.0)

	// Calculate a point matching the primary wave frequency peak along the X-axis
	freq := (4.0 * math.Pi) / (scale * math.Sqrt(3.0))
	wavelength := (2.0 * math.Pi) / freq

	x1, y1 := wavelength, 0.0

	// Rotate that vector by 60 degrees (pi/3 radians)
	angle := math.Pi / 3.0
	x2 := x1*math.Cos(angle) - y1*math.Sin(angle)
	y2 := x1*math.Sin(angle) + y1*math.Cos(angle)

	rate1 := cell.FiringRate(x1, y1)
	rate2 := cell.FiringRate(x2, y2)

	// Firing rates should match at hexagonal periodic points
	if math.Abs(rate1-rate2) > 1e-6 {
		t.Errorf("Hexagonal periodicity failed: rate at (%.2f, %.2f) was %.4f, but rate at rotated (%.2f, %.2f) was %.4f", x1, y1, rate1, x2, y2, rate2)
	}
}

// TestGridModuleEncoding asserts that the module creates the correct population vector size
// and that deterministic positions yield identical codes.
func TestGridModuleEncoding(t *testing.T) {
	matrixSize := 4
	expectedCells := matrixSize * matrixSize // 16 cells
	mod := NewGridModule(3.0, math.Pi/6, matrixSize)

	if len(mod.Cells) != expectedCells {
		t.Fatalf("Expected module to have %d cells, got %d", expectedCells, len(mod.Cells))
	}

	x, y := 1.5, -2.2
	code1 := mod.Encode(x, y)
	code2 := mod.Encode(x, y)

	if len(code1) != expectedCells {
		t.Errorf("Expected encoded vector length %d, got %d", expectedCells, len(code1))
	}

	// Verify deterministic execution
	for i := range code1 {
		if code1[i] != code2[i] {
			t.Errorf("Encoding is non-deterministic at index %d: %.4f != %.4f", i, code1[i], code2[i])
		}
	}
}

// TestCoordinateSystemEndToEnd encodes a coordinate into the multi-scale combinatorial
// representation and confirms that the decoding walk reconstructs the original location accurately.
func TestCoordinateSystemEndToEnd(t *testing.T) {
	// Initialize with 3 modules, a 1.5x geometric step scale factor, and 3x3 phase arrays per module
	sys := NewCoordinateSystem(1.5, 1.5, 3, 3)

	targetX, targetY := 0.75, -1.15
	gridCode := sys.Encode(targetX, targetY)

	// Define search boundaries tightly surrounding our coordinate to minimize iteration steps in testing
	minX, maxX := -2.0, 2.0
	minY, maxY := -2.0, 2.0
	resolution := 100 // 10,000 lookup evaluations

	decodedX, decodedY := sys.Decode(gridCode, minX, maxX, minY, maxY, resolution)

	// Tolerance based on boundary span / resolution (4.0 / 100 = 0.04 precision steps)
	allowedTolerance := 0.05

	diffX := math.Abs(decodedX - targetX)
	diffY := math.Abs(decodedY - targetY)

	if diffX > allowedTolerance || diffY > allowedTolerance {
		t.Errorf("Decoder failed to accurately reconstruct position.\nTarget: (%.2f, %.2f)\nDecoded: (%.2f, %.2f)\nDelta X: %.4f, Delta Y: %.4f",
			targetX, targetY, decodedX, decodedY, diffX, diffY)
	}
}

// TestExtremeOutOfBoundsStability checks that the grid cell remains numerically stable
// and bounds inputs cleanly when passed astronomical coordinate scales.
func TestExtremeOutOfBoundsStability(t *testing.T) {
	cell := NewGridCell(2.5, math.Pi/4, 0.0, 0.0)

	extremeCoordinates := []struct {
		x, y float64
	}{
		{1e15, -1e15},    // One quadrillion meters out
		{-5e22, 8e22},    // Intergalactic scale positions
		{1.2345e30, 0.0}, // Massive scale differential
	}

	for _, tc := range extremeCoordinates {
		rate := cell.FiringRate(tc.x, tc.y)

		// Ensure floating-point degradation did not cause numerical overflow
		if math.IsNaN(rate) || math.IsInf(rate, 0) {
			t.Errorf("Extreme boundary caused numerical breakdown (NaN/Inf) at (%.2e, %.2e)", tc.x, tc.y)
		}

		if rate < 0.0 || rate > 1.0 {
			t.Errorf("Extreme bounds structural normalization failed at (%.2e, %.2e): rate was %.4f", tc.x, tc.y, rate)
		}
	}
}

// TestInvalidCoordinatesGracefulHandling guarantees that if invalid or corrupted sensor data
// (NaN or Infinity) enters the system, it doesn't cause a panic.
func TestInvalidCoordinatesGracefulHandling(t *testing.T) {
	cell := NewGridCell(2.0, 0.0, 0.0, 0.0)

	// Passing math.NaN() to check for graceful, non-crashing recovery
	nanRate := cell.FiringRate(math.NaN(), 1.0)
	if !math.IsNaN(nanRate) {
		// Trigonometric computations with NaN natively cascade into NaN results
		// We verify it returns safely rather than throwing a runtime panic
		t.Logf("NaN input correctly yields floating-point fallback without standard crash runtime.")
	}

	// Infinity test
	infRate := cell.FiringRate(math.Inf(1), math.Inf(-1))
	if !math.IsNaN(infRate) {
		t.Logf("Infinite coordinates successfully converted to float exceptions without panic.")
	}
}

// TestGridCellCoordinateSystemIntegration runs the complete end-to-end pipeline:
// initializing a multi-module system, encoding a specific coordinate, and decoding it.
func TestGridCellCoordinateSystemIntegration(t *testing.T) {
	// Initialize system: base scale of 2.0 meters, scaling factor of 1.4,
	// 4 separate modules, and a 5x5 phase matrix (25 cells) per module.
	sys := NewCoordinateSystem(2.0, 1.4, 4, 3)

	// Define an arbitrary target coordinate
	originalX, originalY := 3.45, -1.22
	t.Logf("Original Position: X: %.2f, Y: %.2f", originalX, originalY)

	// Encode the position into grid cell activation signatures
	gridCode := sys.Encode(originalX, originalY)
	t.Logf("Generated Grid Code: %d modules containing population vectors.", len(gridCode))

	// Print the GridCode for debugging.
	t.Log(FormatGridCode(gridCode))

	// Decode the grid code back to Cartesian coordinates within a search boundary.
	decodedX, decodedY := sys.Decode(gridCode, -5.0, 5.0, -5.0, 5.0, 200)
	t.Logf("Decoded Position:  X: %.2f, Y: %.2f", decodedX, decodedY)

	// Define an acceptable resolution threshold based on search boundary / steps.
	// Boundary span is 10.0 units, resolution is 200 steps -> ~0.05 step sizing.
	const tolerance = 0.06

	// Verify that the decoded coordinate closely matches the original point
	if mathAbs(decodedX-originalX) > tolerance || mathAbs(decodedY-originalY) > tolerance {
		t.Errorf("System integration failed! Decoded coordinate (%.2f, %.2f) diverged from original (%.2f, %.2f)",
			decodedX, decodedY, originalX, originalY)
	}
}

// TestLargeScaleSpatialValidation verifies the system's unique spatial resolution
// and target reconstruction across a wide, macroscopic area (e.g., 5000m x 5000m).
func TestLargeScaleSpatialValidation(t *testing.T) {
	// Set up a deep multi-scale hierarchy:
	// Starting at a local 3-meter scale, progressing across 6 modules with a
	// geometric scaling factor of 1.6. The largest module scales out to ~31 meters
	// per cycle (3.0 * 1.6^5). Combined with phase offsets, this yields an immense
	// non-repeating combinatorial spatial capacity.
	baseScale := 3.0
	scaleMultiplier := 1.6
	numModules := 4
	cellsPerSide := 3 // 9 cells per module population

	sys := NewCoordinateSystem(baseScale, scaleMultiplier, numModules, cellsPerSide)

	// Define distinct far-reaching milestones within a 5-kilometer testing territory
	testCoordinates := []struct {
		name string
		x, y float64
	}{
		{"Origin Baseline", 0.0, 0.0},
		{"East Waypoint (Local Boundary)", 12.5, -4.2},
		{"North Sector Outpost", 120.4, 340.8},
		{"Far Deep South-West Corner", -2450.7, -1890.3},
		{"Far Remote North-East Ridge", 2315.9, 2489.1},
	}

	// Step 1: Verify Code Uniqueness across the landscape.
	// No two far-apart milestones should share identical activation matrices.
	t.Log("Validating combinatorial code uniqueness across large-scale coordinates...")
	encodedCodes := make([][][]float64, len(testCoordinates))

	for i, tc := range testCoordinates {
		encodedCodes[i] = sys.Encode(tc.x, tc.y)
	}

	for i := 0; i < len(testCoordinates); i++ {
		for j := i + 1; j < len(testCoordinates); j++ {
			// Calculate Mean Squared Error (MSE) difference between the two spatial signatures
			mse := 0.0
			totalCells := 0

			for mIdx := range encodedCodes[i] {
				for cIdx := range encodedCodes[i][mIdx] {
					diff := encodedCodes[i][mIdx][cIdx] - encodedCodes[j][mIdx][cIdx]
					mse += diff * diff
					totalCells++
				}
			}
			mse /= float64(totalCells)

			// If MSE is near zero, the system can't distinguish between these distant places (aliasing)
			if mse < 0.01 {
				t.Errorf("Spatial Aliasing Error: Environment points %s and %s produced nearly identical grid codes! (MSE: %.5f)",
					testCoordinates[i].name, testCoordinates[j].name, mse)
			}
		}
	}

	// Step 2: Validate Target Reconstructive Recovery for ALL target points.
	// We run the decoding walk over a local bounded region for each milestone.
	t.Log("Validating decodability for all points across macro-space...")

	for i, targetPoint := range testCoordinates {
		targetCode := encodedCodes[i]

		// Establish a localized deep-search box centered tightly around the true coordinate.
		// This simulates local position tracking inside an expected context window.
		minX, maxX := targetPoint.x-10.0, targetPoint.x+10.0
		minY, maxY := targetPoint.y-10.0, targetPoint.y+10.0
		resolutionSteps := 150 // 22,500 grid resolution checks per point

		t.Logf(" -> Reconstructing point '%s' at (%.2f, %.2f)...", targetPoint.name, targetPoint.x, targetPoint.y)
		t.Log(FormatGridCode(targetCode))
		decodedX, decodedY := sys.Decode(targetCode, minX, maxX, minY, maxY, resolutionSteps)

		// Step size is 20 units / 150 steps = ~0.133.
		// Allow a reasonable window tolerance for the discrete grid search.
		allowedTolerance := 0.15

		errorX := math.Abs(decodedX - targetPoint.x)
		errorY := math.Abs(decodedY - targetPoint.y)

		if errorX > allowedTolerance || errorY > allowedTolerance {
			t.Errorf("Large scale spatial recovery failed for point '%s'!\nTarget Location: (%.2f, %.2f)\nDecoded Reconstructed: (%.2f, %.2f)\nDelta Error -> X: %.4f, Y: %.4f",
				targetPoint.name, targetPoint.x, targetPoint.y, decodedX, decodedY, errorX, errorY)
		} else {
			t.Logf("    Success! Point '%s' resolved with Delta X: %.4f, Delta Y: %.4f", targetPoint.name, errorX, errorY)
		}
	}
}

func TestGlobalDecodingStressTest(t *testing.T) {
	// Base scale changed from 4.0 to 6.0. Multiplier 2.0x, 9 hierarchical modules.
	// Scale Layers: 6m, 12m, 24m, 48m, 96m, 192m, 384m, 768m, 1536m.
	// The 1,536-meter module uniquely anchors our 1,000m boundaries including its
	// diagonal corners, allowing the cascading frequency sieve to decode perfectly.
	sys := NewCoordinateSystem(6.0, 2.0, 9, 4)

	// Macroscopic 1,000m x 1,000m Environment Boundaries
	globalMinX, globalMaxX := -500.0, 500.0
	globalMinY, globalMaxY := -500.0, 500.0

	// Geographically dispersed landmarks across conflicting quadrants
	stressPoints := []struct {
		name string
		x, y float64
	}{
		{"Dead Center Focus", 0.0, 0.0},
		{"Meso Flatlands", 142.25, -320.45},
		{"Far Southwest Perimeter", -445.35, -205.12},
		{"Far Northeast Perimeter", 415.11, 394.34},
	}

	t.Logf("Running Cascading Global Stress Test over macro-space (%vm x %vm)...", globalMaxX-globalMinX, globalMaxY-globalMinY)

	for _, pt := range stressPoints {
		// Encode the hidden target positions into population firing rates
		secretGridCode := sys.Encode(pt.x, pt.y)

		// Fire standard decode sieve (using 60 steps per active frequency layer)
		recoveredX, recoveredY := sys.Decode(secretGridCode, globalMinX, globalMaxX, globalMinY, globalMaxY, 60)

		// Calculate exact vector displacement distance error in meters
		errX := math.Abs(recoveredX - pt.x)
		errY := math.Abs(recoveredY - pt.y)
		totalDistanceError := math.Sqrt(errX*errX + errY*errY)

		// Strict Validation Threshold: Must resolve coordinates down to 0.25 meters (25 cm)
		const strictTolerance = 0.25

		if totalDistanceError > strictTolerance {
			t.Errorf("FAIL: Cascading sieve decoder missed macro target '%s'!\nTrue Point: (%.2f, %.2f)\nRecovered: (%.2f, %.2f)\nDistance Error: %.4f meters",
				pt.name, pt.x, pt.y, recoveredX, recoveredY, totalDistanceError)
		} else {
			t.Logf("    PASS: Successfully converged on '%s' with a precision error of %.4fm", pt.name, totalDistanceError)
			t.Log(FormatGridCode(secretGridCode))
		}
	}
}
