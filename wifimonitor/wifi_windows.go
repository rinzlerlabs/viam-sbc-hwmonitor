package wifimonitor

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"go.viam.com/rdk/logging"
)

func (c *Config) newWifiMonitor(adapter string) WifiMonitor {
	return &wifiMonitor{adapter: adapter, logger: c.logger}
}

type wifiMonitor struct {
	adapter string
	logger  logging.Logger
}

func (w *wifiMonitor) GetNetworkStatus() (*networkStatus, error) {
	cmd := exec.Command("netsh", "wlan", "show", "interfaces")
	out, err := cmd.Output()
	if err != nil {
		return nil, errors.Join(err, errors.New("error running command"))
	}

	networkStatuses, err := w.parseNetworkStatus(out)
	if err != nil {
		return nil, err
	}
	if len(networkStatuses) == 0 {
		return nil, errors.Join(ErrNotConnected, errors.New("no network status found"))
	}
	for _, status := range networkStatuses {
		if strings.EqualFold(status.NetworkName, w.adapter) {
			return &status, nil
		}
	}
	// If we reach here, it means we found no matching adapter
	return nil, fmt.Errorf("no active WiFi adapter found")
}

func (w *wifiMonitor) parseNetworkStatus(output []byte) ([]networkStatus, error) {
	networkStatuses := make([]networkStatus, 0)
	reader := bytes.NewReader(output)
	scanner := bufio.NewScanner(reader)
	var ssid, adapterName, signalStrength, txSpeed, rxSpeed string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Name") {
			adapterName = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "SSID") {
			ssid = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "Signal") {
			signalStrength = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "Receive rate (Mbps)") {
			txSpeed = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "Transmit rate (Mbps)") {
			rxSpeed = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if ssid != "" && adapterName != "" && signalStrength != "" && txSpeed != "" {
			signal, err := strconv.Atoi(strings.TrimSuffix(signalStrength, "%"))
			if err != nil {
				return nil, fmt.Errorf("error parsing signal strength: %w", err)
			}
			txSpeedMbps, err := strconv.ParseFloat(txSpeed, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing link speed: %w", err)
			}
			rxSpeedMbps, err := strconv.ParseFloat(rxSpeed, 64)
			if err != nil {
				return nil, fmt.Errorf("error parsing receive speed: %w", err)
			}
			networkStatuses = append(networkStatuses, networkStatus{
				NetworkName:    ssid,
				SignalStrength: signal,
				TxSpeedMbps:    txSpeedMbps,
				RxSpeedMbps:    rxSpeedMbps,
			})
		}
	}
	return networkStatuses, scanner.Err()
}
