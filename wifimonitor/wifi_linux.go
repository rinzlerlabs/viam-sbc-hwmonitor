package wifimonitor

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"go.viam.com/rdk/logging"
)

func (c *Config) newWifiMonitor(adapter string) WifiMonitor {
	// iw has the best stats
	if _, err := exec.LookPath("iw"); err == nil {
		c.logger.Infof("Using iw for wifi stats")
		return &iwWifiMonitor{adapter: adapter, logger: c.logger}
	}
	// nmcli has good stats
	if _, err := exec.LookPath("nmcli"); err == nil {
		c.logger.Infof("Using nmcli for wifi stats")
		return &nmcliWifiMonitor{adapter: adapter, logger: c.logger}
	}
	// proc has basic stats
	if _, err := os.Stat("/proc/net/wireless"); err == nil {
		c.logger.Infof("Using /proc/net/wireless for wifi stats")
		return &procWifiMonitor{adapter: adapter, logger: c.logger}
	}
	return nil
}

type nmcliWifiMonitor struct {
	logger  logging.Logger
	adapter string
}

func (w *nmcliWifiMonitor) GetNetworkStatus() (*networkStatus, error) {
	cmd := exec.Command("nmcli", "-t", "-f", "ACTIVE,NAME,SSID,CHAN,FREQ,RATE,SIGNAL,DEVICE", "dev", "wifi")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return w.parseNetworkStatus(string(out))
}

func (w *nmcliWifiMonitor) parseNetworkStatus(out string) (*networkStatus, error) {
	adapterFound := false
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if !strings.HasSuffix(line, w.adapter) {
			continue
		}
		adapterFound = true
		if strings.HasPrefix(line, "yes:") {
			var e error = nil
			col := strings.Split(line, ":")
			signalStrength, err := strconv.Atoi(col[6])
			if err != nil {
				signalStrength = -1
				e = errors.Join(e, err)
			}

			linkSpeed, err := strconv.ParseFloat(strings.Split(col[5], " ")[0], 64)
			if err != nil {
				linkSpeed = -1
				e = errors.Join(e, err)
			}

			return &networkStatus{
				NetworkName:    col[2],
				SignalStrength: -1 * signalStrength,
				TxSpeedMbps:    linkSpeed,
			}, e
		}
	}
	if !adapterFound {
		return nil, ErrAdapterNotFound
	} else {
		return nil, ErrNotConnected
	}
}

type iwWifiMonitor struct {
	logger  logging.Logger
	adapter string
}

func (w *iwWifiMonitor) GetNetworkStatus() (*networkStatus, error) {
	cmd := exec.Command("iw", "dev", w.adapter, "link")
	out, err := cmd.Output()
	if err != nil {
		if err.Error() == "exit status 237" {
			return nil, ErrAdapterNotFound
		}
		return nil, err
	}

	return w.parseNetworkStatus(string(out))
}

func (w *iwWifiMonitor) parseNetworkStatus(out string) (*networkStatus, error) {
	var e error = nil
	if strings.Contains(string(out), "Not connected") {
		return nil, ErrNotConnected
	}
	if strings.Contains(string(out), "No such device") {
		return nil, ErrAdapterNotFound
	}
	status := &networkStatus{}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SSID:") {
			col := strings.Split(line, ":")
			status.NetworkName = strings.TrimSpace(col[1])
		} else if strings.HasPrefix(line, "signal:") {
			col := strings.Split(line, ":")
			signalStrength, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSpace(col[1]), " dBm"))
			if err != nil {
				signalStrength = -1
				e = errors.Join(e, err)
			}
			status.SignalStrength = signalStrength
		} else if strings.HasPrefix(line, "rx bitrate:") {
			col := strings.Split(line, ":")
			linkSpeed, err := strconv.ParseFloat(strings.Split(col[1], " ")[1], 64)
			if err != nil {
				linkSpeed = -1
				e = errors.Join(e, err)
			}
			status.TxSpeedMbps = linkSpeed
		}
	}

	return status, e
}

type procWifiMonitor struct {
	logger  logging.Logger
	adapter string
}

func (w *procWifiMonitor) GetNetworkStatus() (*networkStatus, error) {
	out, err := os.ReadFile("/proc/net/wireless")
	if err != nil {
		return nil, err
	}
	return w.parseNetworkStatus(string(out))
}

func (w *procWifiMonitor) parseNetworkStatus(out string) (*networkStatus, error) {
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, w.adapter) {
			col := strings.Fields(line)
			signalStrength, err := strconv.Atoi(strings.TrimSuffix(col[3], "."))
			if err != nil {
				return nil, err
			}
			linkSpeed, err := strconv.ParseFloat(col[2], 64)
			if err != nil {
				return nil, err
			}
			return &networkStatus{
				NetworkName:    "unknown",
				SignalStrength: signalStrength,
				TxSpeedMbps:    linkSpeed,
			}, nil
		}
	}
	return nil, ErrAdapterNotFound
}
