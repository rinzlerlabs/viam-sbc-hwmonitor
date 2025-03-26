package wifimonitor

import "errors"

var (
	ErrNotConnected    = errors.New("not connected to a network")
	ErrAdapterNotFound = errors.New("adapter not found")
	ErrNoAdaptersFound = errors.New("no adapters found")
)

type WifiMonitor interface {
	GetNetworkStatus() (*networkStatus, error)
}

type networkStatus struct {
	NetworkName    string
	SignalStrength int
	TxSpeedMbps    float64
	RxSpeedMbps    float64
}
