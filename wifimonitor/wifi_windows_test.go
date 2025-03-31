package wifimonitor

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNetShParsing(t *testing.T) {
	netshOutput, err := os.ReadFile("testdata/netsh_output.txt")
	require.NoError(t, err, "Failed to read netsh output file")
	monitor := &wifiMonitor{adapter: "Wi-Fi"}
	statuses, err := monitor.parseNetworkStatus([]byte(netshOutput))
	if err != nil {
		t.Fatalf("Failed to parse network status: %v", err)
	}

	require.Len(t, statuses, 1, "Expected one network status")
	status := statuses[0]
	require.Equal(t, "bob-wifi", status.NetworkName, "Expected network name 'bob-wifi'")
	require.Equal(t, 83, status.SignalStrength, "Expected signal strength 83")
	require.Equal(t, 865.0, status.RxSpeedMbps, "Expected Rx speed 865 Mbps")
	require.Equal(t, 961.0, status.TxSpeedMbps, "Expected Tx speed 961 Mbps")
}
