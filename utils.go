package gridcell

import (
	"bytes"
	"fmt"
)

// TotalElements counts the combined number of individual float64 values
// across all sub-slices within a 2D float64 structure.
func TotalElements(matrix [][]float64) int {
	count := 0
	for i := range matrix {
		count += len(matrix[i])
	}
	return count
}

// FormatGridCode takes the raw output of CoordinateSystem.Encode() and converts it
// into a visual text block displaying activation density sparklines for debugging.
func FormatGridCode(code [][]float64) string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("GridCode Debug Matrix (%d Modules with %d values):\n", len(code), TotalElements(code)))

	for modIdx, moduleVectors := range code {
		buf.WriteString(fmt.Sprintf("  Module [%d] (%d cells): [", modIdx, len(moduleVectors)))

		for _, rate := range moduleVectors {
			// Map the [0.0, 1.0] biological firing range to visual density blocks
			var visualChar string
			switch {
			case rate < 0.15:
				visualChar = " " // Inactive
			case rate < 0.4:
				visualChar = "░" // Weak activation
			case rate < 0.7:
				visualChar = "▒" // Moderate activation
			case rate < 0.85:
				visualChar = "▓" // Strong activation
			default:
				visualChar = "█" // Saturated peak firing
			}
			buf.WriteString(visualChar)
		}
		buf.WriteString("]\n")
	}

	return buf.String()
}
