// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"fmt"

	"github.com/intel/powertelemetry/internal/cpuid"
	"github.com/intel/powertelemetry/internal/cpumodel"
)

// MaxTurboFreq is an item of a list of max turbo frequencies and related active cores.
type MaxTurboFreq struct {
	Value       uint64 // Maximum reachable turbo frequency in MHz
	ActiveCores uint32 // Maximum number of active cores for the reachable turbo frequency
	Secondary   bool   // Attribute indicating if the list item is related to secondary cores of a hybrid architecture
}

// isHybrid points to a function that checks if CPU is hybrid.
var isHybrid = cpuid.IsCPUHybrid

// GetMaxTurboFreqList returns a list of max turbo frequencies and related active cores
// according to the package ID.
func (pt *PowerTelemetry) GetMaxTurboFreqList(packageID int) ([]MaxTurboFreq, error) {
	if pt.msr == nil {
		return nil, &ModuleNotInitializedError{Name: "msr"}
	}

	cpuID, err := pt.getCPUIDFromPackageID(packageID)
	if err != nil {
		return nil, err
	}

	model := pt.topology.getCPUModel()
	list := []MaxTurboFreq{}
	if hasHswTurboRatioLimit(model) {
		out, err := pt.dumpHswTurboRatioLimits(cpuID)
		if err != nil {
			return nil, fmt.Errorf("dump hsw: %w", err)
		}
		list = append(list, out...)
	}

	if hasIvtTurboRatioLimit(model) {
		out, err := pt.dumpIvtTurboRatioLimits(cpuID)
		if err != nil {
			return nil, fmt.Errorf("dump ivt: %w", err)
		}
		list = append(list, out...)
	}

	if hasTurboRatioLimit(model) {
		out, err := pt.dumpTurboRatioLimits(turboRatioLimit, model, cpuID)
		if err != nil {
			return nil, fmt.Errorf("dump turbo ratio limits: %w", err)
		}
		list = append(list, out...)

		if isHybrid != nil && isHybrid() {
			out, err := pt.dumpTurboRatioLimits(secondaryTurboRatioLimit, model, cpuID)
			if err != nil {
				return nil, err
			}
			list = append(list, out...)
		}
	}

	if hasAtomTurboRatioLimit(model) {
		out, err := pt.dumpAtomTurboRatioLimits(cpuID)
		if err != nil {
			return nil, fmt.Errorf("dump atom: %w", err)
		}
		list = append(list, out...)
	}

	if hasKnlTurboRatioLimit(model) {
		out, err := pt.dumpKnlTurboRatioLimits(cpuID)
		if err != nil {
			return nil, fmt.Errorf("dump knl: %w", err)
		}
		list = append(list, out...)
	}

	return list, nil
}

// hasHswTurboRatioLimit checks if the model supports the Haswell turbo ratio limit.
func hasHswTurboRatioLimit(model int) bool {
	return model == cpumodel.INTEL_FAM6_HASWELL_X // HSW Xeon
}

// hasKnlTurboRatioLimit checks if the model supports the Knights Landing turbo ratio limit.
func hasKnlTurboRatioLimit(model int) bool {
	switch model {
	case
		cpumodel.INTEL_FAM6_XEON_PHI_KNL, // Knights Landing
		cpumodel.INTEL_FAM6_XEON_PHI_KNM: // Knights Mill
		return true
	}
	return false
}

// hasIvtTurboRatioLimit checks if the model supports the Ivy Bridge turbo ratio limit.
func hasIvtTurboRatioLimit(model int) bool {
	switch model {
	case
		cpumodel.INTEL_FAM6_IVYBRIDGE_X, // IVB Xeon
		cpumodel.INTEL_FAM6_HASWELL_X:   // HSW Xeon
		return true
	}
	return false
}

// hasSlvMsrs checks if the model supports Silvermont MSRs.
func hasSlvMsrs(model int) bool {
	switch model {
	case
		cpumodel.INTEL_FAM6_ATOM_SILVERMONT,
		cpumodel.INTEL_FAM6_ATOM_SILVERMONT_MID,
		cpumodel.INTEL_FAM6_ATOM_SILVERMONT_SMARTPHONE: // INTEL_FAM6_ATOM_AIRMONT_MID in turbostat
		return true
	}
	return false
}

// hasAtomTurboRatioLimit checks if the model supports the Atom turbo ratio limit.
func hasAtomTurboRatioLimit(model int) bool {
	return hasSlvMsrs(model)
}

// hasTurboRatioLimit checks if the model has turbo ratio limit support.
func hasTurboRatioLimit(model int) bool {
	if hasSlvMsrs(model) {
		return false
	}

	switch model {
	// Nehalem compatible, but do not include turbo-ratio limit support
	case
		cpumodel.INTEL_FAM6_NEHALEM_EX, // Nehalem-EX Xeon - Beckton
		cpumodel.INTEL_FAM6_WESTMERE_EX,
		cpumodel.INTEL_FAM6_XEON_PHI_KNL, // PHI - Knights Landing (different MSR definition)
		cpumodel.INTEL_FAM6_XEON_PHI_KNM: // Knights Mill
		return false
	}

	return true
}

// hasTurboRatioGroupLimits checks if the model supports turbo ratio group limits.
func hasTurboRatioGroupLimits(model int) bool {
	switch model {
	case
		cpumodel.INTEL_FAM6_ATOM_GOLDMONT,
		cpumodel.INTEL_FAM6_SKYLAKE_X,
		cpumodel.INTEL_FAM6_ICELAKE_X,
		cpumodel.INTEL_FAM6_ICELAKE_D,
		cpumodel.INTEL_FAM6_SAPPHIRERAPIDS_X,
		cpumodel.INTEL_FAM6_EMERALDRAPIDS_X,
		cpumodel.INTEL_FAM6_ATOM_GOLDMONT_D,
		cpumodel.INTEL_FAM6_ATOM_TREMONT_D:
		return true
	default:
		return false
	}
}

// dumpHswTurboRatioLimits returns a list of max turbo frequencies and related active cores
// of a Haswell based CPU.
func (pt *PowerTelemetry) dumpHswTurboRatioLimits(cpuID int) ([]MaxTurboFreq, error) {
	msrValue, err := pt.msr.read(turboRatioLimit2, cpuID)
	if err != nil {
		return nil, fmt.Errorf("can't read MSR 0x%X: %w", turboRatioLimit2, err)
	}

	list := []MaxTurboFreq{}

	// Get two least significant octets of the 64-bit MSR value, which represent ratios, and add items with positive ratios to the list.
	ratio := (msrValue >> 8) & 0xFF
	if ratio > 0 {
		list = append(list, MaxTurboFreq{ActiveCores: 18, Value: uint64(float64(ratio) * pt.busClock)})
	}
	ratio = (msrValue >> 0) & 0xFF
	if ratio > 0 {
		list = append(list, MaxTurboFreq{ActiveCores: 17, Value: uint64(float64(ratio) * pt.busClock)})
	}

	return list, nil
}

// dumpIvtTurboRatioLimits returns a list of max turbo frequencies and related active cores
// of an Ivy Bridge based CPU.
func (pt *PowerTelemetry) dumpIvtTurboRatioLimits(cpuID int) ([]MaxTurboFreq, error) {
	msrValue, err := pt.msr.read(turboRatioLimit1, cpuID)
	if err != nil {
		return nil, fmt.Errorf("can't read MSR 0x%X: %w", turboRatioLimit1, err)
	}

	list := []MaxTurboFreq{}

	// Get 8 octets of the 64-bit MSR value, which represent ratios, and add items with positive ratios to the list.
	ratio := (msrValue >> 56) & 0xFF
	if ratio > 0 {
		list = append(list, MaxTurboFreq{ActiveCores: 16, Value: uint64(float64(ratio) * pt.busClock)})
	}
	ratio = (msrValue >> 48) & 0xFF
	if ratio > 0 {
		list = append(list, MaxTurboFreq{ActiveCores: 15, Value: uint64(float64(ratio) * pt.busClock)})
	}
	ratio = (msrValue >> 40) & 0xFF
	if ratio > 0 {
		list = append(list, MaxTurboFreq{ActiveCores: 14, Value: uint64(float64(ratio) * pt.busClock)})
	}
	ratio = (msrValue >> 32) & 0xFF
	if ratio > 0 {
		list = append(list, MaxTurboFreq{ActiveCores: 13, Value: uint64(float64(ratio) * pt.busClock)})
	}
	ratio = (msrValue >> 24) & 0xFF
	if ratio > 0 {
		list = append(list, MaxTurboFreq{ActiveCores: 12, Value: uint64(float64(ratio) * pt.busClock)})
	}
	ratio = (msrValue >> 16) & 0xFF
	if ratio > 0 {
		list = append(list, MaxTurboFreq{ActiveCores: 11, Value: uint64(float64(ratio) * pt.busClock)})
	}
	ratio = (msrValue >> 8) & 0xFF
	if ratio > 0 {
		list = append(list, MaxTurboFreq{ActiveCores: 10, Value: uint64(float64(ratio) * pt.busClock)})
	}
	ratio = (msrValue >> 0) & 0xFF
	if ratio > 0 {
		list = append(list, MaxTurboFreq{ActiveCores: 9, Value: uint64(float64(ratio) * pt.busClock)})
	}

	return list, nil
}

// dumpTurboRatioLimits returns a list of max turbo frequencies and related active cores
// of a CPU supporting turbo ratio limits.
func (pt *PowerTelemetry) dumpTurboRatioLimits(trlMsrOffset uint32, model int, cpuID int) ([]MaxTurboFreq, error) {
	msrValue, err := pt.msr.read(trlMsrOffset, cpuID)
	if err != nil {
		return nil, fmt.Errorf("can't read MSR 0x%X: %w", trlMsrOffset, err)
	}

	secondary := trlMsrOffset == secondaryTurboRatioLimit

	var coreCounts uint64
	if hasTurboRatioGroupLimits(model) {
		coreCounts, err = pt.msr.read(turboRatioLimit1, cpuID)
		if err != nil {
			return nil, fmt.Errorf("can't read MSR 0x%X: %w", turboRatioLimit1, err)
		}
	} else {
		coreCounts = 0x0807060504030201
	}

	list := []MaxTurboFreq{}

	// Iterate over 8 octets of the 64-bit MSR value and the core counts value, get the ratio and the group size,
	// then add items with positive ratios to the list.
	for shift := 56; shift >= 0; shift -= 8 {
		ratio := (msrValue >> shift) & 0xFF
		groupSize := (coreCounts >> shift) & 0xFF
		if ratio > 0 {
			list = append(list, MaxTurboFreq{
				ActiveCores: uint32(groupSize),
				Value:       uint64(float64(ratio) * pt.busClock),
				Secondary:   secondary,
			})
		}
	}

	return list, nil
}

// dumpAtomTurboRatioLimits returns a list of max turbo frequencies and related active cores
// of an Atom based CPU.
func (pt *PowerTelemetry) dumpAtomTurboRatioLimits(cpuID int) ([]MaxTurboFreq, error) {
	msrValue, err := pt.msr.read(atomCoreTurboRatios, cpuID)
	if err != nil {
		return nil, fmt.Errorf("can't read MSR 0x%X: %w", atomCoreTurboRatios, err)
	}

	list := []MaxTurboFreq{}

	// Get 4 least significant octets of the 64-bit MSR value, which represent ratios,
	// and add items with positive ratios to the list.
	ratio := (msrValue >> 24) & 0x3F
	if ratio > 0 {
		list = append(list, MaxTurboFreq{ActiveCores: 4, Value: uint64(float64(ratio) * pt.busClock)})
	}
	ratio = (msrValue >> 16) & 0x3F
	if ratio > 0 {
		list = append(list, MaxTurboFreq{ActiveCores: 3, Value: uint64(float64(ratio) * pt.busClock)})
	}
	ratio = (msrValue >> 8) & 0x3F
	if ratio > 0 {
		list = append(list, MaxTurboFreq{ActiveCores: 2, Value: uint64(float64(ratio) * pt.busClock)})
	}
	ratio = (msrValue >> 0) & 0x3F
	if ratio > 0 {
		list = append(list, MaxTurboFreq{ActiveCores: 1, Value: uint64(float64(ratio) * pt.busClock)})
	}

	return list, nil
}

// dumpKnlTurboRatioLimits returns a list of max turbo frequencies and related active cores
// of a Knights Landing based CPU.
func (pt *PowerTelemetry) dumpKnlTurboRatioLimits(cpuID int) ([]MaxTurboFreq, error) {
	msrValue, err := pt.msr.read(turboRatioLimit, cpuID)
	if err != nil {
		return nil, fmt.Errorf("can't read MSR 0x%X: %w", turboRatioLimit, err)
	}

	list := []MaxTurboFreq{}

	const bucketsNo = 7
	cores := [bucketsNo]uint64{}
	ratio := [bucketsNo]uint64{}

	bNr := 0
	cores[bNr] = (msrValue & 0xFF) >> 1 // Get maximum number of cores in Group 0
	ratio[bNr] = (msrValue >> 8) & 0xFF // Get maximum ratio limit for Group 0

	// Iterate over octets 3..8 of the 64-bit MSR value, get the number of incremental cores added to Group 1..6
	// and get the group ratio delta for Group 1..6
	for i := 16; i < 64; i += 8 {
		deltaCores := (msrValue >> i) & 0x1F
		deltaRatio := (msrValue >> (i + 5)) & 0x7

		cores[bNr+1] = cores[bNr] + deltaCores
		ratio[bNr+1] = ratio[bNr] - deltaRatio
		bNr++
	}

	// Iterate over the pairs of cores and ratios, add the first pair along with others with unique ratios to the list
	for i := bucketsNo - 1; i >= 0; i-- {
		if ((i > 0) && (ratio[i] != ratio[i-1])) || (i == 0) {
			list = append(list, MaxTurboFreq{
				ActiveCores: uint32(cores[i]),
				Value:       uint64(float64(ratio[i]) * pt.busClock),
			})
		}
	}

	return list, nil
}
