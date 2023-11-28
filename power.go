// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"errors"
	"fmt"
	"strings"

	"github.com/intel/powertelemetry/internal/cpumodel"
)

// MSR offset definitions.
const (
	uncorePerfStatus = 0x621 // UNCORE_PERF_STATUS

	fsbFreq      = 0xCD // MSR_FSB_FREQ
	platformInfo = 0xCE // MSR_PLATFORM_INFO

	temperatureTarget = 0x1A2 // MSR_TEMPERATURE_TARGET
	thermalStatus     = 0x19C // IA32_THERM_STATUS

	c3Residency          = 0x3FC // MSR_CORE_C3_RESIDENCY
	c6Residency          = 0x3FD // MSR_CORE_C6_RESIDENCY
	c7Residency          = 0x3FE // MSR_CORE_C7_RESIDENCY
	maxFreqClockCount    = 0xE7  // IA32_MPERF
	actualFreqClockCount = 0xE8  // IA32_APERF
	timestampCounter     = 0x10  // IA32_TIME_STAMP_COUNTER

	turboRatioLimit          = 0x1AD // MSR_TURBO_RATIO_LIMIT
	turboRatioLimit1         = 0x1AE // MSR_TURBO_RATIO_LIMIT1
	turboRatioLimit2         = 0x1AF // MSR_TURBO_RATIO_LIMIT2
	secondaryTurboRatioLimit = 0x650 // MSR_SECONDARY_TURBO_RATIO_LIMIT
	atomCoreTurboRatios      = 0x66C // MSR_ATOM_CORE_TURBO_RATIOS
)

// GetInitialUncoreFrequencyMin retrieves the minimum initial uncore frequency limit (in MHz) for the specified package and die.
func (pt *PowerTelemetry) GetInitialUncoreFrequencyMin(packageID, dieID int) (float64, error) {
	if pt.uncoreFreq == nil {
		return 0.0, &ModuleNotInitializedError{Name: "uncore_frequency"}
	}
	return pt.uncoreFreq.getUncoreFrequencyMhz(packageID, dieID, "initial_min")
}

// GetCustomizedUncoreFrequencyMin retrieves the minimum customized uncore frequency limit (in MHz) for the specified package and die.
func (pt *PowerTelemetry) GetCustomizedUncoreFrequencyMin(packageID, dieID int) (float64, error) {
	if pt.uncoreFreq == nil {
		return 0.0, &ModuleNotInitializedError{Name: "uncore_frequency"}
	}
	return pt.uncoreFreq.getUncoreFrequencyMhz(packageID, dieID, "min")
}

// GetInitialUncoreFrequencyMax retrieves the maximum initial uncore frequency limit (in MHz) for the specified package and die.
func (pt *PowerTelemetry) GetInitialUncoreFrequencyMax(packageID, dieID int) (float64, error) {
	if pt.uncoreFreq == nil {
		return 0.0, &ModuleNotInitializedError{Name: "uncore_frequency"}
	}
	return pt.uncoreFreq.getUncoreFrequencyMhz(packageID, dieID, "initial_max")
}

// GetCustomizedUncoreFrequencyMax retrieves the maximum customized uncore frequency limit (in MHz) for the specified package and die.
func (pt *PowerTelemetry) GetCustomizedUncoreFrequencyMax(packageID, dieID int) (float64, error) {
	if pt.uncoreFreq == nil {
		return 0.0, &ModuleNotInitializedError{Name: "uncore_frequency"}
	}
	return pt.uncoreFreq.getUncoreFrequencyMhz(packageID, dieID, "max")
}

// GetCurrentUncoreFrequency takes a package ID and returns the current uncore frequency
// value (in MHz). First it tries to retrieve this value from sysfs. In case of error,
// it attempts to get this value from the CPU ID's MSR corresponding to the given package ID.
func (pt *PowerTelemetry) GetCurrentUncoreFrequency(packageID, dieID int) (float64, error) {
	// Get current uncore frequency value from sysfs
	if pt.uncoreFreq != nil {
		currFreq, err := pt.uncoreFreq.getUncoreFrequencyMhz(packageID, dieID, "current")
		if err == nil {
			return currFreq, nil
		}
	}

	// Fallback method to get the value via MSR
	if pt.msr == nil {
		return 0.0, &ModuleNotInitializedError{Name: "msr"}
	}

	// Get CPU ID within the package ID
	cpuID, err := pt.getCPUIDFromPackageID(packageID)
	if err != nil {
		return 0.0, err
	}

	// 32-bit [31:0] value of UNCORE_PERF_STATUS msr offset.
	res, err := pt.msr.read(uncorePerfStatus, cpuID)
	if err != nil {
		return 0.0, err
	}

	// Mask to obtain 6:0 bits corresponding to CURRENT_CLR_RATIO
	// in steps of 100 MHz. Multiplying by a factor of 100 results in MHz unit.
	return float64(res&0x3F) * 100, nil
}

// GetCPUBaseFrequency returns the base frequency of the CPU in MHz. It takes
// a package ID as an argument and calculates the frequency using msr value.
// If an error occurs, it returns 0 and the error message.
func (pt *PowerTelemetry) GetCPUBaseFrequency(packageID int) (uint64, error) {
	if pt.msr == nil {
		return 0, &ModuleNotInitializedError{Name: "msr"}
	}

	model := pt.topology.getCPUModel()
	if err := CheckIfCPUBaseFrequencySupported(model); err != nil {
		return 0, err
	}

	cpuID, err := pt.getCPUIDFromPackageID(packageID)
	if err != nil {
		return 0, fmt.Errorf("could not find CPU ID for package ID %v: %w", packageID, err)
	}

	res, err := pt.msr.read(platformInfo, cpuID)
	if err != nil {
		return 0, err
	}

	// Mask to obtain 15:8 bits corresponding to MSR_PLATFORM_INFO
	// and the nominal TSC frequency can be determined by multiplying this number by
	// the proper bus speed.
	return uint64(float64((res>>8)&0xFF) * pt.busClock), nil
}

// GetCPUFrequency returns current frequency value of specified cpu (in MHz).
func (pt *PowerTelemetry) GetCPUFrequency(cpuID int) (float64, error) {
	if pt.cpuFreq == nil {
		return 0.0, &ModuleNotInitializedError{Name: "cpu_frequency"}
	}
	return pt.cpuFreq.getCPUFrequencyMhz(cpuID)
}

// isCPUSupported returns true if the processor is supported by the library and false otherwise.
func isCPUSupported(t topologyReader) (bool, error) {
	family, err := t.getCPUFamily(0)
	if err != nil {
		return false, fmt.Errorf("error retrieving the CPU family: %w", err)
	}
	vendorID, err := t.getCPUVendor(0)
	if err != nil {
		return false, fmt.Errorf("error retrieving the CPU vendorID: %w", err)
	}
	return strings.Contains(family, "6") && strings.Contains(vendorID, "GenuineIntel"), nil
}

// GetCurrentPackagePowerConsumptionWatts takes a package ID and returns the current package domain
// power consumption package, in Watts.
func (pt *PowerTelemetry) GetCurrentPackagePowerConsumptionWatts(packageID int) (float64, error) {
	if pt.rapl == nil {
		return 0.0, &ModuleNotInitializedError{Name: "rapl"}
	}
	return pt.rapl.getCurrentPowerConsumptionWatts(packageID, packageDomain.String())
}

// GetCurrentDramPowerConsumptionWatts takes a package ID and returns the current package domain
// power consumption package, in Watts.
func (pt *PowerTelemetry) GetCurrentDramPowerConsumptionWatts(packageID int) (float64, error) {
	if pt.rapl == nil {
		return 0.0, &ModuleNotInitializedError{Name: "rapl"}
	}
	return pt.rapl.getCurrentPowerConsumptionWatts(packageID, dramDomain.String())
}

// GetPackageThermalDesignPowerWatts takes a package ID and returns its maximum allowed power, in Watts.
func (pt *PowerTelemetry) GetPackageThermalDesignPowerWatts(packageID int) (float64, error) {
	if pt.rapl == nil {
		return 0.0, &ModuleNotInitializedError{Name: "rapl"}
	}
	return pt.rapl.getMaxPowerConstraintWatts(packageID)
}

// getBusClock returns the bus clock of CPU according to its model. If calculating the
// bus clock speed requires reading of MSR, then cpuID is being used to access the appropriate
// register. If the model is unknown, then 0 is returned as the bus clock with an appropriate error.
func (pt *PowerTelemetry) getBusClock(model int) (float64, error) {
	switch model {
	case
		cpumodel.INTEL_FAM6_ATOM_SILVERMONT,
		cpumodel.INTEL_FAM6_ATOM_SILVERMONT_D,
		cpumodel.INTEL_FAM6_ATOM_SILVERMONT_MID,
		cpumodel.INTEL_FAM6_ATOM_SILVERMONT_SMARTPHONE:
		cpuID, err := pt.getFirstAvailableCPU()
		if err != nil {
			return 0.0, err
		}
		return pt.getSilvermontBusClock(cpuID)

	case cpumodel.INTEL_FAM6_ATOM_AIRMONT:
		cpuID, err := pt.getFirstAvailableCPU()
		if err != nil {
			return 0.0, err
		}
		return pt.getAirmontBusClock(cpuID)

	case
		cpumodel.INTEL_FAM6_SANDYBRIDGE,
		cpumodel.INTEL_FAM6_SANDYBRIDGE_X,
		cpumodel.INTEL_FAM6_IVYBRIDGE,
		cpumodel.INTEL_FAM6_IVYBRIDGE_X,
		cpumodel.INTEL_FAM6_HASWELL,
		cpumodel.INTEL_FAM6_HASWELL_X,
		cpumodel.INTEL_FAM6_HASWELL_L,
		cpumodel.INTEL_FAM6_HASWELL_G,
		cpumodel.INTEL_FAM6_BROADWELL,
		cpumodel.INTEL_FAM6_BROADWELL_G,
		cpumodel.INTEL_FAM6_BROADWELL_X,
		cpumodel.INTEL_FAM6_BROADWELL_D,
		cpumodel.INTEL_FAM6_SKYLAKE_L,
		cpumodel.INTEL_FAM6_SKYLAKE,
		cpumodel.INTEL_FAM6_SKYLAKE_X,
		cpumodel.INTEL_FAM6_KABYLAKE_L,
		cpumodel.INTEL_FAM6_KABYLAKE,
		cpumodel.INTEL_FAM6_COMETLAKE,
		cpumodel.INTEL_FAM6_COMETLAKE_L,
		cpumodel.INTEL_FAM6_CANNONLAKE_L,
		cpumodel.INTEL_FAM6_ICELAKE_X,
		cpumodel.INTEL_FAM6_ICELAKE_D,
		cpumodel.INTEL_FAM6_ICELAKE,
		cpumodel.INTEL_FAM6_ICELAKE_L,
		cpumodel.INTEL_FAM6_ICELAKE_NNPI,
		cpumodel.INTEL_FAM6_ROCKETLAKE,
		cpumodel.INTEL_FAM6_TIGERLAKE_L,
		cpumodel.INTEL_FAM6_TIGERLAKE,
		cpumodel.INTEL_FAM6_SAPPHIRERAPIDS_X,
		cpumodel.INTEL_FAM6_EMERALDRAPIDS_X,
		cpumodel.INTEL_FAM6_GRANITERAPIDS_X,
		cpumodel.INTEL_FAM6_LAKEFIELD,
		cpumodel.INTEL_FAM6_ALDERLAKE,
		cpumodel.INTEL_FAM6_ALDERLAKE_L,
		cpumodel.INTEL_FAM6_RAPTORLAKE,
		cpumodel.INTEL_FAM6_RAPTORLAKE_P,
		cpumodel.INTEL_FAM6_RAPTORLAKE_S,
		cpumodel.INTEL_FAM6_METEORLAKE,
		cpumodel.INTEL_FAM6_METEORLAKE_L,
		cpumodel.INTEL_FAM6_ARROWLAKE,
		cpumodel.INTEL_FAM6_LUNARLAKE_M,
		cpumodel.INTEL_FAM6_ATOM_GOLDMONT,
		cpumodel.INTEL_FAM6_ATOM_GOLDMONT_D,
		cpumodel.INTEL_FAM6_ATOM_GOLDMONT_PLUS,
		cpumodel.INTEL_FAM6_ATOM_TREMONT_D,
		cpumodel.INTEL_FAM6_ATOM_TREMONT,
		cpumodel.INTEL_FAM6_ATOM_TREMONT_L,
		cpumodel.INTEL_FAM6_ATOM_GRACEMONT,
		cpumodel.INTEL_FAM6_ATOM_CRESTMONT_X,
		cpumodel.INTEL_FAM6_ATOM_CRESTMONT,
		cpumodel.INTEL_FAM6_XEON_PHI_KNL,
		cpumodel.INTEL_FAM6_XEON_PHI_KNM:
		return 100.0, nil

	case
		cpumodel.INTEL_FAM6_NEHALEM,
		cpumodel.INTEL_FAM6_NEHALEM_G,
		cpumodel.INTEL_FAM6_NEHALEM_EP,
		cpumodel.INTEL_FAM6_NEHALEM_EX,
		cpumodel.INTEL_FAM6_WESTMERE,
		cpumodel.INTEL_FAM6_WESTMERE_EP,
		cpumodel.INTEL_FAM6_WESTMERE_EX:
		return 133.0, nil

	default:
		return 0.0, fmt.Errorf("busClock is not supported by the CPU model: %v", model)
	}
}

// getSilvermontBusClock returns busClock for Silvermont-based processors. It
// takes a cpuID argument and reads the frequency value from the MSR.
// If an error occurs during the execution, it returns 0.0 and a proper error message.
func (pt *PowerTelemetry) getSilvermontBusClock(cpuID int) (float64, error) {
	if pt.msr == nil {
		return 0.0, &ModuleNotInitializedError{Name: "msr"}
	}

	// From MSR_FSB_FREQ for Silvermont Microarchitecture
	silvermontFreqTable := []float64{83.3, 100.0, 133.3, 116.7, 80.0}
	res, err := pt.msr.read(fsbFreq, cpuID)
	if err != nil {
		return 0.0, fmt.Errorf("error while reading MSR value: %w", err)
	}
	// Since register has 3 bits we mask 0x7 or 111 to extract three least significant bits
	indx := int(res & 0x7)
	if indx >= len(silvermontFreqTable) {
		return 0.0, fmt.Errorf("error while getting bus clock: index %d is outside of bounds", indx)
	}
	return silvermontFreqTable[indx], nil
}

// getAirmontBusClock returns busClock for Airmont-based processors. It
// takes a cpuID argument and reads the frequency value from the MSR.
// If an error occurs during the execution, it returns 0.0 and a proper error message.
func (pt *PowerTelemetry) getAirmontBusClock(cpuID int) (float64, error) {
	if pt.msr == nil {
		return 0.0, &ModuleNotInitializedError{Name: "msr"}
	}

	// From MSR_FSB_FREQ for Airmont Microarchitecture
	airmontFreqTable := []float64{83.3, 100.0, 133.3, 116.7, 80.0, 93.3, 90.0, 88.9, 87.5}
	res, err := pt.msr.read(fsbFreq, cpuID)
	if err != nil {
		return 0.0, fmt.Errorf("error while reading MSR value: %w", err)
	}
	// Since register has 4 bits we mask 0xF or 1111 to extract four least significant bits
	indx := int(res & 0xF)
	if indx >= len(airmontFreqTable) {
		return 0.0, fmt.Errorf("error while getting bus clock: index %d is outside of bounds", indx)
	}
	return airmontFreqTable[indx], nil
}

// GetCPUTemperature takes a cpu ID and returns its temperature, in degrees Celsius.
// CPU temperature is calculated based on cpu-specific msr offsets:
// temp[C] = MSR_TEMPERATURE_TARGET[23:16] - IA32_THERM_STATUS[22:16]
// If an error occurs while reading msr offsets, the function returns zero value for
// the temperature and the corresponding error.
func (pt *PowerTelemetry) GetCPUTemperature(cpuID int) (uint64, error) {
	if pt.msr == nil {
		return 0, &ModuleNotInitializedError{Name: "msr"}
	}

	model := pt.topology.getCPUModel()
	if err := CheckIfCPUTemperatureSupported(model); err != nil {
		return 0, err
	}

	// 64-bit [63:0] value of MSR_TEMPERATURE_TARGET msr offset.
	res, err := pt.msr.read(uint32(temperatureTarget), cpuID)
	if err != nil {
		return 0, err
	}
	// Throttle temperature corresponds to MSR_TEMPERATURE_TARGET[23:16] in degree Celsius.
	throttleTemp := (res >> 16) & 0xFF

	// 64-bit [63:0] value of IA32_THERM_STATUS msr offset.
	res, err = pt.msr.read(uint32(thermalStatus), cpuID)
	if err != nil {
		return 0, err
	}
	// Temperature offset corresponds to IA32_THERM_STATUS[22:16] in degree Celsius.
	temp := (res >> 16) & 0x7F
	return throttleTemp - temp, nil
}

// GetCPUC0StateResidency takes a CPU ID and returns its C0 state residency metric, as a percentage.
func (pt *PowerTelemetry) GetCPUC0StateResidency(cpuID int) (float64, error) {
	if pt.msr == nil {
		return 0, &ModuleNotInitializedError{Name: "msr"}
	}

	deltas, err := pt.msr.getOffsetDeltas(cpuID)
	if err != nil {
		return 0.0, fmt.Errorf("error retrieving offset deltas for CPU ID %v: %w", cpuID, err)
	}

	mperfDelta, ok := deltas[maxFreqClockCount]
	if !ok {
		return 0.0, fmt.Errorf("mperf offset delta not found for CPU ID: %v", cpuID)
	}

	tscDelta, ok := deltas[timestampCounter]
	if !ok {
		return 0.0, fmt.Errorf("timestamp counter offset delta not found for CPU ID: %v", cpuID)
	}

	if tscDelta == 0 {
		return 0.0, fmt.Errorf("timestamp counter offset delta is zero for CPU ID: %v", cpuID)
	}
	return (float64(mperfDelta) / float64(tscDelta)) * 100, nil
}

// GetCPUC1StateResidency takes a CPU ID and returns its C1 state residency metric, as a percentage.
func (pt *PowerTelemetry) GetCPUC1StateResidency(cpuID int) (float64, error) {
	if pt.msr == nil {
		return 0, &ModuleNotInitializedError{Name: "msr"}
	}

	model := pt.topology.getCPUModel()
	if err := CheckIfCPUC1StateResidencySupported(model); err != nil {
		return 0, err
	}

	deltas, err := pt.msr.getOffsetDeltas(cpuID)
	if err != nil {
		return 0.0, fmt.Errorf("error retrieving offset deltas for CPU ID %v: %w", cpuID, err)
	}

	mperfDelta, ok := deltas[maxFreqClockCount]
	if !ok {
		return 0.0, fmt.Errorf("mperf offset delta not found for CPU ID: %v", cpuID)
	}

	c3Delta, ok := deltas[c3Residency]
	if !ok {
		return 0.0, fmt.Errorf("c3 state residency offset delta not found for CPU ID: %v", cpuID)
	}

	c6Delta, ok := deltas[c6Residency]
	if !ok {
		return 0.0, fmt.Errorf("c6 state residency offset delta not found for CPU ID: %v", cpuID)
	}

	c7Delta, ok := deltas[c7Residency]
	if !ok {
		return 0.0, fmt.Errorf("c7 state residency offset delta not found for CPU ID: %v", cpuID)
	}

	tscDelta, ok := deltas[timestampCounter]
	if !ok {
		return 0.0, fmt.Errorf("timestamp counter offset delta not found for CPU ID: %v", cpuID)
	}

	if tscDelta == 0 {
		return 0.0, fmt.Errorf("timestamp counter offset delta is zero for CPU ID: %v", cpuID)
	}
	c1Norm := float64(tscDelta-mperfDelta-c3Delta-c6Delta-c7Delta) / float64(tscDelta)
	return c1Norm * 100, nil
}

// GetCPUC3StateResidency takes a CPU ID and returns its C3 state residency metric, as a percentage.
func (pt *PowerTelemetry) GetCPUC3StateResidency(cpuID int) (float64, error) {
	if pt.msr == nil {
		return 0, &ModuleNotInitializedError{Name: "msr"}
	}

	model := pt.topology.getCPUModel()
	if err := CheckIfCPUC3StateResidencySupported(model); err != nil {
		return 0, err
	}

	deltas, err := pt.msr.getOffsetDeltas(cpuID)
	if err != nil {
		return 0.0, fmt.Errorf("error retrieving offset deltas for CPU ID %v: %w", cpuID, err)
	}

	// MSR_CORE_C3_RESIDENCY[63:0]_2 - MSR_CORE_C3_RESIDENCY[63:0]_1
	c3Delta, ok := deltas[c3Residency]
	if !ok {
		return 0.0, fmt.Errorf("c3 state residency offset delta not found for CPU ID: %v", cpuID)
	}

	// IA32_TIME_STAMP_COUNTER[63:0]_2 - IA32_TIME_STAMP_COUNTER[63:0]_1
	tscDelta, ok := deltas[timestampCounter]
	if !ok {
		return 0.0, fmt.Errorf("timestamp counter offset delta not found for CPU ID: %v", cpuID)
	}

	if tscDelta == 0 {
		return 0.0, fmt.Errorf("timestamp counter offset delta is zero for CPU ID: %v", cpuID)
	}
	return (float64(c3Delta) / float64(tscDelta)) * 100, nil
}

// GetCPUC6StateResidency takes a CPU ID and returns its C6 state residency metric, as a percentage.
// C6 state residency is calculated within a time interval and the formula is as follows:
// c6[%] = 100 *(MSR_CORE_C6_RESIDENCY_2 - MSR_CORE_C6_RESIDENCY_1) / (IA32_TIME_STAMP_COUNTER_2 - IA32_TIME_STAMP_COUNTER_1).
func (pt *PowerTelemetry) GetCPUC6StateResidency(cpuID int) (float64, error) {
	if pt.msr == nil {
		return 0, &ModuleNotInitializedError{Name: "msr"}
	}

	model := pt.topology.getCPUModel()
	if err := CheckIfCPUC6StateResidencySupported(model); err != nil {
		return 0, err
	}

	deltas, err := pt.msr.getOffsetDeltas(cpuID)
	if err != nil {
		return 0.0, fmt.Errorf("error retrieving offset deltas for CPU ID %v: %w", cpuID, err)
	}

	// MSR_CORE_C6_RESIDENCY[63:0]_2 - MSR_CORE_C6_RESIDENCY[63:0]_1
	c6Delta, ok := deltas[c6Residency]
	if !ok {
		return 0.0, fmt.Errorf("c6 state residency offset delta not found for CPU ID: %v", cpuID)
	}

	// IA32_TIME_STAMP_COUNTER[63:0]_2 - IA32_TIME_STAMP_COUNTER[63:0]_1
	tscDelta, ok := deltas[timestampCounter]
	if !ok {
		return 0.0, fmt.Errorf("timestamp counter offset delta not found for CPU ID: %v", cpuID)
	}

	if tscDelta == 0 {
		return 0.0, fmt.Errorf("timestamp counter offset delta is zero for CPU ID: %v", cpuID)
	}
	return (float64(c6Delta) / float64(tscDelta)) * 100, nil
}

// GetCPUC7StateResidency takes a CPU ID and returns its C7 state residency metric, as a percentage.
func (pt *PowerTelemetry) GetCPUC7StateResidency(cpuID int) (float64, error) {
	if pt.msr == nil {
		return 0, &ModuleNotInitializedError{Name: "msr"}
	}

	model := pt.topology.getCPUModel()
	if err := CheckIfCPUC7StateResidencySupported(model); err != nil {
		return 0, err
	}

	deltas, err := pt.msr.getOffsetDeltas(cpuID)
	if err != nil {
		return 0.0, fmt.Errorf("error retrieving offset deltas for CPU ID %v: %w", cpuID, err)
	}

	// MSR_CORE_C7_RESIDENCY[63:0]_2 - MSR_CORE_C7_RESIDENCY[63:0]_1
	c7Delta, ok := deltas[c7Residency]
	if !ok {
		return 0.0, fmt.Errorf("c7 state residency offset delta not found for CPU ID: %v", cpuID)
	}

	// IA32_TIME_STAMP_COUNTER[63:0]_2 - IA32_TIME_STAMP_COUNTER[63:0]_1
	tscDelta, ok := deltas[timestampCounter]
	if !ok {
		return 0.0, fmt.Errorf("timestamp counter offset delta not found for CPU ID: %v", cpuID)
	}

	if tscDelta == 0 {
		return 0.0, fmt.Errorf("timestamp counter offset delta is zero for CPU ID: %v", cpuID)
	}
	return (float64(c7Delta) / float64(tscDelta)) * 100, nil
}

// GetCPUBusyFrequencyMhz takes a CPU ID and returns its busy frequency metric, in MHz.
func (pt *PowerTelemetry) GetCPUBusyFrequencyMhz(cpuID int) (float64, error) {
	if pt.msr == nil {
		return 0, &ModuleNotInitializedError{Name: "msr"}
	}

	deltas, err := pt.msr.getOffsetDeltas(cpuID)
	if err != nil {
		return 0.0, fmt.Errorf("error retrieving offset deltas for CPU ID %v: %w", cpuID, err)
	}

	timestampDelta, err := pt.msr.getTimestampDelta(cpuID)
	if err != nil {
		return 0.0, fmt.Errorf("error retrieving timestamp delta for CPU ID %v: %w", cpuID, err)
	}
	if timestampDelta <= 0 {
		return 0.0, errors.New("timestamp delta must be greater than zero")
	}

	mperfDelta, ok := deltas[maxFreqClockCount]
	if !ok {
		return 0.0, fmt.Errorf("mperf offset delta not found for CPU ID: %v", cpuID)
	}
	if mperfDelta == 0 {
		return 0.0, fmt.Errorf("mperf offset delta is zero for CPU ID: %v", cpuID)
	}

	aperfDelta, ok := deltas[actualFreqClockCount]
	if !ok {
		return 0.0, fmt.Errorf("aperf offset delta not found for CPU ID: %v", cpuID)
	}

	tscDelta, ok := deltas[timestampCounter]
	if !ok {
		return 0.0, fmt.Errorf("timestamp counter offset delta not found for CPU ID: %v", cpuID)
	}

	return float64(tscDelta) * fromProcessorCyclesToHertz *
		(float64(aperfDelta) / float64(mperfDelta)) / (float64(timestampDelta.Nanoseconds()) * fromNanosecondsToSecondsRatio), nil
}

// UpdatePerCPUMetrics takes a CPU ID and updates the msr storage with offset values corresponding to
// msr file for CPU ID.
func (pt *PowerTelemetry) UpdatePerCPUMetrics(cpuID int) error {
	if pt.msr == nil {
		return &ModuleNotInitializedError{Name: "msr"}
	}
	return pt.msr.update(cpuID)
}

// IsFlagSupported takes a flag's value and returns true if first CPU supports it and false if it doesn't.
func (pt *PowerTelemetry) IsFlagSupported(flag string) (bool, error) {
	flags, err := pt.topology.getCPUFlags(0)
	if err != nil {
		return false, fmt.Errorf("error retrieving CPU flags: %w", err)
	}
	for _, f := range flags {
		if f == flag {
			return true, nil
		}
	}
	return false, nil
}

// ReadPerfEvents reads the perf events related to supported C0 state residency metrics
// and updates the storage to make metrics available. If one or more events could not be
// read an error is returned.
func (pt *PowerTelemetry) ReadPerfEvents() error {
	if pt.perf == nil {
		return &ModuleNotInitializedError{Name: "perf"}
	}
	return pt.perf.update()
}

// DeactivatePerfEvents deactivates all active events. If an event or events could not
// be successfully deactivated, an error is returned.
// This method should be explicitly called to avoid resource leakage.
func (pt *PowerTelemetry) DeactivatePerfEvents() error {
	if pt.perf == nil {
		return &ModuleNotInitializedError{Name: "perf"}
	}
	return pt.perf.deactivate()
}

// GetCPUC0SubstateC01Percent takes a CPU ID and returns a value indicating the percentage of time
// the processor spent in its C0.1 substate out of the total time in the C0 state.
// C0.1 is characterized by a light-weight slower wakeup time but more power-saving optimized state.
func (pt *PowerTelemetry) GetCPUC0SubstateC01Percent(cpuID int) (float64, error) {
	return pt.getPerfMetricRatio(cpuID, c01.String(), thread.String())
}

// GetCPUC0SubstateC02Percent takes a CPU ID and returns a value indicating the percentage of time
// the processor spent in its C0.2 substate out of the total time in the C0 state.
// C0.2 is characterized by a light-weight faster wakeup time but less power saving optimized state.
func (pt *PowerTelemetry) GetCPUC0SubstateC02Percent(cpuID int) (float64, error) {
	return pt.getPerfMetricRatio(cpuID, c02.String(), thread.String())
}

// GetCPUC0SubstateC0WaitPercent takes a CPU ID and returns a value indicating the percentage of time
// the processor spent in its C0_Wait substate out of the total time in the C0 state.
// CPU is in C0_Wait substate when the thread is in the C0.1 or C0.2 or running a PAUSE in C0 ACPI state.
func (pt *PowerTelemetry) GetCPUC0SubstateC0WaitPercent(cpuID int) (float64, error) {
	return pt.getPerfMetricRatio(cpuID, c0Wait.String(), thread.String())
}

// getPerfMetricRatio is a helper method that takes a CPU ID, a target metric name and reference metric name.
// First, it fetches the specified metrics from the perf storage. Then, it calculates the percentage of the target
// metric, with respect to the reference metric.
func (pt *PowerTelemetry) getPerfMetricRatio(cpuID int, target, reference string) (float64, error) {
	if pt.perf == nil {
		return 0.0, &ModuleNotInitializedError{Name: "perf"}
	}
	coreMetrics := pt.perf.getCoreMetrics(cpuID)
	if len(coreMetrics) == 0 {
		return 0.0, fmt.Errorf("no core metrics found for CPU ID: %v", cpuID)
	}

	targetMetric, err := getMetric(coreMetrics, target)
	if err != nil {
		return 0.0, err
	}

	refMetric, err := getMetric(coreMetrics, reference)
	if err != nil {
		return 0.0, err
	}

	if refMetric.scaled == 0 {
		return 0.0, fmt.Errorf("zero scaled value for reference metric: %q", reference)
	}

	return float64(targetMetric.scaled) / float64(refMetric.scaled) * 100, nil
}

// getMetric is a helper function that takes a slice of coreMetrics and a string name,
// and returns the first coreMetric corresponding to the name specified.
func getMetric(metrics []coreMetric, name string) (coreMetric, error) {
	for _, metric := range metrics {
		if strings.Contains(metric.name, name) {
			return metric, nil
		}
	}
	return coreMetric{}, fmt.Errorf("could not find metric: %q", name)
}

// getCPUIDFromPackageID takes a package ID and returns a CPU ID within that package ID
// that can be used to read msr values from.
func (pt *PowerTelemetry) getCPUIDFromPackageID(packageID int) (int, error) {
	for _, cpu := range pt.cpus {
		pkgID, _ := pt.topology.getCPUPackageID(cpu)
		if pkgID == packageID {
			return cpu, nil
		}
	}
	return 0, fmt.Errorf("unable to get CPU ID for package ID: %v", packageID)
}

// getFirstAvailableCPU returns the first CPU ID from the slice of available CPUs
// for which msr can be accessed. If no CPUs are available it returns an error.
func (pt *PowerTelemetry) getFirstAvailableCPU() (int, error) {
	if len(pt.cpus) == 0 {
		return 0, errors.New("no available CPUs were found")
	}
	return pt.cpus[0], nil
}

// GetPackageIDs returns a slice with ordered package IDs of the host.
func (pt *PowerTelemetry) GetPackageIDs() []int {
	return pt.topology.getPackageIDs()
}

// GetRaplPackageIDs returns a slice with package IDs of the host for which rapl has access to.
// If rapl is not initialized, it returns nil.
func (pt *PowerTelemetry) GetRaplPackageIDs() []int {
	if pt.rapl == nil {
		return nil
	}
	return pt.rapl.getPackageIDs()
}

// GetMsrCPUIDs returns a slice with available CPU IDs of the host, for which msr has access to.
func (pt *PowerTelemetry) GetMsrCPUIDs() []int {
	return pt.cpus
}

// GetPerfCPUIDs returns a slice with available CPU IDs of the host, for which perf has access to.
func (pt *PowerTelemetry) GetPerfCPUIDs() []int {
	// TODO: This implementation should be changed when this library will
	// support hybrid CPUs. Only performance cores should be returned here
	// so that the result may be a subset of slice pt.cpus.
	return pt.cpus
}

// GetCPUPackageID gets cpu's package ID value. If no cpu is found for the corresponding cpuID
// an error is returned.
func (pt *PowerTelemetry) GetCPUPackageID(cpuID int) (int, error) {
	packageID, err := pt.topology.getCPUPackageID(cpuID)
	if err != nil {
		return 0, fmt.Errorf("error retrieving package ID: %w", err)
	}

	return packageID, nil
}

// GetCPUCoreID gets cpu's core ID value. If no cpu is found for the corresponding cpuID
// an error is returned.
func (pt *PowerTelemetry) GetCPUCoreID(cpuID int) (int, error) {
	coreID, err := pt.topology.getCPUCoreID(cpuID)
	if err != nil {
		return 0, fmt.Errorf("error retrieving core ID: %w", err)
	}

	return coreID, nil
}

// GetPackageDieIDs gets package's die ID values. If no package is found for the corresponding packageID
// an error is returned.
func (pt *PowerTelemetry) GetPackageDieIDs(packageID int) ([]int, error) {
	dies, err := pt.topology.getPackageDieIDs(packageID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving dies: %w", err)
	}

	return dies, nil
}
