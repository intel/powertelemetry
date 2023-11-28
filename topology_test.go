// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type topologyMock struct {
	mock.Mock
}

func (m *topologyMock) initTopology() error {
	args := m.Called()
	return args.Error(0)
}

func (m *topologyMock) getCPUVendor(cpuID int) (string, error) {
	args := m.Called(cpuID)
	return args.String(0), args.Error(1)
}

func (m *topologyMock) getCPUFamily(cpuID int) (string, error) {
	args := m.Called(cpuID)
	return args.String(0), args.Error(1)
}

func (m *topologyMock) getCPUDieID(cpuID int) (int, error) {
	args := m.Called(cpuID)
	return args.Int(0), args.Error(1)
}

func (m *topologyMock) getCPUPackageID(cpuID int) (int, error) {
	args := m.Called(cpuID)
	return args.Int(0), args.Error(1)
}

func (m *topologyMock) getCPUCoreID(cpuID int) (int, error) {
	args := m.Called(cpuID)
	return args.Int(0), args.Error(1)
}

func (m *topologyMock) getCPUModel() int {
	args := m.Called()
	return args.Int(0)
}

func (m *topologyMock) getCPUFlags(cpuID int) ([]string, error) {
	args := m.Called(cpuID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *topologyMock) getCPUsNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *topologyMock) getPackageDieIDs(packageID int) ([]int, error) {
	args := m.Called(packageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int), args.Error(1)
}

func (m *topologyMock) getPackageIDs() []int {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func TestCpuFields(t *testing.T) {
	flagsExp := []string{"msr", "dts"}
	topology := &topologyData{
		topologyMap: map[int]*cpuInfo{
			0: {
				vendorID:  "vendorID",
				family:    "family",
				flags:     flagsExp,
				dieID:     2,
				packageID: 2,
			},
		},
	}

	t.Run("Validate cpu's family value", func(t *testing.T) {
		cpuID := 0
		expected := "family"
		res, err := topology.getCPUFamily(cpuID)
		require.NoError(t, err)
		require.EqualValues(t, expected, res)
	})

	t.Run("Validate cpu's vendorID value", func(t *testing.T) {
		cpuID := 0
		expected := "vendorID"
		res, err := topology.getCPUVendor(cpuID)
		require.NoError(t, err)
		require.EqualValues(t, expected, res)
	})

	t.Run("Family's value doesn't exist in map for given key", func(t *testing.T) {
		cpuID := 1
		expected := ""
		res, err := topology.getCPUFamily(cpuID)
		require.EqualValues(t, expected, res)
		require.Error(t, err)
		require.ErrorContains(t, err, fmt.Sprintf("cpu: %d doesn't exist", cpuID))
	})

	t.Run("VendorID's value doesn't exist in map for given key", func(t *testing.T) {
		cpuID := 1
		expected := ""
		res, err := topology.getCPUVendor(cpuID)
		require.Error(t, err)
		require.ErrorContains(t, err, fmt.Sprintf("cpu: %d doesn't exist", cpuID))
		require.EqualValues(t, expected, res)
	})

	t.Run("PackageID value doesn't exist in map for given key", func(t *testing.T) {
		cpuID := 1
		expected := 0
		res, err := topology.getCPUPackageID(cpuID)
		require.Error(t, err)
		require.ErrorContains(t, err, fmt.Sprintf("cpu: %d doesn't exist", cpuID))
		require.EqualValues(t, expected, res)
	})

	t.Run("CoreID value doesn't exist in map for given key", func(t *testing.T) {
		cpuID := 1000
		expected := 0
		res, err := topology.getCPUCoreID(cpuID)
		require.Error(t, err)
		require.ErrorContains(t, err, fmt.Sprintf("cpu: %d doesn't exist", cpuID))
		require.EqualValues(t, expected, res)
	})

	t.Run("DieID's value doesn't exist in map for given key", func(t *testing.T) {
		cpuID := 1
		expected := 0
		res, err := topology.getCPUDieID(cpuID)
		require.Error(t, err)
		require.ErrorContains(t, err, fmt.Sprintf("cpu: %d doesn't exist", cpuID))
		require.EqualValues(t, expected, res)
	})

	t.Run("Validate cpu's packageID value", func(t *testing.T) {
		cpuID := 0
		expected := 2
		res, err := topology.getCPUPackageID(cpuID)
		require.NoError(t, err)
		require.EqualValues(t, expected, res)
	})

	t.Run("Validate cpu's coreID value", func(t *testing.T) {
		cpuID := 0
		expected := 0
		res, err := topology.getCPUCoreID(cpuID)
		require.NoError(t, err)
		require.EqualValues(t, expected, res)
	})

	t.Run("Validate cpu's dieID value", func(t *testing.T) {
		cpuID := 0
		expected := 2
		res, err := topology.getCPUDieID(cpuID)
		require.NoError(t, err)
		require.EqualValues(t, expected, res)
	})

	t.Run("ValidCPUID", func(t *testing.T) {
		cpuID := 0
		expected := flagsExp
		res, err := topology.getCPUFlags(cpuID)
		require.NoError(t, err)
		require.EqualValues(t, expected, res)
	})

	t.Run("InvalidCPUID", func(t *testing.T) {
		cpuID := 1
		expected := []string(nil)
		res, err := topology.getCPUFlags(cpuID)
		require.Error(t, err)
		require.ErrorContains(t, err, fmt.Sprintf("cpu: %d doesn't exist", cpuID))
		require.EqualValues(t, expected, res)
	})
}

// TestExtractDieID checks if dieID value extracted from file is correct in different test cases.
func TestExtractDieID(t *testing.T) {
	testCases := []struct {
		desc          string
		filePath      string
		expectedDieID int
		err           error
	}{
		{
			desc:          "Extracted",
			filePath:      "testdata/die-id-valid/cpu1/topology/die_id",
			expectedDieID: 1,
			err:           nil,
		},
		{
			desc:          "EmptyFilename",
			filePath:      "",
			expectedDieID: 0,
			err:           errors.New("error opening file \"\""),
		},
		{
			desc:          "DirectoryInsteadOfFile",
			filePath:      "testdata",
			expectedDieID: 0,
			err:           errors.New("error reading file \"testdata\": error while reading file from path \"testdata\""),
		},
		{
			desc:          "FileNotExist",
			filePath:      "testdata/die-id-valid/cpu1/topology/die_id_badPath",
			expectedDieID: 0,
			err:           nil,
		},
		{
			desc:          "NotExtracted",
			filePath:      "testdata/die-id-invalid/cpu1/topology/die_id",
			expectedDieID: 0,
			err:           errors.New("error converting die ID value from the file \"testdata/die-id-invalid/cpu1/topology/die_id\" to int"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			resDieID, err := extractDieID(testCase.filePath)
			require.Equal(t, testCase.expectedDieID, resDieID)
			if testCase.err != nil {
				require.ErrorContains(t, err, testCase.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestInitTopology checks if cpuInfo value of topology struct is correct in different test cases.
func TestInitTopology(t *testing.T) {
	testCases := []struct {
		name           string
		cpuInfoPath    string
		diePath        string
		topologyMapExp map[int]*cpuInfo
		packageDiesExp map[int][]int
		packageIDsExp  []int
		err            error
	}{
		{
			name:        "InitializedWithValidDieIDPath",
			diePath:     "testdata/die-id-valid",
			cpuInfoPath: "testdata/cpuinfo_good/",
			topologyMapExp: map[int]*cpuInfo{
				1: {
					vendorID:  "IdOfVendor",
					family:    "13",
					dieID:     1,
					packageID: 2,
					coreID:    66,
					flags:     []string{"no", "flags"},
				},
			},
			packageDiesExp: map[int][]int{
				2: {
					1,
				},
			},
			packageIDsExp: []int{2},
			err:           nil,
		},
		{
			name:        "InitializedWithInvalidDieIDPath",
			diePath:     "testdata/die-id-invalid",
			cpuInfoPath: "testdata/cpuinfo_good/",
			topologyMapExp: map[int]*cpuInfo{
				1: {
					vendorID:  "IdOfVendor",
					family:    "13",
					dieID:     0,
					packageID: 2,
					coreID:    66,
					flags:     []string{"no", "flags"},
				},
			},
			packageDiesExp: map[int][]int{},
			packageIDsExp:  []int{2},
			err:            nil,
		},
		{
			name:        "InitializedWithoutDieIDPath",
			cpuInfoPath: "testdata/cpuinfo_good/",
			topologyMapExp: map[int]*cpuInfo{
				1: {
					vendorID:  "IdOfVendor",
					family:    "13",
					dieID:     0,
					packageID: 2,
					coreID:    66,
					flags:     []string{"no", "flags"},
				},
			},
			packageDiesExp: map[int][]int{
				2: {
					0,
				},
			},
			packageIDsExp: []int{2},
			err:           nil,
		},
		{
			name:           "InvalidProcessorField",
			diePath:        "testdata/die-id-valid",
			cpuInfoPath:    "testdata/cpuinfo_bad1/",
			topologyMapExp: nil,
			err:            errors.New("error occurred while parsing CPU information"),
		},
		{
			name:           "InvalidSteppingField",
			diePath:        "testdata/die-id-valid",
			cpuInfoPath:    "testdata/cpuinfo_bad2/",
			topologyMapExp: nil,
			err:            errors.New("error occurred while parsing CPU information"),
		},
		{
			name:           "InvalidCacheSizeField",
			diePath:        "testdata/die-id-valid",
			cpuInfoPath:    "testdata/cpuinfo_bad3/",
			topologyMapExp: nil,
			err:            errors.New("error occurred while parsing CPU information"),
		},
		{
			name:           "InvalidCpuInfoPath",
			diePath:        "testdata/die-id-valid",
			cpuInfoPath:    "testdata/cpuinfo_bad_path",
			topologyMapExp: nil,
			err:            errors.New("no CPUs were found"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("HOST_PROC", tc.cpuInfoPath)

			newTopology := &topologyData{
				dieIDPath: tc.diePath,
			}

			err := newTopology.initTopology()
			require.Equal(t, tc.topologyMapExp, newTopology.topologyMap)
			require.Equal(t, tc.packageDiesExp, newTopology.packageDies)
			require.Equal(t, tc.packageIDsExp, newTopology.getPackageIDs())
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetCPUsNumber(t *testing.T) {
	testCases := []struct {
		name         string
		topology     topologyReader
		numberOfCPUs int
	}{
		{
			name: "0_CPUs",
			topology: &topologyData{
				topologyMap: make(map[int]*cpuInfo),
			},
			numberOfCPUs: 0,
		},
		{
			name: "3_CPUs",
			topology: &topologyData{
				topologyMap: map[int]*cpuInfo{
					0: nil,
					1: nil,
					2: nil,
				},
			},
			numberOfCPUs: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.topology.getCPUsNumber()
			require.Equal(t, tc.numberOfCPUs, actual)
		})
	}
}

func TestPackageIDs(t *testing.T) {
	testCases := []struct {
		name        string
		topologyMap map[int]*cpuInfo
		packageIDs  []int
	}{
		{
			name:        "EmptyTopologyMap",
			topologyMap: map[int]*cpuInfo{},
			packageIDs:  []int{},
		},
		{
			name: "Found",
			topologyMap: map[int]*cpuInfo{
				0: {
					packageID: 0,
				},
				1: {
					packageID: 1,
				},
				2: {
					packageID: 0,
				},
				3: {
					packageID: 0,
				},
				4: {
					packageID: 1,
				},
				5: {
					packageID: 2,
				},
			},
			packageIDs: []int{0, 1, 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.packageIDs, packageIDs(tc.topologyMap))
		})
	}
}
