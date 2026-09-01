package sensors

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDegradedReadings(t *testing.T) {
	require.Equal(t, map[string]any{"error": "no clock readings available"}, DegradedReadings("no clock readings available"))
}
