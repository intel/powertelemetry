// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMaxTurboFreq_hasHswTurboRatioLimit(t *testing.T) {
	selectedModels := []int{
		0x3F, // INTEL_FAM6_HASWELL_X
	}

	m := make(map[int]interface{})
	for _, v := range selectedModels {
		m[v] = struct{}{}
	}

	for model := 0; model < 0xFF; model++ {
		ret := hasHswTurboRatioLimit(model)
		require.Equalf(t, m[model] != nil, ret, "Model 0x%X", model)
	}
}

func TestMaxTurboFreq_hasKnlTurboRatioLimit(t *testing.T) {
	selectedModels := []int{
		0x57, // INTEL_FAM6_XEON_PHI_KNL
		0x85, // INTEL_FAM6_XEON_PHI_KNM
	}

	m := make(map[int]interface{})
	for _, v := range selectedModels {
		m[v] = struct{}{}
	}

	for model := 0; model < 0xFF; model++ {
		ret := hasKnlTurboRatioLimit(model)
		require.Equalf(t, m[model] != nil, ret, "Model 0x%X", model)
	}
}

func TestMaxTurboFreq_hasIvtTurboRatioLimit(t *testing.T) {
	selectedModels := []int{
		0x3E, // INTEL_FAM6_IVYBRIDGE_X
		0x3F, // INTEL_FAM6_HASWELL_X
	}

	m := make(map[int]interface{})
	for _, v := range selectedModels {
		m[v] = struct{}{}
	}

	for model := 0; model < 0xFF; model++ {
		ret := hasIvtTurboRatioLimit(model)
		require.Equalf(t, m[model] != nil, ret, "Model 0x%X", model)
	}
}

func TestMaxTurboFreq_hasSlvMsrs(t *testing.T) {
	selectedModels := []int{
		0x37, // INTEL_FAM6_ATOM_SILVERMONT
		0x4A, // INTEL_FAM6_ATOM_SILVERMONT_MID
		0x5A, // INTEL_FAM6_ATOM_AIRMONT_MID
	}

	m := make(map[int]interface{})
	for _, v := range selectedModels {
		m[v] = struct{}{}
	}

	for model := 0; model < 0xFF; model++ {
		ret := hasSlvMsrs(model)
		require.Equalf(t, m[model] != nil, ret, "Model 0x%X", model)
	}
}

func TestMaxTurboFreq_hasAtomTurboRatioLimit(t *testing.T) {
	selectedModels := []int{
		0x37, // INTEL_FAM6_ATOM_SILVERMONT
		0x4A, // INTEL_FAM6_ATOM_SILVERMONT_MID
		0x5A, // INTEL_FAM6_ATOM_AIRMONT_MID
	}

	m := make(map[int]interface{})
	for _, v := range selectedModels {
		m[v] = struct{}{}
	}

	for model := 0; model < 0xFF; model++ {
		ret := hasAtomTurboRatioLimit(model)
		require.Equalf(t, m[model] != nil, ret, "Model 0x%X", model)
	}
}

func TestMaxTurboFreq_hasTurboRatioLimit(t *testing.T) {
	selectedModels := []int{
		0x37, // INTEL_FAM6_ATOM_SILVERMONT
		0x4A, // INTEL_FAM6_ATOM_SILVERMONT_MID
		0x5A, // INTEL_FAM6_ATOM_AIRMONT_MID
		0x2E, // INTEL_FAM6_NEHALEM_EX
		0x2F, // INTEL_FAM6_WESTMERE_EX
		0x57, // INTEL_FAM6_XEON_PHI_KNL
		0x85, // INTEL_FAM6_XEON_PHI_KNM
	}

	m := make(map[int]interface{})
	for _, v := range selectedModels {
		m[v] = struct{}{}
	}

	for model := 0; model < 0xFF; model++ {
		ret := hasTurboRatioLimit(model)

		// Note: inverse test
		require.Equalf(t, m[model] == nil, ret, "Model 0x%X", model)
	}
}

func TestMaxTurboFreq_hasTurboRatioGroupLimits(t *testing.T) {
	selectedModels := []int{
		0x5C, // INTEL_FAM6_ATOM_GOLDMONT
		0x55, // INTEL_FAM6_SKYLAKE_X
		0x6A, // INTEL_FAM6_ICELAKE_X
		0x6C, // INTEL_FAM6_ICELAKE_D
		0x8F, // INTEL_FAM6_SAPPHIRERAPIDS_X
		0x5F, // INTEL_FAM6_ATOM_GOLDMONT_D
		0x86, // INTEL_FAM6_ATOM_TREMONT_D
		0xCF, // INTEL_FAM6_EMERALDRAPIDS_X
	}

	m := make(map[int]interface{})
	for _, v := range selectedModels {
		m[v] = struct{}{}
	}

	for model := 0; model < 0xFF; model++ {
		ret := hasTurboRatioGroupLimits(model)
		require.Equalf(t, m[model] != nil, ret, "Model 0x%X", model)
	}
}

func TestMaxTurboFreq_dumpHswTurboRatioLimits(t *testing.T) {
	cases := []struct {
		description string
		cpuID       int
		busClock    float64
		msrValue    uint64
		expected    []MaxTurboFreq
	}{
		{
			description: "Normal case A", busClock: 133.3, msrValue: 0x0000000000001716,
			expected: []MaxTurboFreq{
				{ActiveCores: 18, Value: 3065},
				{ActiveCores: 17, Value: 2932},
			},
		},
		{
			description: "Normal case B", busClock: 100.0, msrValue: 0x0000000000001110,
			expected: []MaxTurboFreq{
				{ActiveCores: 18, Value: 1700},
				{ActiveCores: 17, Value: 1600},
			},
		},
		{
			description: "One ratio limit is zero", busClock: 100.0, msrValue: 0x0000000000001100,
			expected: []MaxTurboFreq{
				{ActiveCores: 18, Value: 1700},
			},
		},
		{
			description: "More ratio limits returned", busClock: 100.0, msrValue: 0x1716151413121110,
			expected: []MaxTurboFreq{
				{ActiveCores: 18, Value: 1700},
				{ActiveCores: 17, Value: 1600},
			},
		},
		{
			description: "Ratio limits zeroed", busClock: 100.0, msrValue: 0x0000000000000000,
			expected: []MaxTurboFreq{},
		},
	}

	for _, v := range cases {
		t.Run(v.description, func(t *testing.T) {
			m := &msrMock{}
			pt := &PowerTelemetry{msr: m, busClock: v.busClock}

			// Mock reading from MSR_TURBO_RATIO_LIMIT2
			m.On("read", uint32(0x1AF), 0).Return(v.msrValue, nil).Once()

			list, err := pt.dumpHswTurboRatioLimits(v.cpuID)
			require.NoError(t, err)
			require.Equal(t, v.expected, list)
		})
	}
}

func TestMaxTurboFreq_dumpIvtTurboRatioLimits(t *testing.T) {
	cases := []struct {
		description string
		cpuID       int
		busClock    float64
		msrValue    uint64
		expected    []MaxTurboFreq
	}{
		{
			description: "Normal case A", busClock: 133.3, msrValue: 0x1817161514131211,
			expected: []MaxTurboFreq{
				{ActiveCores: 16, Value: 3199},
				{ActiveCores: 15, Value: 3065},
				{ActiveCores: 14, Value: 2932},
				{ActiveCores: 13, Value: 2799},
				{ActiveCores: 12, Value: 2666},
				{ActiveCores: 11, Value: 2532},
				{ActiveCores: 10, Value: 2399},
				{ActiveCores: 9, Value: 2266},
			},
		},
		{
			description: "Normal case B", busClock: 100.0, msrValue: 0x1716151413121110,
			expected: []MaxTurboFreq{
				{ActiveCores: 16, Value: 2300},
				{ActiveCores: 15, Value: 2200},
				{ActiveCores: 14, Value: 2100},
				{ActiveCores: 13, Value: 2000},
				{ActiveCores: 12, Value: 1900},
				{ActiveCores: 11, Value: 1800},
				{ActiveCores: 10, Value: 1700},
				{ActiveCores: 9, Value: 1600},
			},
		},
		{
			description: "One ratio limit is zero", busClock: 100.0, msrValue: 0x1716150013121110,
			expected: []MaxTurboFreq{
				{ActiveCores: 16, Value: 2300},
				{ActiveCores: 15, Value: 2200},
				{ActiveCores: 14, Value: 2100},
				{ActiveCores: 12, Value: 1900},
				{ActiveCores: 11, Value: 1800},
				{ActiveCores: 10, Value: 1700},
				{ActiveCores: 9, Value: 1600},
			},
		},
		{
			description: "Ratio limits zeroed", busClock: 100.0, msrValue: 0x0000000000000000,
			expected: []MaxTurboFreq{},
		},
	}

	for _, v := range cases {
		t.Run(v.description, func(t *testing.T) {
			m := &msrMock{}
			pt := &PowerTelemetry{msr: m, busClock: v.busClock}

			// Mock reading from MSR_TURBO_RATIO_LIMIT1
			m.On("read", uint32(0x1AE), 0).Return(v.msrValue, nil).Once()

			list, err := pt.dumpIvtTurboRatioLimits(v.cpuID)
			require.NoError(t, err)
			require.Equal(t, v.expected, list)
		})
	}
}

func TestMaxTurboFreq_dumpTurboRatioLimits(t *testing.T) {
	cases := []struct {
		description string
		cpuID       int
		busClock    float64
		model       int
		msrValue1   uint64 // Mock reading from MSR_TURBO_RATIO_LIMIT
		msrValue2   uint64 // Mock reading from MSR_TURBO_RATIO_LIMIT1
		expected    []MaxTurboFreq
	}{
		{
			description: "Normal case, the model supports group limits", // INTEL_FAM6_ICELAKE_D
			busClock:    133.3, msrValue1: 0x1817161514131211, msrValue2: 0x100E0C0A08060402, model: 0x6C,
			expected: []MaxTurboFreq{
				{ActiveCores: 16, Value: 3199},
				{ActiveCores: 14, Value: 3065},
				{ActiveCores: 12, Value: 2932},
				{ActiveCores: 10, Value: 2799},
				{ActiveCores: 8, Value: 2666},
				{ActiveCores: 6, Value: 2532},
				{ActiveCores: 4, Value: 2399},
				{ActiveCores: 2, Value: 2266},
			},
		},
		{
			description: "Normal case, the model doesn't support group limits",
			busClock:    100.0, msrValue1: 0x1716151413121110, msrValue2: 0x0000000000000000, model: 0x00,
			expected: []MaxTurboFreq{
				{ActiveCores: 8, Value: 2300},
				{ActiveCores: 7, Value: 2200},
				{ActiveCores: 6, Value: 2100},
				{ActiveCores: 5, Value: 2000},
				{ActiveCores: 4, Value: 1900},
				{ActiveCores: 3, Value: 1800},
				{ActiveCores: 2, Value: 1700},
				{ActiveCores: 1, Value: 1600},
			},
		},
		{
			description: "One ratio limit is zero",
			busClock:    100.0, msrValue1: 0x1700151413121110, msrValue2: 0x0000000000000000, model: 0x00,
			expected: []MaxTurboFreq{
				{ActiveCores: 8, Value: 2300},
				{ActiveCores: 6, Value: 2100},
				{ActiveCores: 5, Value: 2000},
				{ActiveCores: 4, Value: 1900},
				{ActiveCores: 3, Value: 1800},
				{ActiveCores: 2, Value: 1700},
				{ActiveCores: 1, Value: 1600},
			},
		},
		{
			description: "Ratio limits zeroed",
			busClock:    100.0, msrValue1: 0x0000000000000000, msrValue2: 0x0000000000000000, model: 0x00,
			expected: []MaxTurboFreq{},
		},
	}

	for _, v := range cases {
		t.Run(v.description, func(t *testing.T) {
			m := &msrMock{}
			pt := &PowerTelemetry{msr: m, busClock: v.busClock}

			// Mock reading from MSR_TURBO_RATIO_LIMIT
			m.On("read", uint32(0x1AD), 0).Return(v.msrValue1, nil).Once()

			// Mock reading from MSR_TURBO_RATIO_LIMIT1
			m.On("read", uint32(0x1AE), 0).Return(v.msrValue2, nil).Once()

			list, err := pt.dumpTurboRatioLimits(turboRatioLimit, v.model, v.cpuID)
			require.NoError(t, err)
			require.Equal(t, v.expected, list)
		})
	}
}

func TestMaxTurboFreq_dumpAtomTurboRatioLimits(t *testing.T) {
	cases := []struct {
		description string
		cpuID       int
		busClock    float64
		msrValue    uint64
		expected    []MaxTurboFreq
	}{
		{
			description: "Normal case A", busClock: 133.3, msrValue: 0x0000000014131211,
			expected: []MaxTurboFreq{
				{ActiveCores: 4, Value: 2666},
				{ActiveCores: 3, Value: 2532},
				{ActiveCores: 2, Value: 2399},
				{ActiveCores: 1, Value: 2266},
			},
		},
		{
			description: "Normal case B", busClock: 100.0, msrValue: 0x0000000013121110,
			expected: []MaxTurboFreq{
				{ActiveCores: 4, Value: 1900},
				{ActiveCores: 3, Value: 1800},
				{ActiveCores: 2, Value: 1700},
				{ActiveCores: 1, Value: 1600},
			},
		},
		{
			description: "One ratio limit is zero", busClock: 100.0, msrValue: 0x0000000013120010,
			expected: []MaxTurboFreq{
				{ActiveCores: 4, Value: 1900},
				{ActiveCores: 3, Value: 1800},
				{ActiveCores: 1, Value: 1600},
			},
		},
		{
			description: "More ratio limits returned", busClock: 100.0, msrValue: 0x1716151413121110,
			expected: []MaxTurboFreq{
				{ActiveCores: 4, Value: 1900},
				{ActiveCores: 3, Value: 1800},
				{ActiveCores: 2, Value: 1700},
				{ActiveCores: 1, Value: 1600},
			},
		},
		{
			description: "Ratio limits zeroed", busClock: 100.0, msrValue: 0x0000000000000000,
			expected: []MaxTurboFreq{},
		},
	}

	for _, v := range cases {
		t.Run(v.description, func(t *testing.T) {
			m := &msrMock{}
			pt := &PowerTelemetry{msr: m, busClock: v.busClock}

			// Mock reading from MSR_ATOM_CORE_TURBO_RATIOS
			m.On("read", uint32(0x66C), 0).Return(v.msrValue, nil).Once()

			list, err := pt.dumpAtomTurboRatioLimits(v.cpuID)
			require.NoError(t, err)
			require.Equal(t, v.expected, list)
		})
	}
}

func TestMaxTurboFreq_dumpKnlTurboRatioLimits(t *testing.T) {
	cases := []struct {
		description    string
		cpuID          int
		busClock       float64
		msrBinaryValue string // String binary representation for easier understanding of bit field values
		expected       []MaxTurboFreq
	}{
		{
			description: "Normal case A", busClock: 133.3,
			msrBinaryValue: "" + // Note: this it to fix indentation to improve readability
				"001" + "00010" + // bits 63:56
				"001" + "00010" + // bits 55:48
				"001" + "00010" + // bits 47:40
				"001" + "00010" + // bits 39:32
				"001" + "00010" + // bits 31:24
				"001" + "00010" + // bits 23:16
				"00010000" + // bits 15:8
				"0000010" + "0", // bits 7:0
			expected: []MaxTurboFreq{
				{ActiveCores: 14, Value: 1333},
				{ActiveCores: 12, Value: 1466},
				{ActiveCores: 10, Value: 1599},
				{ActiveCores: 8, Value: 1732},
				{ActiveCores: 6, Value: 1866},
				{ActiveCores: 4, Value: 1999},
				{ActiveCores: 2, Value: 2132},
			},
		},
		{
			description: "Normal case B", busClock: 100.0,
			msrBinaryValue: "" +
				"001" + "00001" + // bits 63:56
				"001" + "00001" + // bits 55:48
				"001" + "00001" + // bits 47:40
				"001" + "00001" + // bits 39:32
				"001" + "00001" + // bits 31:24
				"001" + "00001" + // bits 23:16
				"00010100" + // bits 15:8
				"0000100" + "0", // bits 7:0
			expected: []MaxTurboFreq{
				{ActiveCores: 10, Value: 1400},
				{ActiveCores: 9, Value: 1500},
				{ActiveCores: 8, Value: 1600},
				{ActiveCores: 7, Value: 1700},
				{ActiveCores: 6, Value: 1800},
				{ActiveCores: 5, Value: 1900},
				{ActiveCores: 4, Value: 2000},
			},
		},
		{
			description: "One group ratio delta is zero", busClock: 100.0,
			msrBinaryValue: "" +
				"001" + "00001" + // bits 63:56
				"001" + "00001" + // bits 55:48
				"000" + "00001" + // bits 47:40
				"001" + "00001" + // bits 39:32
				"001" + "00001" + // bits 31:24
				"001" + "00001" + // bits 23:16
				"00010100" + // bits 15:8
				"0000100" + "0", // bits 7:0
			expected: []MaxTurboFreq{
				{ActiveCores: 10, Value: 1500},
				{ActiveCores: 9, Value: 1600},
				{ActiveCores: 7, Value: 1700},
				{ActiveCores: 6, Value: 1800},
				{ActiveCores: 5, Value: 1900},
				{ActiveCores: 4, Value: 2000},
			},
		},
		{
			description: "One cores delta is zero", busClock: 100.0,
			msrBinaryValue: "" +
				"001" + "00001" + // bits 63:56
				"001" + "00001" + // bits 55:48
				"001" + "00001" + // bits 47:40
				"001" + "00000" + // bits 39:32
				"001" + "00001" + // bits 31:24
				"001" + "00001" + // bits 23:16
				"00010100" + // bits 15:8
				"0000100" + "0", // bits 7:0
			expected: []MaxTurboFreq{
				{ActiveCores: 9, Value: 1400},
				{ActiveCores: 8, Value: 1500},
				{ActiveCores: 7, Value: 1600},
				{ActiveCores: 6, Value: 1700},
				{ActiveCores: 6, Value: 1800},
				{ActiveCores: 5, Value: 1900},
				{ActiveCores: 4, Value: 2000},
			},
		},
		{
			description: "Ratio limits zeroed", busClock: 100.0,
			msrBinaryValue: "" +
				"000" + "00000" + // bits 63:56
				"000" + "00000" + // bits 55:48
				"000" + "00000" + // bits 47:40
				"000" + "00000" + // bits 39:32
				"000" + "00000" + // bits 31:24
				"000" + "00000" + // bits 23:16
				"00000000" + // bits 15:8
				"0000000" + "0", // bits 7:0
			expected: []MaxTurboFreq{
				{ActiveCores: 0, Value: 0},
			},
		},
	}

	for _, v := range cases {
		t.Run(v.description, func(t *testing.T) {
			m := &msrMock{}
			pt := &PowerTelemetry{msr: m, busClock: v.busClock}

			msr, err := strconv.ParseUint(v.msrBinaryValue, 2, 64)
			require.NoError(t, err)

			// Mock reading from MSR_TURBO_RATIO_LIMIT
			m.On("read", uint32(0x1AD), 0).Return(msr, nil).Once()

			list, err := pt.dumpKnlTurboRatioLimits(v.cpuID)
			require.NoError(t, err)
			require.Equal(t, v.expected, list)
		})
	}
}

func TestGetMaxTurboFreqList(t *testing.T) {
	cases := []struct {
		description    string
		model          int
		hybrid         bool
		busClock       float64
		msrValue1      uint64 // MSR_TURBO_RATIO_LIMIT
		msrValue2      uint64 // MSR_TURBO_RATIO_LIMIT1
		msrValue3      uint64 // MSR_TURBO_RATIO_LIMIT2
		msrValue4      uint64 // MSR_ATOM_CORE_TURBO_RATIOS
		msrValue5      uint64 // MSR_SECONDARY_TURBO_RATIO_LIMIT
		msrBinaryValue string // MSR_TURBO_RATIO_LIMIT (string binary representation)
		expected       []MaxTurboFreq
	}{
		{
			description: "Haswell X", busClock: 100.0, model: 0x3F, // INTEL_FAM6_HASWELL_X
			hybrid:         false,
			msrValue1:      0x1716151413121110,
			msrValue2:      0x1F1E1D1C1B1A1918,
			msrValue3:      0x0000000000002120,
			msrValue4:      0x0000000000000000,
			msrValue5:      0x0000000000000000,
			msrBinaryValue: "",
			expected: []MaxTurboFreq{
				{ActiveCores: 18, Value: 3300},
				{ActiveCores: 17, Value: 3200},
				{ActiveCores: 16, Value: 3100},
				{ActiveCores: 15, Value: 3000},
				{ActiveCores: 14, Value: 2900},
				{ActiveCores: 13, Value: 2800},
				{ActiveCores: 12, Value: 2700},
				{ActiveCores: 11, Value: 2600},
				{ActiveCores: 10, Value: 2500},
				{ActiveCores: 9, Value: 2400},
				{ActiveCores: 8, Value: 2300},
				{ActiveCores: 7, Value: 2200},
				{ActiveCores: 6, Value: 2100},
				{ActiveCores: 5, Value: 2000},
				{ActiveCores: 4, Value: 1900},
				{ActiveCores: 3, Value: 1800},
				{ActiveCores: 2, Value: 1700},
				{ActiveCores: 1, Value: 1600},
			},
		},
		{
			description: "Ivy Bridge X", busClock: 100.0, model: 0x3E, // INTEL_FAM6_IVYBRIDGE_X
			hybrid:         false,
			msrValue1:      0x1716151413121110,
			msrValue2:      0x1F1E1D1C1B1A1918,
			msrValue3:      0x0000000000000000,
			msrValue4:      0x0000000000000000,
			msrValue5:      0x0000000000000000,
			msrBinaryValue: "",
			expected: []MaxTurboFreq{
				{ActiveCores: 16, Value: 3100},
				{ActiveCores: 15, Value: 3000},
				{ActiveCores: 14, Value: 2900},
				{ActiveCores: 13, Value: 2800},
				{ActiveCores: 12, Value: 2700},
				{ActiveCores: 11, Value: 2600},
				{ActiveCores: 10, Value: 2500},
				{ActiveCores: 9, Value: 2400},
				{ActiveCores: 8, Value: 2300},
				{ActiveCores: 7, Value: 2200},
				{ActiveCores: 6, Value: 2100},
				{ActiveCores: 5, Value: 2000},
				{ActiveCores: 4, Value: 1900},
				{ActiveCores: 3, Value: 1800},
				{ActiveCores: 2, Value: 1700},
				{ActiveCores: 1, Value: 1600},
			},
		},
		{
			description: "Yonah", busClock: 100.0, model: 0x0E, // INTEL_FAM6_CORE_YONAH
			hybrid:         false,
			msrValue1:      0x1716151413121110,
			msrValue2:      0x0000000000000000,
			msrValue3:      0x0000000000000000,
			msrValue4:      0x0000000000000000,
			msrValue5:      0x0000000000000000,
			msrBinaryValue: "",
			expected: []MaxTurboFreq{
				{ActiveCores: 8, Value: 2300},
				{ActiveCores: 7, Value: 2200},
				{ActiveCores: 6, Value: 2100},
				{ActiveCores: 5, Value: 2000},
				{ActiveCores: 4, Value: 1900},
				{ActiveCores: 3, Value: 1800},
				{ActiveCores: 2, Value: 1700},
				{ActiveCores: 1, Value: 1600},
			},
		},
		{
			description: "Alderlake", busClock: 100.0, model: 0x97, // INTEL_FAM6_ALDERLAKE
			hybrid:         true,
			msrValue1:      0x1716151413121110,
			msrValue2:      0x0000000000000000,
			msrValue3:      0x0000000000000000,
			msrValue4:      0x0000000000000000,
			msrValue5:      0x1716151413121110,
			msrBinaryValue: "",
			expected: []MaxTurboFreq{
				{ActiveCores: 8, Value: 2300, Secondary: false},
				{ActiveCores: 7, Value: 2200, Secondary: false},
				{ActiveCores: 6, Value: 2100, Secondary: false},
				{ActiveCores: 5, Value: 2000, Secondary: false},
				{ActiveCores: 4, Value: 1900, Secondary: false},
				{ActiveCores: 3, Value: 1800, Secondary: false},
				{ActiveCores: 2, Value: 1700, Secondary: false},
				{ActiveCores: 1, Value: 1600, Secondary: false},
				{ActiveCores: 8, Value: 2300, Secondary: true},
				{ActiveCores: 7, Value: 2200, Secondary: true},
				{ActiveCores: 6, Value: 2100, Secondary: true},
				{ActiveCores: 5, Value: 2000, Secondary: true},
				{ActiveCores: 4, Value: 1900, Secondary: true},
				{ActiveCores: 3, Value: 1800, Secondary: true},
				{ActiveCores: 2, Value: 1700, Secondary: true},
				{ActiveCores: 1, Value: 1600, Secondary: true},
			},
		},
		{
			description: "Atom Silvermont", busClock: 100.0, model: 0x37, // INTEL_FAM6_ATOM_SILVERMONT
			hybrid:         false,
			msrValue1:      0x0000000000000000,
			msrValue2:      0x0000000000000000,
			msrValue3:      0x0000000000000000,
			msrValue4:      0x0000000013121110,
			msrValue5:      0x0000000000000000,
			msrBinaryValue: "",
			expected: []MaxTurboFreq{
				{ActiveCores: 4, Value: 1900},
				{ActiveCores: 3, Value: 1800},
				{ActiveCores: 2, Value: 1700},
				{ActiveCores: 1, Value: 1600},
			},
		},
		{
			description: "Nehalem Ex", busClock: 100.0, model: 0x2E, // INTEL_FAM6_NEHALEM_EX
			hybrid:         false,
			msrValue1:      0x1716151413121110,
			msrValue2:      0x1817161514131211,
			msrValue3:      0x1918171615141312,
			msrValue4:      0x1A19181716151413,
			msrValue5:      0x0000000000000000,
			msrBinaryValue: "",
			expected:       []MaxTurboFreq{},
		},
		{
			description: "Knights Landing", busClock: 100.0, model: 0x57, // INTEL_FAM6_XEON_PHI_KNL
			hybrid:    false,
			msrValue1: 0x0000000000000000,
			msrValue2: 0x0000000000000000,
			msrValue3: 0x0000000000000000,
			msrValue4: 0x0000000000000000,
			msrValue5: 0x0000000000000000,
			msrBinaryValue: "" +
				"001" + "00010" + // bits 63:56
				"001" + "00010" + // bits 55:48
				"001" + "00010" + // bits 47:40
				"001" + "00010" + // bits 39:32
				"001" + "00010" + // bits 31:24
				"001" + "00010" + // bits 23:16
				"00010000" + // bits 15:8
				"0000010" + "0", // bits 7:0
			expected: []MaxTurboFreq{
				{ActiveCores: 14, Value: 1000},
				{ActiveCores: 12, Value: 1100},
				{ActiveCores: 10, Value: 1200},
				{ActiveCores: 8, Value: 1300},
				{ActiveCores: 6, Value: 1400},
				{ActiveCores: 4, Value: 1500},
				{ActiveCores: 2, Value: 1600},
			},
		},
		{
			description: "Unknown model, zeroed MSRs", busClock: 100.0, model: 0x00,
			hybrid:         false,
			msrValue1:      0x0000000000000000,
			msrValue2:      0x0000000000000000,
			msrValue3:      0x0000000000000000,
			msrValue4:      0x0000000000000000,
			msrValue5:      0x0000000000000000,
			msrBinaryValue: "",
			expected:       []MaxTurboFreq{},
		},
	}

	for _, v := range cases {
		t.Run(v.description, func(t *testing.T) {
			m := &msrMock{}
			pt := &PowerTelemetry{
				msr:      m,
				busClock: v.busClock,
				topology: &topologyData{
					topologyMap: map[int]*cpuInfo{
						0: {
							packageID: 0,
						},
					},
					model: v.model,
				},
				cpus: []int{0},
			}

			isHybrid = func() bool {
				return v.hybrid
			}

			// If msrBinaryValue is not defined, use msrValue1 to pass to MSR_TURBO_RATIO_LIMIT
			var msr uint64
			msr, err := strconv.ParseUint(v.msrBinaryValue, 2, 64)
			if err != nil {
				msr = v.msrValue1
			}

			// Mock reading from MSR_TURBO_RATIO_LIMIT
			m.On("read", uint32(0x1AD), 0).Return(msr, nil).Once()

			// Mock reading from MSR_TURBO_RATIO_LIMIT1
			m.On("read", uint32(0x1AE), 0).Return(v.msrValue2, nil).Once()

			// Mock reading from MSR_TURBO_RATIO_LIMIT2
			m.On("read", uint32(0x1AF), 0).Return(v.msrValue3, nil).Once()

			// Mock reading from MSR_ATOM_CORE_TURBO_RATIOS
			m.On("read", uint32(0x66C), 0).Return(v.msrValue4, nil).Once()

			// Mock reading from MSR_SECONDARY_TURBO_RATIO_LIMIT
			m.On("read", uint32(0x650), 0).Return(v.msrValue5, nil).Once()

			list, err := pt.GetMaxTurboFreqList(0)
			require.NoError(t, err)
			require.Equal(t, v.expected, list)
		})
	}
}
