// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetCPUFrequencyMhz(t *testing.T) {
	testCases := []struct {
		name                 string
		baseCPUFrequencyPath string
		cpuID                int
		expected             float64
		err                  error
	}{
		{
			name:                 "Correct value of cpu's frequency has been returned.",
			baseCPUFrequencyPath: "testdata/cpu-freq",
			cpuID:                0,
			expected:             888.888,
			err:                  nil,
		},
		{
			name:                 "NonNumericContent",
			baseCPUFrequencyPath: "testdata/cpu-freq-invalid",
			cpuID:                0,
			expected:             0,
			err:                  errors.New("error while converting value from file \"testdata/cpu-freq-invalid/cpu0/cpufreq/scaling_cur_freq\""),
		},
		{
			name:                 "InvalidPath",
			baseCPUFrequencyPath: "testdata/cpu-freq-invalid-path",
			cpuID:                0,
			expected:             0,
			err: errors.New("error reading file \"testdata/cpu-freq-invalid-path/cpu0" +
				"/cpufreq/scaling_cur_freq\": file \"testdata/cpu-freq-invalid-path/cpu0/cpufreq/scaling_cur_freq\" does not exist"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := cpuFreqData{
				cpuFrequencyFilePath: tc.baseCPUFrequencyPath,
			}
			cpuFrequencyValue, err := c.getCPUFrequencyMhz(tc.cpuID)
			if tc.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, cpuFrequencyValue)
			}
		})
	}
}

func TestCPUFreqData_Init(t *testing.T) {
	testCases := []struct {
		name        string
		cpuFreqPath string
		err         error
	}{
		{
			name:        "Initialized",
			cpuFreqPath: "testdata/cpu-freq",
			err:         nil,
		},
		{
			name:        "EmptyString",
			cpuFreqPath: "",
			err:         errors.New("base path of CPU core frequency cannot be empty"),
		},
		{
			name:        "WrongPath",
			cpuFreqPath: "/dummy/path",
			err:         errors.New("invalid base path of CPU core frequency"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpuFreq := &cpuFreqData{
				cpuFrequencyFilePath: tc.cpuFreqPath,
			}

			err := cpuFreq.init()
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
