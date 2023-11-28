// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckIfCPUC1StateResidencySupported(t *testing.T) {
	m := make(map[int]interface{})
	for _, v := range c1c6BaseTempModels {
		m[v] = struct{}{}
	}

	for model := 0; model < 0xFF; model++ {
		err := CheckIfCPUC1StateResidencySupported(model)
		if m[model] != nil {
			require.NoError(t, err, "CPU model 0x%X should support c1 state residency", model)
		} else {
			require.ErrorContains(t, err, fmt.Sprintf("c1 state residency metric not supported by CPU model: 0x%X", model),
				"CPU model 0x%X shouldn't support c1 state residency", model)
		}
	}
}

func TestCheckIfCPUC3StateResidencySupported(t *testing.T) {
	m := make(map[int]interface{})
	for _, v := range c3Models {
		m[v] = struct{}{}
	}

	for model := 0; model < 0xFF; model++ {
		err := CheckIfCPUC3StateResidencySupported(model)
		if m[model] != nil {
			require.NoError(t, err, "CPU model 0x%X should support c3 state residency", model)
		} else {
			require.ErrorContains(t, err, fmt.Sprintf("c3 state residency metric not supported by CPU model: 0x%X", model),
				"CPU model 0x%X shouldn't support c3 state residency", model)
		}
	}
}

func TestCheckIfCPUC6StateResidencySupported(t *testing.T) {
	m := make(map[int]interface{})
	for _, v := range c1c6BaseTempModels {
		m[v] = struct{}{}
	}

	for model := 0; model < 0xFF; model++ {
		err := CheckIfCPUC6StateResidencySupported(model)
		if m[model] != nil {
			require.NoError(t, err, "CPU model 0x%X should support c6 state residency", model)
		} else {
			require.ErrorContains(t, err, fmt.Sprintf("c6 state residency metric not supported by CPU model: 0x%X", model),
				"CPU model 0x%X shouldn't support c6 state residency", model)
		}
	}
}

func TestCheckIfCPUC7StateResidencySupported(t *testing.T) {
	m := make(map[int]interface{})
	for _, v := range c7Models {
		m[v] = struct{}{}
	}

	for model := 0; model < 0xFF; model++ {
		err := CheckIfCPUC7StateResidencySupported(model)
		if m[model] != nil {
			require.NoError(t, err, "CPU model 0x%X should support c7 state residency", model)
		} else {
			require.ErrorContains(t, err, fmt.Sprintf("c7 state residency metric not supported by CPU model: 0x%X", model),
				"CPU model 0x%X shouldn't support c7 state residency", model)
		}
	}
}

func TestCheckIfCPUBaseFrequencySupported(t *testing.T) {
	m := make(map[int]interface{})
	for _, v := range c1c6BaseTempModels {
		m[v] = struct{}{}
	}

	for model := 0; model < 0xFF; model++ {
		err := CheckIfCPUBaseFrequencySupported(model)
		if m[model] != nil {
			require.NoError(t, err, "CPU model 0x%X should support cpu base frequency", model)
		} else {
			require.ErrorContains(t, err, fmt.Sprintf("cpu base frequency metric not supported by CPU model: 0x%X", model),
				"CPU model 0x%X shouldn't support cpu base frequency", model)
		}
	}
}

func TestCheckIfCPUTemperatureSupported(t *testing.T) {
	m := make(map[int]interface{})
	for _, v := range c1c6BaseTempModels {
		m[v] = struct{}{}
	}

	for model := 0; model < 0xFF; model++ {
		err := CheckIfCPUTemperatureSupported(model)
		if m[model] != nil {
			require.NoError(t, err, "CPU model 0x%X should support cpu temperature", model)
		} else {
			require.ErrorContains(t, err, fmt.Sprintf("cpu temperature metric not supported by CPU model: 0x%X", model),
				"CPU model 0x%X shouldn't support cpu temperature", model)
		}
	}
}

var (
	c1c6BaseTempModels = []int{
		0x1E, // INTEL_FAM6_NEHALEM
		0x1F, // INTEL_FAM6_NEHALEM_G
		0x1A, // INTEL_FAM6_NEHALEM_EP
		0x2E, // INTEL_FAM6_NEHALEM_EX
		0x25, // INTEL_FAM6_WESTMERE
		0x2C, // INTEL_FAM6_WESTMERE_EP
		0x2F, // INTEL_FAM6_WESTMERE_EX
		0x2A, // INTEL_FAM6_SANDYBRIDGE
		0x2D, // INTEL_FAM6_SANDYBRIDGE_X
		0x3A, // INTEL_FAM6_IVYBRIDGE
		0x3E, // INTEL_FAM6_IVYBRIDGE_X
		0x3C, // INTEL_FAM6_HASWELL
		0x3F, // INTEL_FAM6_HASWELL_X
		0x45, // INTEL_FAM6_HASWELL_L
		0x46, // INTEL_FAM6_HASWELL_G
		0x3D, // INTEL_FAM6_BROADWELL
		0x47, // INTEL_FAM6_BROADWELL_G
		0x4F, // INTEL_FAM6_BROADWELL_X
		0x56, // INTEL_FAM6_BROADWELL_D
		0x4E, // INTEL_FAM6_SKYLAKE_L
		0x5E, // INTEL_FAM6_SKYLAKE
		0x55, // INTEL_FAM6_SKYLAKE_X
		0x8E, // INTEL_FAM6_KABYLAKE_L
		0x9E, // INTEL_FAM6_KABYLAKE
		0xA5, // INTEL_FAM6_COMETLAKE
		0xA6, // INTEL_FAM6_COMETLAKE_L
		0x66, // INTEL_FAM6_CANNONLAKE_L
		0x6A, // INTEL_FAM6_ICELAKE_X
		0x6C, // INTEL_FAM6_ICELAKE_D
		0x7D, // INTEL_FAM6_ICELAKE
		0x7E, // INTEL_FAM6_ICELAKE_L
		0x9D, // INTEL_FAM6_ICELAKE_NNPI
		0xA7, // INTEL_FAM6_ROCKETLAKE
		0x8C, // INTEL_FAM6_TIGERLAKE_L
		0x8D, // INTEL_FAM6_TIGERLAKE
		0x8F, // INTEL_FAM6_SAPPHIRERAPIDS_X
		0xCF, // INTEL_FAM6_EMERALDRAPIDS_X
		0xAD, // INTEL_FAM6_GRANITERAPIDS_X
		0x8A, // INTEL_FAM6_LAKEFIELD
		0x97, // INTEL_FAM6_ALDERLAKE
		0x9A, // INTEL_FAM6_ALDERLAKE_L
		0xB7, // INTEL_FAM6_RAPTORLAKE
		0xBA, // INTEL_FAM6_RAPTORLAKE_P
		0xBF, // INTEL_FAM6_RAPTORLAKE_S
		0xAC, // INTEL_FAM6_METEORLAKE
		0xAA, // INTEL_FAM6_METEORLAKE_L
		0xC6, // INTEL_FAM6_ARROWLAKE
		0xBD, // INTEL_FAM6_LUNARLAKE_M
		0x37, // INTEL_FAM6_ATOM_SILVERMONT
		0x4D, // INTEL_FAM6_ATOM_SILVERMONT_D
		0x4A, // INTEL_FAM6_ATOM_SILVERMONT_MID
		0x5A, // INTEL_FAM6_ATOM_SILVERMONT_SMARTPHONE
		0x4C, // INTEL_FAM6_ATOM_AIRMONT
		0x5C, // INTEL_FAM6_ATOM_GOLDMONT
		0x5F, // INTEL_FAM6_ATOM_GOLDMONT_D
		0x7A, // INTEL_FAM6_ATOM_GOLDMONT_PLUS
		0x86, // INTEL_FAM6_ATOM_TREMONT_D
		0x96, // INTEL_FAM6_ATOM_TREMONT
		0x9C, // INTEL_FAM6_ATOM_TREMONT_L
		0xBE, // INTEL_FAM6_ATOM_GRACEMONT
		0xAF, // INTEL_FAM6_ATOM_CRESTMONT_X
		0xB6, // INTEL_FAM6_ATOM_CRESTMONT
		0x57, // INTEL_FAM6_XEON_PHI_KNL
		0x85, // INTEL_FAM6_XEON_PHI_KNM
	}

	c3Models = []int{
		0x1E, // INTEL_FAM6_NEHALEM
		0x1F, // INTEL_FAM6_NEHALEM_G
		0x1A, // INTEL_FAM6_NEHALEM_EP
		0x2E, // INTEL_FAM6_NEHALEM_EX
		0x25, // INTEL_FAM6_WESTMERE
		0x2C, // INTEL_FAM6_WESTMERE_EP
		0x2F, // INTEL_FAM6_WESTMERE_EX
		0x2A, // INTEL_FAM6_SANDYBRIDGE
		0x2D, // INTEL_FAM6_SANDYBRIDGE_X
		0x3A, // INTEL_FAM6_IVYBRIDGE
		0x3E, // INTEL_FAM6_IVYBRIDGE_X
		0x3C, // INTEL_FAM6_HASWELL
		0x3F, // INTEL_FAM6_HASWELL_X
		0x45, // INTEL_FAM6_HASWELL_L
		0x46, // INTEL_FAM6_HASWELL_G
		0x3D, // INTEL_FAM6_BROADWELL
		0x47, // INTEL_FAM6_BROADWELL_G
		0x4F, // INTEL_FAM6_BROADWELL_X
		0x56, // INTEL_FAM6_BROADWELL_D
		0x4E, // INTEL_FAM6_SKYLAKE_L
		0x5E, // INTEL_FAM6_SKYLAKE
		0x8E, // INTEL_FAM6_KABYLAKE_L
		0x9E, // INTEL_FAM6_KABYLAKE
		0xA5, // INTEL_FAM6_COMETLAKE
		0xA6, // INTEL_FAM6_COMETLAKE_L
		0x4C, // INTEL_FAM6_ATOM_AIRMONT
		0x5C, // INTEL_FAM6_ATOM_GOLDMONT
		0x7A, // INTEL_FAM6_ATOM_GOLDMONT_PLUS
	}

	c7Models = []int{
		0x2A, // INTEL_FAM6_SANDYBRIDGE
		0x2D, // INTEL_FAM6_SANDYBRIDGE_X
		0x3A, // INTEL_FAM6_IVYBRIDGE
		0x3E, // INTEL_FAM6_IVYBRIDGE_X
		0x3C, // INTEL_FAM6_HASWELL
		0x3F, // INTEL_FAM6_HASWELL_X
		0x45, // INTEL_FAM6_HASWELL_L
		0x46, // INTEL_FAM6_HASWELL_G
		0x3D, // INTEL_FAM6_BROADWELL
		0x47, // INTEL_FAM6_BROADWELL_G
		0x4E, // INTEL_FAM6_SKYLAKE_L
		0x5E, // INTEL_FAM6_SKYLAKE
		0x8E, // INTEL_FAM6_KABYLAKE_L
		0x9E, // INTEL_FAM6_KABYLAKE
		0xA5, // INTEL_FAM6_COMETLAKE
		0xA6, // INTEL_FAM6_COMETLAKE_L
		0x66, // INTEL_FAM6_CANNONLAKE_L
		0x7E, // INTEL_FAM6_ICELAKE_L
		0x9D, // INTEL_FAM6_ICELAKE_NNPI
		0xA7, // INTEL_FAM6_ROCKETLAKE
		0x8C, // INTEL_FAM6_TIGERLAKE_L
		0x8D, // INTEL_FAM6_TIGERLAKE
		0x8A, // INTEL_FAM6_LAKEFIELD
		0x97, // INTEL_FAM6_ALDERLAKE
		0x9A, // INTEL_FAM6_ALDERLAKE_L
		0xB7, // INTEL_FAM6_RAPTORLAKE
		0xBA, // INTEL_FAM6_RAPTORLAKE_P
		0xBF, // INTEL_FAM6_RAPTORLAKE_S
		0xAC, // INTEL_FAM6_METEORLAKE
		0xAA, // INTEL_FAM6_METEORLAKE_L
		0xC6, // INTEL_FAM6_ARROWLAKE
		0xBD, // INTEL_FAM6_LUNARLAKE_M
		0x5C, // INTEL_FAM6_ATOM_GOLDMONT
		0x7A, // INTEL_FAM6_ATOM_GOLDMONT_PLUS
		0x96, // INTEL_FAM6_ATOM_TREMONT
		0x9C, // INTEL_FAM6_ATOM_TREMONT_L
		0xBE, // INTEL_FAM6_ATOM_GRACEMONT
	}
)
