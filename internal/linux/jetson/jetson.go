package jetson

import "github.com/rinzlerlabs/sbcidentify"

// IsJetson reports whether the host is a Jetson board. It checks
// sbcidentify's board identification first, then falls back to HasJetsonGpu
// (probing directly for known Jetson GPU sysfs nodes), since sbcidentify can
// fail to identify some Jetsons entirely — confirmed on a Jetson AGX Orin
// Developer Kit, where sbcidentify.GetBoardType returns "unknown board" /
// "cannot identify NVIDIA board".
//
// This is the single place Jetson detection should happen; callers that need
// to know "is this a Jetson" should use this instead of their own
// sbcidentify check, so a future detection fix (or another fallback) only
// needs to be made once.
func IsJetson() bool {
	return sbcidentify.IsJetson() || HasJetsonGpu()
}
