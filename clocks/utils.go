package clocks

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func getRasPiClockFrequencies() (Arm, Core, H264, ISP, V3D, UART, PWM, EMMC, Pixel, Vec, HDMI, DPI int, Err error) {
	proc := exec.Command("vcgencmd", "measure_clock", "arm", "core", "h264", "isp", "v3d", "uart", "pwm", "emmc", "pixel", "vec", "hdmi", "dpi")
	outputBytes, err := proc.Output()
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}
	output := strings.Split(string(outputBytes), "\n")
	arm, err := parseRasPiClockFrequency(output[0])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	core, err := parseRasPiClockFrequency(output[1])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	h264, err := parseRasPiClockFrequency(output[2])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	isp, err := parseRasPiClockFrequency(output[3])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	v3d, err := parseRasPiClockFrequency(output[4])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	uart, err := parseRasPiClockFrequency(output[5])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	pwm, err := parseRasPiClockFrequency(output[6])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	emmc, err := parseRasPiClockFrequency(output[7])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	pixel, err := parseRasPiClockFrequency(output[8])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	vec, err := parseRasPiClockFrequency(output[9])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	hdmi, err := parseRasPiClockFrequency(output[10])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	dpi, err := parseRasPiClockFrequency(output[11])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	return arm, core, h264, isp, v3d, uart, pwm, emmc, pixel, vec, hdmi, dpi, nil
}

func parseRasPiClockFrequency(clock string) (int, error) {
	t := strings.TrimSpace(clock)
	if !strings.HasPrefix(t, "frequency") {
		return 0, fmt.Errorf("unexpected clock frequency line: %s", clock)
	}
	parts := strings.Split(t, "=")
	if len(parts) != 2 {
		return 0, fmt.Errorf("unexpected clock frequency line: %s", clock)
	}
	freq, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, fmt.Errorf("unexpected clock frequency line: %s", clock)
	}
	return freq, nil
}
