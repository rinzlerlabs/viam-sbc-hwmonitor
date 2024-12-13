package wifi_monitor

import "errors"

var (
	ErrNotConnected    = errors.New("not connected to a network")
	ErrAdapterNotFound = errors.New("adapter not found")
	ErrNoAdaptersFound = errors.New("no adapters found")
)

type wifiMonitor interface {
	GetNetworkStatus() (*networkStatus, error)
}

type networkStatus struct {
	NetworkName    string
	SignalStrength int
	LinkSpeedMbps  float64
}
