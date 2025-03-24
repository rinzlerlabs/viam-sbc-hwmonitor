package wifi_monitor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	netshOutput = `
There is 1 interface on the system:

    Name                   : Wi-Fi
    Description            : Intel(R) Wi-Fi 6 AX200 160MHz
    GUID                   : f92b8a88-53ea-47b9-bd64-47f8d3197897
    Physical address       : 00:11:22:33:45:67
    Interface type         : Primary
    State                  : connected
    SSID                   : bob-wifi
    AP BSSID               : 01:23:45:67:89:ab
    Band                   : 5 GHz
    Channel                : 161
    Network type           : Infrastructure
    Radio type             : 802.11ax
    Authentication         : WPA2-Personal
    Cipher                 : CCMP
    Connection mode        : Profile
    Receive rate (Mbps)    : 865
    Transmit rate (Mbps)   : 961
    Signal                 : 83%
    Profile                : bob-wifi
    QoS MSCS Configured         : 0
    QoS Map Configured          : 0
    QoS Map Allowed by Policy   : 0

    Hosted network status  : Not available`
)

func TestNetShParsing(t *testing.T) {
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
