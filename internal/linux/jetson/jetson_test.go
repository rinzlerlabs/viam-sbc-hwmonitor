package jetson

import (
	"testing"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/stretchr/testify/require"
)

// TestIsJetsonFallsBackWhenSbcidentifyMisdetects regression-tests IsJetson's
// whole reason for existing: on a real Jetson AGX Orin Developer Kit,
// sbcidentify.IsJetson() (and even GetBoardType) fails to identify the board
// at all ("unknown board" / "cannot identify NVIDIA board"), confirmed via
// `go run ./cmd/sbcidentify` on the device. IsJetson must still report true
// there via the HasJetsonGpu sysfs fallback. This only runs on real Jetson
// hardware (self-selected via HasJetsonGpu), so it's a no-op skip elsewhere.
func TestIsJetsonFallsBackWhenSbcidentifyMisdetects(t *testing.T) {
	if !HasJetsonGpu() {
		t.Skip("test requires Jetson GPU sysfs nodes")
	}
	require.True(t, IsJetson())
	if !sbcidentify.IsJetson() {
		t.Log("confirms the fallback is load-bearing here: sbcidentify.IsJetson() is false on this board")
	}
}
