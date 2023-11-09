package clocks

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func getClockFrequencies() (Arm, Core, H264, ISP, V3D, UART, PWM, EMMC, Pixel, Vec, HDMI, DPI int, Err error) {
	proc := exec.Command("vcgencmd", "measure_clock", "arm", "core", "h264", "isp", "v3d", "uart", "pwm", "emmc", "pixel", "vec", "hdmi", "dpi")
	outputBytes, err := proc.Output()
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}
	output := strings.Split(string(outputBytes), "\n")
	arm, err := parseClockFrequency(output[0])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	core, err := parseClockFrequency(output[1])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	h264, err := parseClockFrequency(output[2])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	isp, err := parseClockFrequency(output[3])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	v3d, err := parseClockFrequency(output[4])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	uart, err := parseClockFrequency(output[5])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	pwm, err := parseClockFrequency(output[6])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	emmc, err := parseClockFrequency(output[7])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	pixel, err := parseClockFrequency(output[8])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	vec, err := parseClockFrequency(output[9])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	hdmi, err := parseClockFrequency(output[10])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	dpi, err := parseClockFrequency(output[11])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	return arm, core, h264, isp, v3d, uart, pwm, emmc, pixel, vec, hdmi, dpi, nil
}

func parseClockFrequency(clock string) (int, error) {
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
