package gridcell

import (
	"math"
	"testing"
)

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
		}
	}
}
