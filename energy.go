package main

import (
	"time"
)

// Assume a fixed power usage (in Watts) per CPU core. This is a simplified model.
// In reality, power consumption can vary significantly based on the workload and hardware.
const powerPerCoreWatt = 50.0

// EnergyEstimate estimates the energy consumption based on response time and CPU utilization.
func EnergyEstimate(responseTime time.Duration, cpuUsage float64) float64 {
	// Convert response time from nanoseconds to seconds
	timeInSeconds := responseTime.Seconds()

	// Estimate energy consumption (Joules) using the formula:
	// Energy (J) = Power (W) × Time (s)
	// Since we're using a fixed power per core, we assume CPU usage as a fraction of one core's power.
	estimatedPower := powerPerCoreWatt * cpuUsage
	energyConsumed := estimatedPower * timeInSeconds

	return energyConsumed
}
