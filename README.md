# viam-raspi-utils

This is a Viam Module that contains a number of sensors and utilities for the Raspberry Pi. These values are derived from the [`vcgencmd`](https://github.com/raspberrypi/documentation/blob/16480247dcac12d1f828c0f2556a3bc430de3c90/raspbian/applications/vcgencmd.md) of the Raspberry Pi.

## clocks

This sensor reports the clock frequencies of various components on the Raspberry Pi.

## cpu_manager

This is both a sensor and a configuration utility. It lets you manage the CPU frequency and governor of the Raspberry PI CPU. Please note, this will automatically install the `cpufrequtils` package using the package manager available on the system.

## power

This reports the voltages of various components on the Raspberry Pi board.

## temperature

This reports the temperature of various components on the Raspberry Pi board.

## throttling

This reports the throttling state of various components of the Raspberry Pi.
