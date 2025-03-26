package wifimonitor

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinuxProcWifiMonitor(t *testing.T) {
	output, err := os.ReadFile("testdata/proc.out")
	assert.NoError(t, err)
	tests := []struct {
		name           string
		adapter        string
		signalStrength int
		linkSpeed      float64
		expectedError  error
	}{
		{"AdapterExists", "wlan0", -64, 46.0, nil},
		{"AdapterDoesNotExist", "wlan1", -1, -1, ErrAdapterNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &procWifiMonitor{adapter: tt.adapter}
			status, err := w.parseNetworkStatus(string(output))
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
				return
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.signalStrength, status.SignalStrength)
				assert.Equal(t, tt.linkSpeed, status.TxSpeedMbps)
			}
		})
	}
}

func TestLinuxIwWifiMonitor(t *testing.T) {
	tests := []struct {
		name           string
		adapter        string
		signalStrength int
		linkSpeed      float64
		expectedError  error
		file           string
	}{
		{"AdapterExistsConnected", "wlan0", -65, 52.0, nil, "iw_wlan0_connected.out"},
		{"AdapterExistsNotConnected", "wlan0", -1, -1, ErrNotConnected, "iw_wlan0_not_connected.out"},
		{"AdapterDoesNotExist", "wlan1", -1, -1, ErrAdapterNotFound, "iw_wlan1_does_not_exist.out"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := os.ReadFile(fmt.Sprintf("testdata/%v", tt.file))
			assert.NoError(t, err)
			w := &iwWifiMonitor{adapter: tt.adapter}
			status, err := w.parseNetworkStatus(string(output))
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
				return
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.signalStrength, status.SignalStrength)
				assert.Equal(t, tt.linkSpeed, status.TxSpeedMbps)
			}
		})
	}
}

func TestLinuxNmcliWifiMonitor(t *testing.T) {
	output, err := os.ReadFile("testdata/nmcli.out")
	assert.NoError(t, err)
	tests := []struct {
		name           string
		adapter        string
		signalStrength int
		linkSpeed      float64
		expectedError  error
	}{
		{"AdapterExists", "wlan0", -55, 195.0, nil},
		{"AdapterExistsNotConnected", "wlan2", -1, -1, ErrNotConnected},
		{"AdapterDoesNotExist", "wlan1", -1, -1, ErrAdapterNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &nmcliWifiMonitor{adapter: tt.adapter}
			status, err := w.parseNetworkStatus(string(output))
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
				return
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.signalStrength, status.SignalStrength)
				assert.Equal(t, tt.linkSpeed, status.TxSpeedMbps)
			}
		})
	}
}
