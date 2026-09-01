package sensors

// DegradedReadings returns the standard shape sensors use to report a
// degraded or unsupported condition through Readings() without failing the
// gRPC call. An empty (or nil) readings map serializes to nil over gRPC,
// which the sensor service rejects ("Readings should not return nil
// readings"), so every sensor that can end up with nothing to report needs
// to put something in the map instead of returning one empty — this is that
// something, kept in one place instead of being reimplemented per sensor.
func DegradedReadings(reason string) map[string]any {
	return map[string]any{"error": reason}
}
