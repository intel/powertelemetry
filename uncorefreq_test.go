// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUncoreFreqTypeToString(t *testing.T) {
	t.Run("InitialMax", func(t *testing.T) {
		freqType := uncoreFreqType(0)
		require.Equal(t, "initial_max", freqType.String())
	})
	t.Run("InitialMin", func(t *testing.T) {
		freqType := uncoreFreqType(1)
		require.Equal(t, "initial_min", freqType.String())
	})
	t.Run("CustomizedMax", func(t *testing.T) {
		freqType := uncoreFreqType(2)
		require.Equal(t, "max", freqType.String())
	})
	t.Run("CustomizedMin", func(t *testing.T) {
		freqType := uncoreFreqType(3)
		require.Equal(t, "min", freqType.String())
	})
	t.Run("Current", func(t *testing.T) {
		freqType := uncoreFreqType(4)
		require.Equal(t, "current", freqType.String())
	})
	t.Run("Invalid", func(t *testing.T) {
		freqType := uncoreFreqType(5)
		require.Equal(t, "", freqType.String())
	})
}

func TestGetUncoreFrequencyPath(t *testing.T) {
	testCases := []struct {
		name      string
		packageID int
		dieID     int
		freqType  string
		expected  string
		err       error
	}{
		{
			name:      "InvalidFreqType",
			packageID: 1,
			dieID:     0,
			freqType:  "invalid",
			expected:  "",
			err:       errors.New("unsupported uncore frequency type \"invalid\""),
		},
		{
			name:      "InitialMax",
			packageID: 1,
			dieID:     0,
			freqType:  "initial_max",
			expected:  "package_01_die_00/initial_max_freq_khz",
			err:       nil,
		},
		{
			name:      "InitialMin",
			packageID: 1,
			dieID:     0,
			freqType:  "initial_min",
			expected:  "package_01_die_00/initial_min_freq_khz",
			err:       nil,
		},
		{
			name:      "CustomizedMax",
			packageID: 1,
			dieID:     0,
			freqType:  "max",
			expected:  "package_01_die_00/max_freq_khz",
			err:       nil,
		},
		{
			name:      "CustomizedMin",
			packageID: 1,
			dieID:     0,
			freqType:  "min",
			expected:  "package_01_die_00/min_freq_khz",
			err:       nil,
		},
		{
			name:      "Current",
			packageID: 0,
			dieID:     1,
			freqType:  "current",
			expected:  "package_00_die_01/current_freq_khz",
			err:       nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uncoreFreqPath, err := getUncoreFreqPath(tc.packageID, tc.dieID, tc.freqType)
			if tc.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, uncoreFreqPath)
			}
		})
	}
}

func TestUncoreFreqData_Init(t *testing.T) {
	testCases := []struct {
		name           string
		uncoreFreqPath string
		err            error
	}{
		{
			name:           "UncoreFreqPathEmpty",
			uncoreFreqPath: "",
			err:            errors.New("base path of uncore frequency cannot be empty"),
		},
		{
			name:           "UncoreFreqPathNotExist",
			uncoreFreqPath: "/dummy/path",
			err:            errors.New("invalid base path of uncore frequency"),
		},
		{
			name:           "UncoreFreqPathValid",
			uncoreFreqPath: "testdata/intel_uncore_frequency",
			err:            nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uFreqData := &uncoreFreqData{
				uncoreFreqBasePath: tc.uncoreFreqPath,
			}

			err := uFreqData.init()
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetUncoreFrequencyMhz(t *testing.T) {
	testCases := []struct {
		name      string
		packageID int
		dieID     int
		freqType  string
		expected  float64
		err       error
	}{
		{
			name:      "InitialMax",
			packageID: 10,
			dieID:     3,
			freqType:  "initial_max",
			expected:  2000,
			err:       nil,
		},
		{
			name:      "InitialMin",
			packageID: 10,
			dieID:     3,
			freqType:  "initial_min",
			expected:  1000,
			err:       nil,
		},
		{
			name:      "CustomizedMax",
			packageID: 10,
			dieID:     3,
			freqType:  "max",
			expected:  1900,
			err:       nil,
		},
		{
			name:      "CustomizedMin",
			packageID: 10,
			dieID:     3,
			freqType:  "min",
			expected:  1100,
			err:       nil,
		},
		{
			name:      "Current",
			packageID: 10,
			dieID:     3,
			freqType:  "current",
			expected:  1500,
			err:       nil,
		},
		{
			name:      "InvalidInitialMaxValue",
			packageID: 9,
			dieID:     12,
			freqType:  "initial_max",
			expected:  0,
			err:       errors.New("failed to convert frequency file content to float64"),
		},
		{
			name:      "InvalidInitialMinValue",
			packageID: 9,
			dieID:     12,
			freqType:  "initial_min",
			expected:  0,
			err:       errors.New("failed to convert frequency file content to float64"),
		},
		{
			name:      "InvalidCustomizedMaxValue",
			packageID: 9,
			dieID:     12,
			freqType:  "max",
			expected:  0,
			err:       errors.New("failed to convert frequency file content to float64"),
		},
		{
			name:      "InvalidCustomizedMinValue",
			packageID: 9,
			dieID:     12,
			freqType:  "min",
			expected:  0,
			err:       errors.New("failed to convert frequency file content to float64"),
		},
		{
			name:      "InvalidFreqType",
			packageID: 9,
			dieID:     12,
			freqType:  "invalid",
			expected:  0,
			err:       errors.New("failed to get frequency path: unsupported uncore frequency type \"invalid\""),
		},
		{
			name:      "FreqTypeFileNotExist",
			packageID: 9,
			dieID:     12,
			freqType:  "current",
			expected:  0,
			err:       errors.New("failed to read frequency file: file \"testdata/intel_uncore_frequency/package_09_die_12/current_freq_khz\" does not exist"),
		},
	}

	u := &uncoreFreqData{
		uncoreFreqBasePath: "testdata/intel_uncore_frequency",
	}
	require.NoError(t, u.init())

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uncoreFreq, err := u.getUncoreFrequencyMhz(tc.packageID, tc.dieID, tc.freqType)
			if tc.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, uncoreFreq)
			}
		})
	}
}
