// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"fmt"

	"github.com/intel/powertelemetry/internal/cpumodel"
)

// CheckIfCPUC1StateResidencySupported checks if CPU C1 state residency metric is supported by CPU model.
// Returns MetricNotSupportedError if metric is not supported by the CPU model; otherwise, returns nil.
func CheckIfCPUC1StateResidencySupported(cpuModel int) error {
	if !isC1C6BaseTempSupported(cpuModel) {
		return &MetricNotSupportedError{fmt.Sprintf("c1 state residency metric not supported by CPU model: 0x%X", cpuModel)}
	}

	return nil
}

// CheckIfCPUC3StateResidencySupported checks if CPU C3 state residency metric is supported by CPU model.
// Returns MetricNotSupportedError if metric is not supported by the CPU model; otherwise, returns nil.
func CheckIfCPUC3StateResidencySupported(cpuModel int) error {
	if !isC3Supported(cpuModel) {
		return &MetricNotSupportedError{fmt.Sprintf("c3 state residency metric not supported by CPU model: 0x%X", cpuModel)}
	}

	return nil
}

// CheckIfCPUC6StateResidencySupported checks if CPU C6 state residency metric is supported by CPU model.
// Returns MetricNotSupportedError if metric is not supported by the CPU model; otherwise, returns nil.
func CheckIfCPUC6StateResidencySupported(cpuModel int) error {
	if !isC1C6BaseTempSupported(cpuModel) {
		return &MetricNotSupportedError{fmt.Sprintf("c6 state residency metric not supported by CPU model: 0x%X", cpuModel)}
	}

	return nil
}

// CheckIfCPUC7StateResidencySupported checks if CPU C7 state residency metric is supported by CPU model.
// Returns MetricNotSupportedError if metric is not supported by the CPU model; otherwise, returns nil.
func CheckIfCPUC7StateResidencySupported(cpuModel int) error {
	if !isC7Supported(cpuModel) {
		return &MetricNotSupportedError{fmt.Sprintf("c7 state residency metric not supported by CPU model: 0x%X", cpuModel)}
	}

	return nil
}

// CheckIfCPUBaseFrequencySupported checks if CPU base frequency metric is supported by CPU model.
// Returns MetricNotSupportedError if metric is not supported by the CPU model; otherwise, returns nil.
func CheckIfCPUBaseFrequencySupported(cpuModel int) error {
	if !isC1C6BaseTempSupported(cpuModel) {
		return &MetricNotSupportedError{fmt.Sprintf("cpu base frequency metric not supported by CPU model: 0x%X", cpuModel)}
	}

	return nil
}

// CheckIfCPUTemperatureSupported checks if CPU temperature metric is supported by CPU model.
// Returns MetricNotSupportedError if metric is not supported by the CPU model; otherwise, returns nil.
func CheckIfCPUTemperatureSupported(cpuModel int) error {
	if !isC1C6BaseTempSupported(cpuModel) {
		return &MetricNotSupportedError{fmt.Sprintf("cpu temperature metric not supported by CPU model: 0x%X", cpuModel)}
	}

	return nil
}

func isC1C6BaseTempSupported(cpuModel int) bool {
	switch cpuModel {
	case
		cpumodel.INTEL_FAM6_NEHALEM,
		cpumodel.INTEL_FAM6_NEHALEM_G,
		cpumodel.INTEL_FAM6_NEHALEM_EP,
		cpumodel.INTEL_FAM6_NEHALEM_EX,
		cpumodel.INTEL_FAM6_WESTMERE,
		cpumodel.INTEL_FAM6_WESTMERE_EP,
		cpumodel.INTEL_FAM6_WESTMERE_EX,
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
		cpumodel.INTEL_FAM6_GRANITERAPIDS_D,
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
		cpumodel.INTEL_FAM6_ATOM_SILVERMONT,
		cpumodel.INTEL_FAM6_ATOM_SILVERMONT_D,
		cpumodel.INTEL_FAM6_ATOM_SILVERMONT_MID,
		cpumodel.INTEL_FAM6_ATOM_SILVERMONT_SMARTPHONE,
		cpumodel.INTEL_FAM6_ATOM_AIRMONT,
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
		return true
	}
	return false
}

func isC3Supported(cpuModel int) bool {
	switch cpuModel {
	case
		cpumodel.INTEL_FAM6_NEHALEM,
		cpumodel.INTEL_FAM6_NEHALEM_G,
		cpumodel.INTEL_FAM6_NEHALEM_EP,
		cpumodel.INTEL_FAM6_NEHALEM_EX,
		cpumodel.INTEL_FAM6_WESTMERE,
		cpumodel.INTEL_FAM6_WESTMERE_EP,
		cpumodel.INTEL_FAM6_WESTMERE_EX,
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
		cpumodel.INTEL_FAM6_KABYLAKE_L,
		cpumodel.INTEL_FAM6_KABYLAKE,
		cpumodel.INTEL_FAM6_COMETLAKE,
		cpumodel.INTEL_FAM6_COMETLAKE_L,
		cpumodel.INTEL_FAM6_ATOM_AIRMONT,
		cpumodel.INTEL_FAM6_ATOM_GOLDMONT,
		cpumodel.INTEL_FAM6_ATOM_GOLDMONT_PLUS:
		return true
	}
	return false
}

func isC7Supported(cpuModel int) bool {
	switch cpuModel {
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
		cpumodel.INTEL_FAM6_SKYLAKE_L,
		cpumodel.INTEL_FAM6_SKYLAKE,
		cpumodel.INTEL_FAM6_KABYLAKE_L,
		cpumodel.INTEL_FAM6_KABYLAKE,
		cpumodel.INTEL_FAM6_COMETLAKE,
		cpumodel.INTEL_FAM6_COMETLAKE_L,
		cpumodel.INTEL_FAM6_CANNONLAKE_L,
		cpumodel.INTEL_FAM6_ICELAKE_L,
		cpumodel.INTEL_FAM6_ICELAKE_NNPI,
		cpumodel.INTEL_FAM6_ROCKETLAKE,
		cpumodel.INTEL_FAM6_TIGERLAKE_L,
		cpumodel.INTEL_FAM6_TIGERLAKE,
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
		cpumodel.INTEL_FAM6_ATOM_GOLDMONT_PLUS,
		cpumodel.INTEL_FAM6_ATOM_TREMONT,
		cpumodel.INTEL_FAM6_ATOM_TREMONT_L,
		cpumodel.INTEL_FAM6_ATOM_GRACEMONT:
		return true
	}
	return false
}
