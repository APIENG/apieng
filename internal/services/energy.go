package services

import (
	"time"
)

// LIMITATION: This energy model is a placeholder and does not measure real energy consumption.
// cpuUsage is always hardcoded to 0.5 in measure.go and ai.go, making energy a linear function of response time.
// Energy = 25 * responseTime (in seconds), which is just a restatement of latency, not actual power draw.
//
// Real energy measurement would require:
// - Actual CPU utilization measurement (via /proc/stat, cgroups, or hardware counters)
// - Memory bandwidth usage
// - I/O operations and device power states
// - Server-side power monitoring, not just client-side round-trip time
//
// This model only measures the client's round-trip time and has no relationship to the server's actual energy draw.
const powerPerCoreWatt = 50.0

// EnergyEstimate estimates the energy consumption based on response time and CPU utilization.
// WARNING: This is a placeholder implementation. See package comments for limitations.
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
