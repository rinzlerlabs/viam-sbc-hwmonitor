package wifi_monitor

type macOsWifiMonitor struct{}

func (w *macOsWifiMonitor) GetNetworkStatus() (*networkStatus, error) {
	return nil, nil
}
