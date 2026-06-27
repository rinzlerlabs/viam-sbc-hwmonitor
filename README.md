# viam-sbc-hwmonitor

This is a Viam Module that contains a number of sensors and utilities for single board computers (or any linux machine).

While this package strives to use no external libraries and executables, sometimes that is unavoidable. For the Raspberry Pi, some values are derived from the [`vcgencmd`](https://github.com/raspberrypi/documentation/blob/16480247dcac12d1f828c0f2556a3bc430de3c90/raspbian/applications/vcgencmd.md).

## clocks

This sensor reports the clock frequencies of various components on the SBC. For the Raspberry Pi, this requires the `vcgencmd` to be present.

## cpu_manager

This is both a sensor and a configuration utility. It lets you manage the CPU frequency and governor of the Raspberry Pi CPU. Please note, this will automatically install the CPU frequency tooling using the package manager available on the system: `linux-cpupower` on Debian Trixie and newer, or `cpufrequtils` on older releases.

All attributes are optional; only the ones you set are applied. Frequency values are in kHz.

| Attribute | Type | Required | Description |
| --- | --- | --- | --- |
| `governor` | string | No | The CPU governor to set, ex: `performance`, `powersave`, `ondemand`. Validated against the system's available governors. |
| `frequency` | int | No | A specific CPU frequency to set, in kHz. Validated against the hardware frequency limits. |
| `minimum` | int | No | The minimum CPU frequency, in kHz. |
| `maximum` | int | No | The maximum CPU frequency, in kHz. |

## cpu_monitor

This is a basic CPU monitor that reports per-core and overall usage percentages.

| Attribute | Type | Required | Description |
| --- | --- | --- | --- |
| `sleep_time_ms` | int | No | Sampling interval in milliseconds used to compute usage percentages. |

## disk_monitor

This reports disk usage and, optionally, I/O counters for the configured disks.

| Attribute | Type | Required | Description |
| --- | --- | --- | --- |
| `disks` | []string | No | The disks to report on, ex: `["sda", "nvme0n1"]`. |
| `include_io_counters` | bool | No | When `true`, include per-disk I/O counters in the readings. |

## gpu_monitor

This is a basic GPU monitor that reports per-component usage. Only currently available for NVIDIA boards.

## memory_monitor

This is a basic memory stats for the SBC.

## power_manager

This applies a power profile to the board and reports the current frequency, limits, and governor. Like `cpu_manager`, it automatically installs the CPU frequency tooling (`linux-cpupower` on Debian Trixie and newer, or `cpufrequtils` on older releases). Configuration is board specific: provide the `raspi` object on a Raspberry Pi or the `jetson` object on an NVIDIA Jetson.

| Attribute | Type | Required | Description |
| --- | --- | --- | --- |
| `raspi` | object | On Raspberry Pi | Raspberry Pi power settings. |
| `jetson` | object | On Jetson | NVIDIA Jetson power settings. |

Raspberry Pi settings (`raspi`):

| Attribute | Type | Required | Description |
| --- | --- | --- | --- |
| `governor` | string | No | The CPU governor to set, ex: `performance`, `powersave`. |
| `frequency` | int | No | A specific CPU frequency to set, in kHz. |
| `minimum` | int | No | The minimum CPU frequency, in kHz. |
| `maximum` | int | No | The maximum CPU frequency, in kHz. |

NVIDIA Jetson settings (`jetson`):

| Attribute | Type | Required | Description |
| --- | --- | --- | --- |
| `power_mode` | int | No | The `nvpmodel` power mode to set. |
| `governor` | string | No | The CPU governor to set. |
| `frequency` | int | No | A specific CPU frequency to set, in kHz. |
| `minimum` | int | No | The minimum CPU frequency, in kHz. |
| `maximum` | int | No | The maximum CPU frequency, in kHz. |

Sample config for a Raspberry Pi:

```json
{
  "raspi": {
    "governor": "performance", // optional, ex: "performance" or "powersave"
    "minimum": 600000,         // optional, minimum frequency in kHz
    "maximum": 1500000         // optional, maximum frequency in kHz
  }
}
```

Sample config for an NVIDIA Jetson:

```json
{
  "jetson": {
    "power_mode": 0,           // optional, the nvpmodel power mode
    "governor": "performance"  // optional, the CPU governor
  }
}
```

## process_monitor

This lets you monitor a specific process and get more information about the environment under which it is running. Exactly one of `name` or `executable_path` is required. Each component monitors a single process; configure additional `process_monitor` components to monitor additional processes.

| Attribute | Type | Required | Description |
| --- | --- | --- | --- |
| `name` | string | Yes* | Process name to monitor, ex: `viam-agent` or `viam-server`. |
| `executable_path` | string | Yes* | Absolute path to the executable, ex: `/opt/viam/bin/viam-agent`. |
| `include_env` | bool | No | Include the process environment variables. |
| `include_cmdline` | bool | No | Include the process command line. |
| `include_cwd` | bool | No | Include the process working directory. |
| `include_open_file_count` | bool | No | Include the count of open files. |
| `include_open_files` | bool | No | Include the list of open files. |
| `include_mem_info` | bool | No | Include memory usage information. |
| `include_ulimits` | bool | No | Include the process ulimits. |
| `include_net_stats` | bool | No | Include network statistics. |
| `sleep_time_ms` | int | No | Sleep time in milliseconds between process checks. |
| `disable_pid_caching` | bool | No | Disable caching of the resolved PID. |

`*` Exactly one of `name` or `executable_path` must be set.

Sample config monitoring `viam-server`:

```json
{
  "name": "viam-server",          // required (or use executable_path)
  "include_cmdline": true,        // optional
  "include_mem_info": true,       // optional
  "include_open_file_count": true, // optional
  "include_net_stats": true,      // optional
  "sleep_time_ms": 1000           // optional, interval between checks
}
```

## pwm_fan

This lets you control a cooling fan for the SBC based on the CPU temperatures. For the Raspberry Pi, the built-in fan is supported.

| Attribute | Type | Required | Description |
| --- | --- | --- | --- |
| `use_internal_fan` | bool | No | Use the Raspberry Pi 5's built-in fan. Only supported on Raspberry Pi 5. |
| `fan_pin` | string | Yes* | The GPIO pin controlling the fan. |
| `board_name` | string | Yes* | The board name used for pin mapping. |
| `temperature_table` | map[string]float64 | Yes | Map of temperature (°C) to fan speed (0–100). Must have at least one entry. |

`*` `fan_pin` and `board_name` are required unless `use_internal_fan` is `true`.

The `temperature_table` maps a CPU temperature (°C) to the fan speed (0–100) to use at or above that temperature.

Sample config using a GPIO-connected fan:

```json
{
  "board_name": "rpi5",  // required unless use_internal_fan is true
  "fan_pin": "33",       // required unless use_internal_fan is true
  "temperature_table": { // required, maps temperature (°C) to fan speed (0–100)
    "40": 0,
    "50": 50,
    "65": 80,
    "75": 100
  }
}
```

Sample config using the Raspberry Pi 5's built-in fan:

```json
{
  "use_internal_fan": true, // Raspberry Pi 5 only
  "temperature_table": {
    "50": 30,
    "65": 70,
    "75": 100
  }
}
```

## temperature

This reports the temperature of various temperature sensors. Available sensors vary by board.

## throttling

This reports the throttling state of various components of the SBC.

## voltages

This reports the voltages of various components on the board. The CPU voltages are generally available for all boards. Some boards also include GPU and total system power.

## wifi_monitor

This reports the status of a wifi adapter, including the connected network name and signal strength. On Linux, it uses `iw` for the most detailed stats.

| Attribute | Type | Required | Description |
| --- | --- | --- | --- |
| `adapter` | string | Yes | The wifi network interface to monitor, ex: `wlan0`. |
