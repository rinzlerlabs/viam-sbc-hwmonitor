# viam-raspi-utils

This is a Viam Module that contains a number of sensors and utilities for single board computers (or any linux machine). 

While this package strives to use no external libraries and executables, sometimes that is unavoidable. For the Raspberry Pi, some values are derived from the [`vcgencmd`](https://github.com/raspberrypi/documentation/blob/16480247dcac12d1f828c0f2556a3bc430de3c90/raspbian/applications/vcgencmd.md).

## clocks

This sensor reports the clock frequencies of various components on the SBC. For the Raspberry Pi, this requires the `vcgencmd` to be present.

## cpu_manager

This is both a sensor and a configuration utility. It lets you manage the CPU frequency and governor of the Raspberry PI CPU. Please note, this will automatically install the `cpufrequtils` package using the package manager available on the system.

## cpu_monitor

This is a basic CPU monitor that reports per-core and overall usage percentages.

## gpu_monitor

This is a basic GPU monitor that reports per-component usage. Only currently available for NVIDIA boards.

## memory_monitor

This is a basic memory stats for the SBC.

## process_monitor

This lets you monitor a specific process and get more information about the environment under which it is running.

Sample Config
```json
{
  "include_open_files": <true|false>,
  "name": "<name>", // ex: "viam-agent" or "viam-server"
  "executable_path": "/absolute/path/to/executable" // ex: "/opt/viam/bin/viam-agent"
  "include_env": <true|false>,
  "include_cmdline": <true|false>,
  "include_ulimits": <true|false>,
  "include_cwd": <true|false>,
  "include_net_stats": <true|false>,
  "include_open_file_count": <true|false>,
  "include_mem_info": <true|false>
}
```

## pwm_fan

This lets you control a cooling fan for the SBC based on the CPU temperatures. For the RaspberryPi, the built-in fan is supported.

## temperature

This reports the temperature of various temperature sensors. Available sensors vary by board.

## throttling

This reports the throttling state of various components of the SBC.

## voltages

This reports the voltages of various components on the board. The CPU voltages are generally available for all boards. Some boards also include GPU and total system power.
