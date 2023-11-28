// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/intel/powertelemetry/internal/cpumodel"
)

// msrMock represents a mock for msrDataWithStorage type. Implements msrReaderWithStorage interface.
type msrMock struct {
	mock.Mock
}

func (m *msrMock) initMsrMap(cpuIDs []int, timeout time.Duration) error {
	args := m.Called(cpuIDs, timeout)
	return args.Error(0)
}

func (m *msrMock) read(offset uint32, cpuID int) (uint64, error) {
	args := m.Called(offset, cpuID)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *msrMock) isMsrLoaded(modulesPath string) (bool, error) {
	args := m.Called(modulesPath)
	return args.Bool(0), args.Error(1)
}

func (m *msrMock) update(cpuID int) error {
	args := m.Called(cpuID)
	return args.Error(0)
}

func (m *msrMock) getOffsetDeltas(cpuID int) (map[uint32]uint64, error) {
	args := m.Called(cpuID)
	return args.Get(0).(map[uint32]uint64), args.Error(1)
}

func (m *msrMock) getTimestampDelta(cpuID int) (time.Duration, error) {
	args := m.Called(cpuID)
	return args.Get(0).(time.Duration), args.Error(1)
}

type coreFreqMock struct {
	mock.Mock
}

func (m *coreFreqMock) init() error {
	args := m.Called()
	return args.Error(0)
}

func (m *coreFreqMock) getCPUFrequencyMhz(cpuID int) (float64, error) {
	args := m.Called(cpuID)
	return args.Get(0).(float64), args.Error(1)
}

// uncoreFreqMock represents a mock for uncoreFreqData type. Implements uncoreFreqReader interface.
type uncoreFreqMock struct {
	mock.Mock
}

func (m *uncoreFreqMock) init() error {
	args := m.Called()
	return args.Error(0)
}

func (m *uncoreFreqMock) getUncoreFrequencyMhz(packageID, dieID int, freqType string) (float64, error) {
	args := m.Called(packageID, dieID, freqType)
	return args.Get(0).(float64), args.Error(1)
}

func TestGetInitialUncoreFrequencyMin(t *testing.T) {
	pt := &PowerTelemetry{
		uncoreFreq: &uncoreFreqData{
			uncoreFreqBasePath: "testdata/intel_uncore_frequency",
		},
	}

	t.Run("UncoreFreqIsNil", func(t *testing.T) {
		packageID := 0
		dieID := 0
		freqExp := 0.0

		ptel := &PowerTelemetry{}

		freqOut, err := ptel.GetInitialUncoreFrequencyMin(packageID, dieID)
		require.Equal(t, freqExp, freqOut)
		require.ErrorContains(t, err, "\"uncore_frequency\" is not initialized")
	})

	t.Run("FreqFileNotExist", func(t *testing.T) {
		packageID := 10
		dieID := 5
		freqExp := 0.0

		freqOut, err := pt.GetInitialUncoreFrequencyMin(packageID, dieID)
		require.ErrorContains(t, err, "failed to read frequency file")
		require.Equal(t, freqExp, freqOut)
	})

	t.Run("InvalidFreqValue", func(t *testing.T) {
		packageID := 9
		dieID := 12
		freqExp := 0.0

		freqOut, err := pt.GetInitialUncoreFrequencyMin(packageID, dieID)
		require.ErrorContains(t, err, "failed to convert frequency file content to float64")
		require.Equal(t, freqExp, freqOut)
	})

	t.Run("Valid", func(t *testing.T) {
		packageID := 10
		dieID := 3
		freqExp := 1000.0

		freqOut, err := pt.GetInitialUncoreFrequencyMin(packageID, dieID)
		require.NoError(t, err)
		require.Equal(t, freqExp, freqOut)
	})
}

func TestGetCustomizedUncoreFrequencyMin(t *testing.T) {
	pt := &PowerTelemetry{
		uncoreFreq: &uncoreFreqData{
			uncoreFreqBasePath: "testdata/intel_uncore_frequency",
		},
	}

	t.Run("UncoreFreqIsNil", func(t *testing.T) {
		packageID := 0
		dieID := 0
		freqExp := 0.0

		ptel := &PowerTelemetry{}

		freqOut, err := ptel.GetCustomizedUncoreFrequencyMin(packageID, dieID)
		require.Equal(t, freqExp, freqOut)
		require.ErrorContains(t, err, "\"uncore_frequency\" is not initialized")
	})

	t.Run("FreqFileNotExist", func(t *testing.T) {
		packageID := 10
		dieID := 5
		freqExp := 0.0

		freqOut, err := pt.GetCustomizedUncoreFrequencyMin(packageID, dieID)
		require.ErrorContains(t, err, "failed to read frequency file")
		require.Equal(t, freqExp, freqOut)
	})

	t.Run("InvalidFreqValue", func(t *testing.T) {
		packageID := 9
		dieID := 12
		freqExp := 0.0

		freqOut, err := pt.GetCustomizedUncoreFrequencyMin(packageID, dieID)
		require.ErrorContains(t, err, "failed to convert frequency file content to float64")
		require.Equal(t, freqExp, freqOut)
	})

	t.Run("Valid", func(t *testing.T) {
		packageID := 10
		dieID := 3
		freqExp := 1100.0

		freqOut, err := pt.GetCustomizedUncoreFrequencyMin(packageID, dieID)
		require.NoError(t, err)
		require.Equal(t, freqExp, freqOut)
	})
}

func TestGetInitialUncoreFrequencyMax(t *testing.T) {
	pt := &PowerTelemetry{
		uncoreFreq: &uncoreFreqData{
			uncoreFreqBasePath: "testdata/intel_uncore_frequency",
		},
	}

	t.Run("UncoreFreqIsNil", func(t *testing.T) {
		packageID := 0
		dieID := 0
		freqExp := 0.0

		ptel := &PowerTelemetry{}

		freqOut, err := ptel.GetInitialUncoreFrequencyMax(packageID, dieID)
		require.Equal(t, freqExp, freqOut)
		require.ErrorContains(t, err, "\"uncore_frequency\" is not initialized")
	})

	t.Run("FreqFileNotExist", func(t *testing.T) {
		packageID := 10
		dieID := 5
		freqExp := 0.0

		freqOut, err := pt.GetInitialUncoreFrequencyMax(packageID, dieID)
		require.ErrorContains(t, err, "failed to read frequency file")
		require.Equal(t, freqExp, freqOut)
	})

	t.Run("InvalidFreqValue", func(t *testing.T) {
		packageID := 9
		dieID := 12
		freqExp := 0.0

		freqOut, err := pt.GetInitialUncoreFrequencyMax(packageID, dieID)
		require.ErrorContains(t, err, "failed to convert frequency file content to float64")
		require.Equal(t, freqExp, freqOut)
	})

	t.Run("Valid", func(t *testing.T) {
		packageID := 10
		dieID := 3
		freqExp := 2000.0

		freqOut, err := pt.GetInitialUncoreFrequencyMax(packageID, dieID)
		require.NoError(t, err)
		require.Equal(t, freqExp, freqOut)
	})
}

func TestGetCustomizedUncoreFrequencyMax(t *testing.T) {
	pt := &PowerTelemetry{
		uncoreFreq: &uncoreFreqData{
			uncoreFreqBasePath: "testdata/intel_uncore_frequency",
		},
	}

	t.Run("UncoreFreqIsNil", func(t *testing.T) {
		packageID := 0
		dieID := 0
		freqExp := 0.0

		ptel := &PowerTelemetry{}

		freqOut, err := ptel.GetCustomizedUncoreFrequencyMax(packageID, dieID)
		require.Equal(t, freqExp, freqOut)
		require.ErrorContains(t, err, "\"uncore_frequency\" is not initialized")
	})

	t.Run("FreqFileNotExist", func(t *testing.T) {
		packageID := 10
		dieID := 5
		freqExp := 0.0

		freqOut, err := pt.GetCustomizedUncoreFrequencyMax(packageID, dieID)
		require.ErrorContains(t, err, "failed to read frequency file")
		require.Equal(t, freqExp, freqOut)
	})

	t.Run("InvalidFreqValue", func(t *testing.T) {
		packageID := 9
		dieID := 12
		freqExp := 0.0

		freqOut, err := pt.GetCustomizedUncoreFrequencyMax(packageID, dieID)
		require.ErrorContains(t, err, "failed to convert frequency file content to float64")
		require.Equal(t, freqExp, freqOut)
	})

	t.Run("Valid", func(t *testing.T) {
		packageID := 10
		dieID := 3
		freqExp := 1900.0

		freqOut, err := pt.GetCustomizedUncoreFrequencyMax(packageID, dieID)
		require.NoError(t, err)
		require.Equal(t, freqExp, freqOut)
	})
}

func TestGetCurrentUncoreFrequency(t *testing.T) {
	newTopology := &topologyData{
		topologyMap: map[int]*cpuInfo{
			0: {
				vendorID:  "IdOfVendor",
				family:    "13",
				packageID: 0,
			},
			1: {
				vendorID:  "IdOfVendor",
				family:    "13",
				packageID: 1,
			},
			2: {
				vendorID:  "IdOfVendor",
				family:    "13",
				packageID: 0,
			},
		},
	}

	t.Run("FromFileSystem", func(t *testing.T) {
		packageID := 0
		dieID := 1
		uncoreFreqExp := 2000.0

		m := &uncoreFreqMock{}

		// mock getting the current uncore frequency
		m.On("getUncoreFrequencyMhz", packageID, dieID, "current").Return(uncoreFreqExp, nil).Once()

		pt := &PowerTelemetry{
			uncoreFreq: m,
		}

		uncoreFreqOut, err := pt.GetCurrentUncoreFrequency(packageID, dieID)

		require.NoError(t, err)
		require.Equal(t, uncoreFreqExp, uncoreFreqOut)
		m.AssertExpectations(t)
	})

	t.Run("FromMsr", func(t *testing.T) {
		t.Run("UncoreFreqFailed", func(t *testing.T) {
			packageID := 1
			dieID := 0
			msrValue := uint64(0xffffff08)
			uncoreFreqExp := 800.0

			mUncoreFreq := &uncoreFreqMock{}

			// mock getting current uncore frequency
			mUncoreFreq.On("getUncoreFrequencyMhz", packageID, dieID, "current").Return(0.0, errors.New("failed to read current uncore frequency file")).Once()

			mMsr := &msrMock{}

			// mock reading msr offset UNCORE_PERF_STATUS of CPU ID 1
			mMsr.On("read", uint32(uncorePerfStatus), 1).Return(msrValue, nil).Once()

			pt := &PowerTelemetry{
				topology:   newTopology,
				uncoreFreq: mUncoreFreq,
				msr:        mMsr,

				cpus: []int{0, 1, 2},
			}

			uncoreFreqOut, err := pt.GetCurrentUncoreFrequency(packageID, dieID)

			require.Equal(t, uncoreFreqExp, uncoreFreqOut)
			require.NoError(t, err)
			mUncoreFreq.AssertExpectations(t)
			mMsr.AssertExpectations(t)
		})

		t.Run("UncoreFreqIsNil", func(t *testing.T) {
			packageID := 1
			dieID := 0
			msrValue := uint64(0xffffff08)
			uncoreFreqExp := 800.0

			mMsr := &msrMock{}

			// mock reading msr offset UNCORE_PERF_STATUS of CPU ID 1
			mMsr.On("read", uint32(uncorePerfStatus), 1).Return(msrValue, nil).Once()

			pt := &PowerTelemetry{
				topology: newTopology,
				msr:      mMsr,

				cpus: []int{0, 1, 2},
			}

			uncoreFreqOut, err := pt.GetCurrentUncoreFrequency(packageID, dieID)

			require.Equal(t, uncoreFreqExp, uncoreFreqOut)
			require.NoError(t, err)
			mMsr.AssertExpectations(t)
		})
	})

	t.Run("FromMsrFailed", func(t *testing.T) {
		t.Run("MsrIsNil", func(t *testing.T) {
			packageID := 0
			dieID := 1
			uncoreFreqExp := 0.0

			pt := &PowerTelemetry{}

			uncoreFreqOut, err := pt.GetCurrentUncoreFrequency(packageID, dieID)

			require.ErrorContains(t, err, "\"msr\" is not initialized")
			require.Equal(t, uncoreFreqExp, uncoreFreqOut)
		})

		t.Run("PackageIDNotFound", func(t *testing.T) {
			packageID := 2
			dieID := 1
			uncoreFreqExp := 0.0

			mMsr := &msrMock{}
			mUncoreFreq := &uncoreFreqMock{}

			// mock getting current uncore frequency
			mUncoreFreq.On("getUncoreFrequencyMhz", packageID, dieID, "current").Return(0.0, errors.New("failed to read current uncore frequency file")).Once()

			pt := &PowerTelemetry{
				topology:   newTopology,
				uncoreFreq: mUncoreFreq,
				msr:        mMsr,

				cpus: []int{0, 1, 2},
			}

			uncoreFreqOut, err := pt.GetCurrentUncoreFrequency(packageID, dieID)
			require.Equal(t, uncoreFreqExp, uncoreFreqOut)
			require.ErrorContains(t, err, "unable to get CPU ID for package ID: 2")
			mUncoreFreq.AssertExpectations(t)
			mMsr.AssertExpectations(t)
		})

		t.Run("ReadMsrError", func(t *testing.T) {
			packageID := 1
			dieID := 0
			uncoreFreqExp := 0.0
			errMsg := "error reading msr file"

			mUncoreFreq := &uncoreFreqMock{}

			// mock getting current uncore frequency
			mUncoreFreq.On("getUncoreFrequencyMhz", packageID, dieID, "current").Return(0.0, errors.New("failed to read current uncore frequency file")).Once()

			mMsr := &msrMock{}

			// mock reading msr offset UNCORE_PERF_STATUS of CPU ID 1
			mMsr.On("read", uint32(uncorePerfStatus), 1).Return(uint64(0), errors.New(errMsg)).Once()

			pt := &PowerTelemetry{
				topology:   newTopology,
				uncoreFreq: mUncoreFreq,
				msr:        mMsr,

				cpus: []int{0, 1, 2},
			}

			uncoreFreqOut, err := pt.GetCurrentUncoreFrequency(packageID, dieID)

			require.Equal(t, uncoreFreqExp, uncoreFreqOut)
			require.ErrorContains(t, err, errMsg)
			mUncoreFreq.AssertExpectations(t)
			mMsr.AssertExpectations(t)
		})
	})
}

func TestGetCPUFrequency(t *testing.T) {
	pt := &PowerTelemetry{
		cpuFreq: &cpuFreqData{
			cpuFrequencyFilePath: "testdata/cpu-freq",
		},
	}

	t.Run("CPUFreqIsNil", func(t *testing.T) {
		expectedFreq := 0.0

		ptel := &PowerTelemetry{}

		actualFreq, err := ptel.GetCPUFrequency(0)
		require.Equal(t, expectedFreq, actualFreq)
		require.ErrorContains(t, err, "\"cpu_frequency\" is not initialized")
	})

	t.Run("Valid", func(t *testing.T) {
		expectedFreq := 888.888
		actualFreq, err := pt.GetCPUFrequency(0)
		require.Equal(t, expectedFreq, actualFreq)
		require.NoError(t, err)
	})

	t.Run("Invalid", func(t *testing.T) {
		expectedFreq := 0.0
		expectedError := "error reading file"
		actualFreq, err := pt.GetCPUFrequency(1)
		require.Equal(t, expectedFreq, actualFreq)
		require.ErrorContains(t, err, expectedError)
	})
}

func TestGetBusClock(t *testing.T) {
	type msrFreqTuple struct {
		msrValue uint64
		freq     float64
	}

	models100 := []int{
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

	models133 := []int{
		0x1E, // INTEL_FAM6_NEHALEM
		0x1F, // INTEL_FAM6_NEHALEM_G
		0x1A, // INTEL_FAM6_NEHALEM_EP
		0x2E, // INTEL_FAM6_NEHALEM_EX
		0x25, // INTEL_FAM6_WESTMERE
		0x2C, // INTEL_FAM6_WESTMERE_EP
		0x2F, // INTEL_FAM6_WESTMERE_EX
	}

	modelsSilvermont := []int{
		0x37, // INTEL_FAM6_ATOM_SILVERMONT
		0x4D, // INTEL_FAM6_ATOM_SILVERMONT_D
		0x4A, // INTEL_FAM6_ATOM_SILVERMONT_MID
		0x5A, // INTEL_FAM6_ATOM_SILVERMONT_SMARTPHONE
	}

	modelsAirmont := []int{
		0x4C, // INTEL_FAM6_ATOM_AIRMONT
	}

	supportedModels := make([]int, 0)
	supportedModels = append(supportedModels, models100...)
	supportedModels = append(supportedModels, models133...)
	supportedModels = append(supportedModels, modelsSilvermont...)
	supportedModels = append(supportedModels, modelsAirmont...)

	supportedModelMap := map[int]interface{}{}
	for _, m := range supportedModels {
		supportedModelMap[m] = struct{}{}
	}

	t.Run("BusClockSilvermont", func(t *testing.T) {
		t.Run("NoCPUsAvailable", func(t *testing.T) {
			busClockExp := 0.0

			for _, model := range modelsSilvermont {
				pt := &PowerTelemetry{
					cpus: []int{}, // no CPU IDs available
				}

				busClockOut, err := pt.getBusClock(model)
				require.Equal(t, busClockExp, busClockOut)
				require.ErrorContains(t, err, "no available CPUs were found")
			}
		})

		t.Run("MsrIsNil", func(t *testing.T) {
			busClockExp := 0.0

			for _, model := range modelsSilvermont {
				pt := &PowerTelemetry{
					cpus: []int{0},
				}

				busClockOut, err := pt.getBusClock(model)
				require.Equal(t, busClockExp, busClockOut)
				require.ErrorContains(t, err, "\"msr\" is not initialized")
			}
		})

		t.Run("FailedToReadMsr", func(t *testing.T) {
			cpuID := 0
			busClockExp := 0.0
			mError := errors.New("mock error")

			for _, model := range modelsSilvermont {
				mMsr := msrMock{}
				mMsr.On("read", uint32(fsbFreq), cpuID).Return(uint64(0), mError).Once()

				pt := &PowerTelemetry{
					msr:  &mMsr,
					cpus: []int{cpuID},
				}

				busClockOut, err := pt.getBusClock(model)
				require.Equal(t, busClockExp, busClockOut)
				require.ErrorContains(t, err, mError.Error())
				mMsr.AssertExpectations(t)
			}
		})

		t.Run("InvalidFrequencyIndex", func(t *testing.T) {
			cpuID := 0
			busClockExp := 0.0

			for _, model := range modelsSilvermont {
				mMsr := msrMock{}
				mMsrValue := uint64(0xF5)
				mMsr.On("read", uint32(fsbFreq), cpuID).Return(mMsrValue, nil).Once()

				pt := &PowerTelemetry{
					msr:  &mMsr,
					cpus: []int{cpuID},
				}

				busClockOut, err := pt.getBusClock(model)
				require.Equal(t, busClockExp, busClockOut)
				require.ErrorContains(t, err, fmt.Sprintf("error while getting bus clock: index %d is outside of bounds", 5))
				mMsr.AssertExpectations(t)
			}
		})

		t.Run("Ok", func(t *testing.T) {
			cpuID := 0
			silvermontTuples := []msrFreqTuple{
				{0x00, 83.3},
				{0x01, 100.0},
				{0x02, 133.3},
				{0x03, 116.7},
				{0x04, 80.0},
			}
			for _, model := range modelsSilvermont {
				for _, tuple := range silvermontTuples {
					mMsr := msrMock{}
					mMsrValue := tuple.msrValue
					mMsr.On("read", uint32(fsbFreq), cpuID).Return(mMsrValue, nil).Once()

					pt := &PowerTelemetry{
						msr:  &mMsr,
						cpus: []int{cpuID},
					}

					busClockExp := tuple.freq
					busClockOut, err := pt.getBusClock(model)
					require.Equal(t, busClockExp, busClockOut)
					require.NoError(t, err)
					mMsr.AssertExpectations(t)
				}
			}
		})
	})

	t.Run("BusClockAirmont", func(t *testing.T) {
		t.Run("NoCPUsAvailable", func(t *testing.T) {
			busClockExp := 0.0

			for _, model := range modelsAirmont {
				pt := &PowerTelemetry{
					cpus: []int{}, // no CPU IDs available
				}

				busClockOut, err := pt.getBusClock(model)
				require.Equal(t, busClockExp, busClockOut)
				require.ErrorContains(t, err, "no available CPUs were found")
			}
		})

		t.Run("MsrIsNil", func(t *testing.T) {
			busClockExp := 0.0

			for _, model := range modelsAirmont {
				pt := &PowerTelemetry{
					cpus: []int{0},
				}

				busClockOut, err := pt.getBusClock(model)
				require.Equal(t, busClockExp, busClockOut)
				require.ErrorContains(t, err, "\"msr\" is not initialized")
			}
		})

		t.Run("FailedToReadMsr", func(t *testing.T) {
			cpuID := 0
			busClockExp := 0.0
			mError := errors.New("mock error")

			for _, model := range modelsAirmont {
				mMsr := msrMock{}
				mMsr.On("read", uint32(fsbFreq), cpuID).Return(uint64(0), mError).Once()

				pt := &PowerTelemetry{
					msr:  &mMsr,
					cpus: []int{cpuID},
				}

				busClockOut, err := pt.getBusClock(model)
				require.Equal(t, busClockExp, busClockOut)
				require.ErrorContains(t, err, mError.Error())
				mMsr.AssertExpectations(t)
			}
		})

		t.Run("InvalidFrequencyIndex", func(t *testing.T) {
			cpuID := 0
			busClockExp := 0.0

			for _, model := range modelsAirmont {
				mMsr := msrMock{}
				mMsrValue := uint64(0xF9)
				mMsr.On("read", uint32(fsbFreq), cpuID).Return(mMsrValue, nil).Once()

				pt := &PowerTelemetry{
					msr:  &mMsr,
					cpus: []int{cpuID},
				}

				busClockOut, err := pt.getBusClock(model)
				require.Equal(t, busClockExp, busClockOut)
				require.ErrorContains(t, err, fmt.Sprintf("error while getting bus clock: index %d is outside of bounds", 9))
				mMsr.AssertExpectations(t)
			}
		})

		t.Run("Ok", func(t *testing.T) {
			cpuID := 0
			airmontTuples := []msrFreqTuple{
				{0x00, 83.3},
				{0x01, 100.0},
				{0x02, 133.3},
				{0x03, 116.7},
				{0x04, 80.0},
				{0x05, 93.3},
				{0x06, 90.0},
				{0x07, 88.9},
				{0x08, 87.5},
			}
			for _, model := range modelsAirmont {
				for _, tuple := range airmontTuples {
					mMsr := msrMock{}
					mMsrValue := tuple.msrValue
					mMsr.On("read", uint32(fsbFreq), cpuID).Return(mMsrValue, nil).Once()

					pt := &PowerTelemetry{
						msr:  &mMsr,
						cpus: []int{cpuID},
					}

					busClockExp := tuple.freq
					busClockOut, err := pt.getBusClock(model)
					require.Equal(t, busClockExp, busClockOut)
					require.NoError(t, err)
					mMsr.AssertExpectations(t)
				}
			}
		})
	})

	t.Run("BusClock100.0", func(t *testing.T) {
		busClockExp := 100.0
		pt := &PowerTelemetry{}

		for _, model := range models100 {
			busClockOut, err := pt.getBusClock(model)
			require.NoError(t, err)
			require.Equalf(t, busClockExp, busClockOut, "Model 0x%X", model)
		}
	})

	t.Run("BusClock133.0", func(t *testing.T) {
		busClockExp := 133.0
		pt := &PowerTelemetry{}

		for _, model := range models133 {
			busClockOut, err := pt.getBusClock(model)
			require.NoError(t, err)
			require.Equalf(t, busClockExp, busClockOut, "Model 0x%X", model)
		}
	})

	t.Run("UnsupportedModels", func(t *testing.T) {
		busClockExp := 0.0
		pt := &PowerTelemetry{}

		for model := 0; model < 0xFF; model++ {
			if supportedModelMap[model] == nil {
				busClockOut, err := pt.getBusClock(model)
				require.Equalf(t, busClockExp, busClockOut, "Model 0x%X", model)
				require.ErrorContains(t, err, fmt.Sprintf("busClock is not supported by the CPU model: %v", model))
			}
		}
	})
}

func TestGetCPUTemperature(t *testing.T) {
	t.Run("MsrIsNil", func(t *testing.T) {
		cpuID := 0

		tempExp := uint64(0)

		pt := &PowerTelemetry{}

		tempOut, err := pt.GetCPUTemperature(cpuID)
		require.Equal(t, tempExp, tempOut)
		require.ErrorContains(t, err, "\"msr\" is not initialized")
	})

	t.Run("FailedToReadTemperatureTarget", func(t *testing.T) {
		cpuID := 0
		mError := errors.New("mock error")

		tempExp := uint64(0)

		m := &msrMock{}
		m.On("read", uint32(temperatureTarget), cpuID).Return(uint64(0), mError).Once()

		pt := &PowerTelemetry{
			topology: &topologyData{
				model: cpumodel.INTEL_FAM6_SAPPHIRERAPIDS_X,
			},
			msr: m,
		}

		temOut, err := pt.GetCPUTemperature(cpuID)
		require.Equal(t, tempExp, temOut)
		require.ErrorContains(t, err, mError.Error())
		m.AssertExpectations(t)
	})

	t.Run("ModelNotSupported", func(t *testing.T) {
		cpuID := 0
		// CPU temp metric not supported by this model.
		cpuModel := cpumodel.INTEL_FAM6_GRANITERAPIDS_D

		tempExp := uint64(0)

		m := &msrMock{}

		pt := &PowerTelemetry{
			msr: m,
			topology: &topologyData{
				model: cpuModel,
			},
		}

		tempOut, err := pt.GetCPUTemperature(cpuID)
		require.Equal(t, tempExp, tempOut)
		require.ErrorContains(t, err, fmt.Sprintf("cpu temperature metric not supported by CPU model: 0x%X", cpuModel))
		m.AssertExpectations(t)
	})

	t.Run("FailedToReadThermalStatus", func(t *testing.T) {
		cpuID := 0
		tempTargetValue := uint64(0x680a00)
		mError := errors.New("mock error")

		tempExp := uint64(0)

		m := &msrMock{}
		m.On("read", uint32(temperatureTarget), cpuID).Return(tempTargetValue, nil).Once()
		m.On("read", uint32(thermalStatus), cpuID).Return(uint64(0), mError).Once()

		pt := &PowerTelemetry{
			topology: &topologyData{
				model: cpumodel.INTEL_FAM6_SAPPHIRERAPIDS_X,
			},
			msr: m,
		}

		temOut, err := pt.GetCPUTemperature(cpuID)
		require.Equal(t, tempExp, temOut)
		require.ErrorContains(t, err, mError.Error())
		m.AssertExpectations(t)
	})

	t.Run("Temp23Celsius", func(t *testing.T) {
		cpuID := 0
		tempTargetValue := uint64(0x680a00)
		thermalStatusValue := uint64(0x88510000)

		tempExp := uint64(23)

		m := &msrMock{}
		m.On("read", uint32(temperatureTarget), cpuID).Return(tempTargetValue, nil).Once()
		m.On("read", uint32(thermalStatus), cpuID).Return(thermalStatusValue, nil).Once()

		pt := &PowerTelemetry{
			topology: &topologyData{
				model: cpumodel.INTEL_FAM6_EMERALDRAPIDS_X,
			},
			msr: m,
		}

		temOut, err := pt.GetCPUTemperature(cpuID)
		require.Equal(t, tempExp, temOut)
		require.NoError(t, err)
		m.AssertExpectations(t)
	})

	t.Run("Temp36Celsius", func(t *testing.T) {
		cpuID := 0
		tempTargetValue := uint64(0x630a00)
		thermalStatusValue := uint64(0x883f0800)

		tempExp := uint64(36)

		m := &msrMock{}
		m.On("read", uint32(temperatureTarget), cpuID).Return(tempTargetValue, nil).Once()
		m.On("read", uint32(thermalStatus), cpuID).Return(thermalStatusValue, nil).Once()

		pt := &PowerTelemetry{
			topology: &topologyData{
				model: cpumodel.INTEL_FAM6_EMERALDRAPIDS_X,
			},
			msr: m,
		}

		temOut, err := pt.GetCPUTemperature(cpuID)
		require.Equal(t, tempExp, temOut)
		require.NoError(t, err)
		m.AssertExpectations(t)
	})
}

func TestGetCPUBaseFrequency(t *testing.T) {
	t.Run("MsrIsNil", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0

		// expected output
		expectedFreq := uint64(0)

		// power telemetry instance definition
		pt := PowerTelemetry{
			topology: &topologyData{},
		}

		actualFreq, err := pt.GetCPUBaseFrequency(cpuID)
		require.ErrorContains(t, err, "\"msr\" is not initialized")
		require.Equal(t, expectedFreq, actualFreq)
	})

	t.Run("NoCPUIDsAvailable", func(t *testing.T) {
		// input arguments for test case
		packageID := 0
		busClk := 100.0

		// expected output
		expectedFreq := uint64(0)

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {
					packageID: 0,
				},
				1: {
					packageID: 0,
				},
				2: {
					packageID: 0,
				},
				3: {
					packageID: 0,
				},
				4: {
					packageID: 0,
				},
			},
			model: cpumodel.INTEL_FAM6_EMERALDRAPIDS_X,
		}

		// msr definition
		m := &msrMock{}

		// power telemetry instance definition
		pt := PowerTelemetry{
			topology: topo,
			msr:      m,
			cpus:     []int{}, // no cpus available
			busClock: busClk,
		}

		actualFreq, err := pt.GetCPUBaseFrequency(packageID)
		require.ErrorContains(t, err, fmt.Sprintf("could not find CPU ID for package ID %v", packageID))
		require.Equal(t, expectedFreq, actualFreq)
		m.AssertExpectations(t)
	})

	t.Run("CPUIDNotAvailable", func(t *testing.T) {
		// input arguments for test case
		packageID := 1
		busClk := 100.0

		// expected output
		expectedFreq := uint64(0)

		// topology definition
		topo := &topologyData{
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
					packageID: 1,
				},
				4: {
					packageID: 0,
				},
			},
			model: cpumodel.INTEL_FAM6_EMERALDRAPIDS_X,
		}

		// msr definition
		m := &msrMock{}

		// power telemetry instance definition
		pt := PowerTelemetry{
			topology: topo,
			msr:      m,
			cpus:     []int{0, 2, 4},
			busClock: busClk,
		}

		actualFreq, err := pt.GetCPUBaseFrequency(packageID)
		require.ErrorContains(t, err, fmt.Sprintf("could not find CPU ID for package ID %v", packageID))
		require.Equal(t, expectedFreq, actualFreq)
		m.AssertExpectations(t)
	})

	t.Run("BaseFreq", func(t *testing.T) {
		// input arguments for test case
		cpuID := 2
		packageID := 1
		cpuModel := cpumodel.INTEL_FAM6_ATOM_SILVERMONT
		msrValue := uint64(0x1234)
		busClk := 100.0

		// expected output
		expectedFreq := uint64(1800)

		// msr definition
		m := &msrMock{}
		m.On("read", uint32(platformInfo), cpuID).Return(msrValue, nil)

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {
					packageID: 0,
				},
				1: {
					packageID: 0,
				},
				2: {
					packageID: packageID,
				},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := PowerTelemetry{
			msr:      m,
			topology: topo,
			cpus:     []int{0, 1, 2},
			busClock: busClk,
		}

		actualFreq, err := pt.GetCPUBaseFrequency(packageID)
		require.NoError(t, err)
		require.Equal(t, expectedFreq, actualFreq)
		m.AssertExpectations(t)
	})

	t.Run("BaseFreqFractional", func(t *testing.T) {
		// input arguments for test case
		packageID := 0
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_ATOM_SILVERMONT
		msrValue := uint64(0x1234)
		busClk := 116.7

		// expected output
		expectedFreq := uint64(2100)

		// msr definition
		m := &msrMock{}
		m.On("read", uint32(platformInfo), cpuID).Return(msrValue, nil)

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {
					packageID: packageID,
				},
				1: {
					packageID: packageID,
				},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := PowerTelemetry{
			msr:      m,
			topology: topo,
			busClock: busClk,
			cpus:     []int{0, 1},
		}

		actualFreq, err := pt.GetCPUBaseFrequency(packageID)
		require.NoError(t, err)
		require.Equal(t, expectedFreq, actualFreq)
		m.AssertExpectations(t)
	})

	t.Run("UnsupportedModel", func(t *testing.T) {
		// input arguments for test case
		packageID := 1
		cpuModel := cpumodel.INTEL_FAM6_CORE2_MEROM

		// expected output
		expectedFreq := uint64(0)
		expectedErr := fmt.Errorf("base frequency metric not supported by CPU model: 0x%X", cpuModel)

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {
					packageID: 0,
				},
				1: {
					packageID: 0,
				},
				2: {
					packageID: packageID,
				},
			},
			model: cpuModel,
		}

		// msr definition
		m := &msrMock{}

		// power telemetry instance definition
		pt := PowerTelemetry{
			topology: topo,
			msr:      m,
			cpus:     []int{0, 1, 2},
		}

		actualFreq, err := pt.GetCPUBaseFrequency(packageID)
		require.Equal(t, expectedFreq, actualFreq)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("ErrorReadingMsr", func(t *testing.T) {
		// input arguments for test case
		packageID := 0
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_ATOM_SILVERMONT
		msrValue := uint64(0x1234)

		// expected output
		expectedFreq := uint64(0)
		expectedErr := errors.New("error reading msr")

		// msr definition
		m := &msrMock{}
		m.On("read", uint32(platformInfo), cpuID).Return(msrValue, errors.New("error reading msr"))

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {
					packageID: packageID,
				},
				1: {
					packageID: packageID,
				},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := PowerTelemetry{
			msr:      m,
			topology: topo,
			cpus:     []int{0, 1},
		}

		actualFreq, err := pt.GetCPUBaseFrequency(cpuID)
		require.ErrorContains(t, err, expectedErr.Error())
		require.Equal(t, expectedFreq, actualFreq)
		m.AssertExpectations(t)
	})
}

type msrGetOffsetDeltasResult struct {
	values map[uint32]uint64
	err    error
}

type msrGetTimestampDeltaResult struct {
	value time.Duration
	err   error
}

func TestGetCPUC0StateResidency(t *testing.T) {
	t.Run("MsrIsNil", func(t *testing.T) {
		cpuID := 1
		c0Exp := 0.0

		pt := &PowerTelemetry{}

		c0Out, err := pt.GetCPUC0StateResidency(cpuID)
		require.Equal(t, c0Exp, c0Out)
		require.ErrorContains(t, err, "\"msr\" is not initialized")
	})

	t.Run("InvalidCPUID", func(t *testing.T) {
		// input arguments for test case
		cpuID := 1
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: nil,
			err:    errors.New("CPU ID 1 not found"),
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("error retrieving offset deltas for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// power telemetry instance definition
		pt := &PowerTelemetry{
			msr: m,
		}

		out, err := pt.GetCPUC0StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("TSCDeltaZero", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				maxFreqClockCount: 200,
				timestampCounter:  0,
			},
			err: nil,
		}
		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("timestamp counter offset delta is zero for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// power telemetry instance definition
		pt := &PowerTelemetry{
			msr: m,
		}

		out, err := pt.GetCPUC0StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("MperfOffsetDeltaNotFound", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				timestampCounter: 5000,
			},
			err: nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("mperf offset delta not found for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// power telemetry instance definition
		pt := &PowerTelemetry{
			msr: m,
		}

		out, err := pt.GetCPUC0StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("TSCOffsetDeltaNotFound", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				maxFreqClockCount: 200,
			},
			err: nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("timestamp counter offset delta not found for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// power telemetry instance definition
		pt := &PowerTelemetry{
			msr: m,
		}

		out, err := pt.GetCPUC0StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("C0State2Per", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				maxFreqClockCount: 100000000,
				timestampCounter:  5000000000,
			},
			err: nil,
		}

		// expected output
		expectedResult := 2.0

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// power telemetry instance definition
		pt := &PowerTelemetry{
			msr: m,
		}

		out, err := pt.GetCPUC0StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.NoError(t, err)
		m.AssertExpectations(t)
	})
}

func TestGetCPUC1StateResidency(t *testing.T) {
	t.Run("MsrIsNil", func(t *testing.T) {
		cpuID := 1
		c1Exp := 0.0

		pt := PowerTelemetry{
			topology: &topologyData{},
		}

		c1Out, err := pt.GetCPUC1StateResidency(cpuID)
		require.Equal(t, c1Exp, c1Out)
		require.ErrorContains(t, err, "\"msr\" is not initialized")
	})

	t.Run("InvalidCPUID", func(t *testing.T) {
		// input arguments for test case
		cpuID := 1
		cpuModel := cpumodel.INTEL_FAM6_ATOM_SILVERMONT
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: nil,
			err:    errors.New("CPU ID 1 not found"),
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("error retrieving offset deltas for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC1StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("MperfOffsetDeltaNotFound", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_ATOM_SILVERMONT
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				timestampCounter: 5000,
			},
			err: nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("mperf offset delta not found for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC1StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("C3OffsetDeltaNotFound", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_ATOM_SILVERMONT
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				timestampCounter:  5000,
				maxFreqClockCount: 100,
			},
			err: nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("c3 state residency offset delta not found for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC1StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("C6OffsetDeltaNotFound", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_ATOM_SILVERMONT
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				timestampCounter:  5000,
				maxFreqClockCount: 100,
				c3Residency:       100,
			},
			err: nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("c6 state residency offset delta not found for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC1StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("C7OffsetDeltaNotFound", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_ATOM_SILVERMONT
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				timestampCounter:  5000,
				maxFreqClockCount: 100,
				c3Residency:       100,
				c6Residency:       100,
			},
			err: nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("c7 state residency offset delta not found for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC1StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("TSCOffsetDeltaNotFound", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_ATOM_SILVERMONT
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				maxFreqClockCount: 100,
				c3Residency:       100,
				c6Residency:       100,
				c7Residency:       100,
			},
			err: nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("timestamp counter offset delta not found for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC1StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("TSCDeltaZero", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_ATOM_SILVERMONT
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				timestampCounter:  0,
				maxFreqClockCount: 100,
				c3Residency:       100,
				c6Residency:       100,
				c7Residency:       100,
			},
			err: nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("timestamp counter offset delta is zero for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC1StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("C1Status20Per", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_ATOM_SILVERMONT
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				timestampCounter:  500,
				maxFreqClockCount: 100,
				c3Residency:       100,
				c6Residency:       100,
				c7Residency:       100,
			},
			err: nil,
		}

		// expected output
		expectedResult := 20.0

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC1StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.NoError(t, err)
		m.AssertExpectations(t)
	})

	t.Run("UnsupportedModel", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_CORE2_MEROM

		// expected output
		expectedResult := 0.0
		expectedErr := fmt.Errorf("c1 state residency metric not supported by CPU model: 0x%X", cpuModel)

		// msr definition
		m := &msrMock{}

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC1StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("InvalidCPUModel", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("c1 state residency metric not supported by CPU model: 0x0")

		// msr definition
		m := &msrMock{}

		// topology definition
		topo := &topologyData{}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC1StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})
}

func TestGetCPUC3StateResidency(t *testing.T) {
	t.Run("MsrIsNil", func(t *testing.T) {
		cpuID := 1
		c3Exp := 0.0

		pt := PowerTelemetry{
			topology: &topologyData{},
		}

		c3Out, err := pt.GetCPUC3StateResidency(cpuID)
		require.Equal(t, c3Exp, c3Out)
		require.ErrorContains(t, err, "\"msr\" is not initialized")
	})

	t.Run("InvalidCPUID", func(t *testing.T) {
		// input arguments for test case
		cpuID := 1
		cpuModel := cpumodel.INTEL_FAM6_HASWELL
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: nil,
			err:    errors.New("CPU ID 1 not found"),
		}

		// expected output
		expectedErr := errors.New("error retrieving offset deltas for CPU ID")
		expectedResult := 0.0

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC3StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("C3OffsetDeltaNotFound", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_HASWELL
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				timestampCounter: 5000,
			},
			err: nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("c3 state residency offset delta not found for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC3StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("C3State4Per", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_HASWELL
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				c3Residency:      200000000,
				timestampCounter: 5000000000,
			},
			err: nil,
		}

		// expected output
		expectedResult := 4.0

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC3StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.NoError(t, err)
		m.AssertExpectations(t)
	})

	t.Run("TSCOffsetDeltaNotFound", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_HASWELL
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				c3Residency: 200000000,
			},
			err: nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("timestamp counter offset delta not found for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC3StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("TSCDeltaZero", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_HASWELL
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				c3Residency:      200,
				timestampCounter: 0,
			},
			err: nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("timestamp counter offset delta is zero for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC3StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("UnsupportedModel", func(t *testing.T) {
		// input arguments for test case
		cpuModel := cpumodel.INTEL_FAM6_CORE2_MEROM
		cpuID := 0

		// expected output
		expectedResult := 0.0
		expectedErr := fmt.Errorf("c3 state residency metric not supported by CPU model: 0x%X", cpuModel)

		// msr definition
		m := &msrMock{}

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC3StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("InvalidCPUModel", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("c3 state residency metric not supported by CPU model: 0x0")

		// msr definition
		m := &msrMock{}

		// topology definition
		topo := &topologyData{}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC3StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})
}

func TestGetCPUC6StateResidency(t *testing.T) {
	t.Run("MsrIsNil", func(t *testing.T) {
		cpuID := 1
		c6Exp := 0.0

		pt := PowerTelemetry{
			topology: &topologyData{},
		}

		c6Out, err := pt.GetCPUC6StateResidency(cpuID)
		require.Equal(t, c6Exp, c6Out)
		require.ErrorContains(t, err, "\"msr\" is not initialized")
	})

	t.Run("InvalidCPUID", func(t *testing.T) {
		// input arguments for test case
		cpuID := 1
		cpuModel := cpumodel.INTEL_FAM6_ATOM_SILVERMONT
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: nil,
			err:    errors.New("CPU ID 1 not found"),
		}

		// expected output
		expectedErr := errors.New("error retrieving offset deltas for CPU ID")
		expectedResult := 0.0

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC6StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("C6OffsetDeltaNotFound", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_ATOM_SILVERMONT
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				timestampCounter: 5000,
			},
			err: nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("c6 state residency offset delta not found for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC6StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("C6State4Per", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_ATOM_SILVERMONT
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				c6Residency:      200000000,
				timestampCounter: 5000000000,
			},
			err: nil,
		}

		// expected output
		expectedResult := 4.0

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC6StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.NoError(t, err)
		m.AssertExpectations(t)
	})

	t.Run("TSCOffsetDeltaNotFound", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_ATOM_SILVERMONT
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				c6Residency: 200000000,
			},
			err: nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("timestamp counter offset delta not found for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC6StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("TSCDeltaZero", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_ATOM_SILVERMONT
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				c6Residency:      200,
				timestampCounter: 0,
			},
			err: nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("timestamp counter offset delta is zero for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC6StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("UnsupportedModel", func(t *testing.T) {
		// input arguments for test case
		cpuModel := cpumodel.INTEL_FAM6_CORE2_MEROM
		cpuID := 0

		// expected output
		expectedResult := 0.0
		expectedErr := fmt.Errorf("c6 state residency metric not supported by CPU model: 0x%X", cpuModel)

		// msr definition
		m := &msrMock{}

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC6StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("InvalidCPUModel", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("c6 state residency metric not supported by CPU model: 0x0")

		// msr definition
		m := &msrMock{}

		// topology definition
		topo := &topologyData{}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC6StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})
}

func TestGetCPUC7StateResidency(t *testing.T) {
	t.Run("MsrIsNil", func(t *testing.T) {
		cpuID := 1
		c7Exp := 0.0

		pt := PowerTelemetry{
			topology: &topologyData{},
		}

		c7Out, err := pt.GetCPUC7StateResidency(cpuID)
		require.Equal(t, c7Exp, c7Out)
		require.ErrorContains(t, err, "\"msr\" is not initialized")
	})

	t.Run("InvalidCPUID", func(t *testing.T) {
		// input arguments for test case
		cpuID := 1
		cpuModel := cpumodel.INTEL_FAM6_SKYLAKE
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: nil,
			err:    errors.New("CPU ID 1 not found"),
		}

		// expected output
		expectedErr := errors.New("error retrieving offset deltas for CPU ID")
		expectedResult := 0.0

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC7StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("C7OffsetDeltaNotFound", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_SKYLAKE
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				timestampCounter: 5000,
			},
			err: nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("c7 state residency offset delta not found for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC7StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("C7State4Per", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_SKYLAKE
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				c7Residency:      200000000,
				timestampCounter: 5000000000,
			},
			err: nil,
		}

		// expected output
		expectedResult := 4.0

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC7StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.NoError(t, err)
		m.AssertExpectations(t)
	})

	t.Run("TSCOffsetDeltaNotFound", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_SKYLAKE
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				c7Residency: 200000000,
			},
			err: nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("timestamp counter offset delta not found for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC7StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("TSCDeltaZero", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_SKYLAKE
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				c7Residency:      200,
				timestampCounter: 0,
			},
			err: nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("timestamp counter offset delta is zero for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC7StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("UnsupportedModel", func(t *testing.T) {
		// input arguments for test case
		cpuModel := cpumodel.INTEL_FAM6_CORE2_MEROM
		cpuID := 0

		// expected output
		expectedResult := 0.0
		expectedErr := fmt.Errorf("c7 state residency metric not supported by CPU model: 0x%X", cpuModel)

		// msr definition
		m := &msrMock{}

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC7StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("InvalidCPUModel", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("c7 state residency metric not supported by CPU model")

		// msr definition
		m := &msrMock{}

		// topology definition
		topo := &topologyData{}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUC7StateResidency(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})
}

func TestGetCPUBusyFrequencyMhz(t *testing.T) {
	t.Run("MsrIsNil", func(t *testing.T) {
		cpuID := 1
		busyFreqExp := 0.0

		pt := &PowerTelemetry{}

		busyFreqOut, err := pt.GetCPUBusyFrequencyMhz(cpuID)
		require.Equal(t, busyFreqExp, busyFreqOut)
		require.ErrorContains(t, err, "\"msr\" is not initialized")
	})

	t.Run("DeltasInvalidCPUID", func(t *testing.T) {
		// input arguments for test case
		cpuID := 1
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: nil,
			err:    errors.New("CPU ID 1 not found"),
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("error retrieving offset deltas for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()

		// power telemetry instance definition
		pt := &PowerTelemetry{
			msr: m,
		}

		out, err := pt.GetCPUBusyFrequencyMhz(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("TimestampDeltaInvalidCPUID", func(t *testing.T) {
		// input arguments for test case
		cpuID := 1
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				actualFreqClockCount: 100000000000,
				maxFreqClockCount:    200000000,
				timestampCounter:     1000000,
			},
			err: nil,
		}
		mockTimestampDelta := &msrGetTimestampDeltaResult{
			value: time.Duration(0),
			err:   errors.New("CPU ID 1 not found"),
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("error retrieving timestamp delta for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()
		m.On("getTimestampDelta", cpuID).Return(mockTimestampDelta.value, mockTimestampDelta.err).Once()

		// power telemetry instance definition
		pt := &PowerTelemetry{
			msr: m,
		}

		out, err := pt.GetCPUBusyFrequencyMhz(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("TimestampDeltaInvalidCPUID", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				actualFreqClockCount: 100000000000,
				maxFreqClockCount:    200000000,
				timestampCounter:     1000000,
			},
			err: nil,
		}
		mockTimestampDelta := &msrGetTimestampDeltaResult{
			value: time.Duration(0),
			err:   errors.New("CPU ID 1 not found"),
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("error retrieving timestamp delta for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()
		m.On("getTimestampDelta", cpuID).Return(mockTimestampDelta.value, mockTimestampDelta.err).Once()

		// power telemetry instance definition
		pt := &PowerTelemetry{
			msr: m,
		}

		out, err := pt.GetCPUBusyFrequencyMhz(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("InvalidTimeInterval", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				actualFreqClockCount: 100000000000,
				maxFreqClockCount:    200000000,
				timestampCounter:     1000000,
			},
			err: nil,
		}
		mockTimestampDelta := &msrGetTimestampDeltaResult{
			value: -100,
			err:   nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("timestamp delta must be greater than zero")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()
		m.On("getTimestampDelta", cpuID).Return(mockTimestampDelta.value, mockTimestampDelta.err).Once()

		// power telemetry instance definition
		pt := &PowerTelemetry{
			msr: m,
		}

		out, err := pt.GetCPUBusyFrequencyMhz(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("MperfOffsetDeltaNotFound", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				timestampCounter: 5000,
			},
			err: nil,
		}
		mockTimestampDelta := &msrGetTimestampDeltaResult{
			value: 1,
			err:   nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("mperf offset delta not found for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()
		m.On("getTimestampDelta", cpuID).Return(mockTimestampDelta.value, mockTimestampDelta.err).Once()

		// power telemetry instance definition
		pt := &PowerTelemetry{
			msr: m,
		}

		out, err := pt.GetCPUBusyFrequencyMhz(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("MperfOffsetDeltaZero", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				timestampCounter:  5000,
				maxFreqClockCount: 0,
			},
			err: nil,
		}
		mockTimestampDelta := &msrGetTimestampDeltaResult{
			value: 1,
			err:   nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("mperf offset delta is zero for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()
		m.On("getTimestampDelta", cpuID).Return(mockTimestampDelta.value, mockTimestampDelta.err).Once()

		// power telemetry instance definition
		pt := &PowerTelemetry{
			msr: m,
		}

		out, err := pt.GetCPUBusyFrequencyMhz(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("AperfOffsetDeltaNotFound", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				timestampCounter:  5000,
				maxFreqClockCount: 100,
			},
			err: nil,
		}
		mockTimestampDelta := &msrGetTimestampDeltaResult{
			value: 1,
			err:   nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("aperf offset delta not found for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()
		m.On("getTimestampDelta", cpuID).Return(mockTimestampDelta.value, mockTimestampDelta.err).Once()

		// power telemetry instance definition
		pt := &PowerTelemetry{
			msr: m,
		}

		out, err := pt.GetCPUBusyFrequencyMhz(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("TSCOffsetDeltaNotFound", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				actualFreqClockCount: 100000000000,
				maxFreqClockCount:    200000000,
			},
			err: nil,
		}
		mockTimestampDelta := &msrGetTimestampDeltaResult{
			value: 1,
			err:   nil,
		}

		// expected output
		expectedResult := 0.0
		expectedErr := errors.New("timestamp counter offset delta not found for CPU ID")

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()
		m.On("getTimestampDelta", cpuID).Return(mockTimestampDelta.value, mockTimestampDelta.err).Once()

		// power telemetry instance definition
		pt := &PowerTelemetry{
			msr: m,
		}

		out, err := pt.GetCPUBusyFrequencyMhz(cpuID)
		require.Equal(t, expectedResult, out)
		require.ErrorContains(t, err, expectedErr.Error())
		m.AssertExpectations(t)
	})

	t.Run("BusyFreq", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_ATOM_SILVERMONT
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				actualFreqClockCount: 100000000000,
				maxFreqClockCount:    200000000,
				timestampCounter:     1000000,
			},
			err: nil,
		}
		mockTimestampDelta := &msrGetTimestampDeltaResult{
			value: 1000000000,
			err:   nil,
		}

		// expected output
		expectedResult := 500.0

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()
		m.On("getTimestampDelta", cpuID).Return(mockTimestampDelta.value, mockTimestampDelta.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUBusyFrequencyMhz(cpuID)
		require.Equal(t, expectedResult, out)
		require.NoError(t, err)
		m.AssertExpectations(t)
	})

	t.Run("UnsupportedModel", func(t *testing.T) {
		// input arguments for test case
		cpuID := 0
		cpuModel := cpumodel.INTEL_FAM6_CORE2_MEROM
		mockOffsetDeltas := &msrGetOffsetDeltasResult{
			values: map[uint32]uint64{
				actualFreqClockCount: 100000000000,
				maxFreqClockCount:    200000000,
				timestampCounter:     1000000,
			},
			err: nil,
		}
		mockTimestampDelta := &msrGetTimestampDeltaResult{
			value: 1000000000,
			err:   nil,
		}

		// expected output
		expectedResult := 500.0

		// msr definition
		m := &msrMock{}
		m.On("getOffsetDeltas", cpuID).Return(mockOffsetDeltas.values, mockOffsetDeltas.err).Once()
		m.On("getTimestampDelta", cpuID).Return(mockTimestampDelta.value, mockTimestampDelta.err).Once()

		// topology definition
		topo := &topologyData{
			topologyMap: map[int]*cpuInfo{
				0: {},
			},
			model: cpuModel,
		}

		// power telemetry instance definition
		pt := &PowerTelemetry{
			topology: topo,
			msr:      m,
		}

		out, err := pt.GetCPUBusyFrequencyMhz(cpuID)
		require.Equal(t, expectedResult, out)
		require.NoError(t, err)
		m.AssertExpectations(t)
	})
}

func TestUpdatePerCPUMetrics(t *testing.T) {
	t.Run("MsrIsNil", func(t *testing.T) {
		cpuID := 0

		pt := &PowerTelemetry{}

		err := pt.UpdatePerCPUMetrics(cpuID)
		require.ErrorContains(t, err, "\"msr\" is not initialized")
	})

	t.Run("FailedToUpdate", func(t *testing.T) {
		cpuID := 0
		errExpected := errors.New("error while updating storage")
		m := &msrMock{}
		m.On("update", cpuID).Return(errExpected).Once()

		pt := &PowerTelemetry{
			msr: m,
		}

		err := pt.UpdatePerCPUMetrics(cpuID)
		require.ErrorContains(t, err, errExpected.Error())
		m.AssertExpectations(t)
	})

	t.Run("Updated", func(t *testing.T) {
		cpuID := 0
		m := &msrMock{}
		m.On("update", cpuID).Return(nil).Once()

		pt := &PowerTelemetry{
			msr: m,
		}

		err := pt.UpdatePerCPUMetrics(cpuID)
		require.NoError(t, err)
		m.AssertExpectations(t)
	})
}

func TestIsCPUSupported(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		mTopology := &topologyMock{}
		mTopology.On("getCPUFamily", 0).Return("6", nil).Once()
		mTopology.On("getCPUVendor", 0).Return("GenuineIntel", nil).Once()

		isSupported, err := isCPUSupported(mTopology)
		require.True(t, isSupported)
		require.NoError(t, err)
		mTopology.AssertExpectations(t)
	})

	t.Run("False", func(t *testing.T) {
		t.Run("FamilyNotIntel6", func(t *testing.T) {
			mTopology := &topologyMock{}
			mTopology.On("getCPUFamily", 0).Return("", nil).Once()
			mTopology.On("getCPUVendor", 0).Return("GenuineIntel", nil).Once()

			isSupported, err := isCPUSupported(mTopology)
			require.False(t, isSupported)
			require.NoError(t, err)
			mTopology.AssertExpectations(t)
		})

		t.Run("VendorNotGenuineIntel", func(t *testing.T) {
			mTopology := &topologyMock{}
			mTopology.On("getCPUFamily", 0).Return("6", nil).Once()
			mTopology.On("getCPUVendor", 0).Return("AuthenticAMD", nil).Once()

			isSupported, err := isCPUSupported(mTopology)
			require.False(t, isSupported)
			require.NoError(t, err)
			mTopology.AssertExpectations(t)
		})

		t.Run("FailedToGetFamily", func(t *testing.T) {
			mError := errors.New("mock error")
			mTopology := &topologyMock{}
			mTopology.On("getCPUFamily", 0).Return("", mError).Once()

			isSupported, err := isCPUSupported(mTopology)
			require.False(t, isSupported)
			require.ErrorContains(t, err, mError.Error())
			mTopology.AssertExpectations(t)
		})

		t.Run("FailedToGetVendor", func(t *testing.T) {
			mError := errors.New("mock error")
			mTopology := &topologyMock{}
			mTopology.On("getCPUFamily", 0).Return("6", nil).Once()
			mTopology.On("getCPUVendor", 0).Return("", mError).Once()

			isSupported, err := isCPUSupported(mTopology)
			require.False(t, isSupported)
			require.ErrorContains(t, err, mError.Error())
			mTopology.AssertExpectations(t)
		})
	})
}

func TestIsFlagSupported(t *testing.T) {
	tests := []struct {
		name     string
		topology topologyReader
		err      error
		expected bool
	}{
		{
			name:     "InvalidCPUID",
			topology: &topologyData{},
			err:      errors.New("error retrieving CPU flags"),
			expected: false,
		},
		{
			name: "FlagNotSupported",
			topology: &topologyData{
				topologyMap: map[int]*cpuInfo{
					0: {
						flags: []string{"flag2"},
					},
				},
			},
			err:      nil,
			expected: false,
		},
		{
			name: "FlagSupported",
			topology: &topologyData{
				topologyMap: map[int]*cpuInfo{
					0: {
						flags: []string{"flag"},
					},
				},
			},
			err:      nil,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// power telemetry instance definition
			pt := &PowerTelemetry{
				topology: tt.topology,
			}
			// actual output
			actual, err := pt.IsFlagSupported("flag")
			require.Equal(t, tt.expected, actual)
			if tt.err != nil {
				require.ErrorContains(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetCPUPackageID(t *testing.T) {
	tests := []struct {
		name     string
		topology topologyReader
		err      error
		expected int
	}{
		{
			name:     "InvalidPackageIDNoCPUs",
			topology: &topologyData{},
			err:      errors.New("error retrieving package ID"),
			expected: 0,
		},
		{
			name: "InvalidPackageIDCPUDoesNotExist",
			topology: &topologyData{
				topologyMap: map[int]*cpuInfo{
					1: {
						packageID: 1,
					},
				},
			},
			err:      errors.New("error retrieving package ID"),
			expected: 0,
		},
		{
			name: "ValidPackageID",
			topology: &topologyData{
				topologyMap: map[int]*cpuInfo{
					0: {
						packageID: 1,
					},
				},
			},
			err:      nil,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// power telemetry instance definition
			pt := &PowerTelemetry{
				topology: tt.topology,
			}
			// actual output
			actual, err := pt.GetCPUPackageID(0)
			require.Equal(t, tt.expected, actual)
			if tt.err != nil {
				require.ErrorContains(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetCPUCoreID(t *testing.T) {
	tests := []struct {
		name     string
		topology topologyReader
		err      error
		expected int
	}{
		{
			name:     "InvalidCoreIDNoCPUs",
			topology: &topologyData{},
			err:      errors.New("error retrieving core ID"),
			expected: 0,
		},
		{
			name: "InvalidCoreIDCPUDoesNotExist",
			topology: &topologyData{
				topologyMap: map[int]*cpuInfo{
					1: {
						coreID: 1,
					},
				},
			},
			err:      errors.New("error retrieving core ID"),
			expected: 0,
		},
		{
			name: "ValidCoreID",
			topology: &topologyData{
				topologyMap: map[int]*cpuInfo{
					0: {
						coreID: 1,
					},
				},
			},
			err:      nil,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// power telemetry instance definition
			pt := &PowerTelemetry{
				topology: tt.topology,
			}
			// actual output
			actual, err := pt.GetCPUCoreID(0)
			require.Equal(t, tt.expected, actual)
			if tt.err != nil {
				require.ErrorContains(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetPackageDieIDs(t *testing.T) {
	tests := []struct {
		name     string
		topology topologyReader
		err      error
		expected []int
	}{
		{
			name:     "InvalidPackageIDNoDies",
			topology: &topologyData{},
			err:      errors.New("error retrieving dies"),
			expected: nil,
		},
		{
			name: "InvalidPackageIDPackageDoesNotExist",
			topology: &topologyData{
				packageDies: map[int][]int{
					1: {
						0, 1,
					},
				},
			},
			err:      errors.New("error retrieving dies"),
			expected: nil,
		},
		{
			name: "ValidPackageID",
			topology: &topologyData{
				packageDies: map[int][]int{
					0: {
						1, 2,
					},
				},
			},
			err:      nil,
			expected: []int{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// power telemetry instance definition
			pt := &PowerTelemetry{
				topology: tt.topology,
			}
			// actual output
			actual, err := pt.GetPackageDieIDs(0)
			require.Equal(t, tt.expected, actual)
			if tt.err != nil {
				require.ErrorContains(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type raplMock struct {
	mock.Mock
}

func (m *raplMock) initZoneMap() error {
	args := m.Called()
	return args.Error(0)
}

func (m *raplMock) getPackageIDs() []int {
	args := m.Called()
	return args.Get(0).([]int)
}

func (m *raplMock) isRaplLoaded(modulesPath string) (bool, error) {
	args := m.Called(modulesPath)
	return args.Bool(0), args.Error(1)
}

func (m *raplMock) getCurrentPowerConsumptionWatts(packageID int, domain string) (float64, error) {
	args := m.Called(packageID, domain)
	return args.Get(0).(float64), args.Error(1)
}

func (m *raplMock) getMaxPowerConstraintWatts(packageID int) (float64, error) {
	args := m.Called(packageID)
	return args.Get(0).(float64), args.Error(1)
}

func TestPower_GetCurrentPackagePowerConsumptionWatts(t *testing.T) {
	t.Run("RaplIsNil", func(t *testing.T) {
		packageID := 0
		currPowerExp := 0.0

		pt := &PowerTelemetry{}
		currPowerOut, err := pt.GetCurrentPackagePowerConsumptionWatts(packageID)
		require.Equal(t, currPowerExp, currPowerOut)
		require.ErrorContains(t, err, "\"rapl\" is not initialized")
	})

	t.Run("FailedToGetCurrentPower", func(t *testing.T) {
		packageID := 0
		currPowerExp := 0.0

		mError := errors.New("mock error")
		mRapl := &raplMock{}
		mRapl.On("getCurrentPowerConsumptionWatts", packageID, packageDomain.String()).Return(currPowerExp, mError).Once()

		pt := &PowerTelemetry{
			rapl: mRapl,
		}

		currPowerOut, err := pt.GetCurrentPackagePowerConsumptionWatts(packageID)
		require.Equal(t, currPowerExp, currPowerOut)
		require.ErrorContains(t, err, mError.Error())
		mRapl.AssertExpectations(t)
	})

	t.Run("Ok", func(t *testing.T) {
		packageID := 0
		currPowerExp := 30.0

		mRapl := &raplMock{}
		mRapl.On("getCurrentPowerConsumptionWatts", packageID, packageDomain.String()).Return(currPowerExp, nil).Once()

		pt := &PowerTelemetry{
			rapl: mRapl,
		}

		currPowerOut, err := pt.GetCurrentPackagePowerConsumptionWatts(packageID)
		require.Equal(t, currPowerExp, currPowerOut)
		require.NoError(t, err)
		mRapl.AssertExpectations(t)
	})
}

func TestPower_GetCurrentDramPowerConsumptionWatts(t *testing.T) {
	t.Run("RaplIsNil", func(t *testing.T) {
		packageID := 0
		currPowerExp := 0.0

		pt := &PowerTelemetry{}
		currPowerOut, err := pt.GetCurrentDramPowerConsumptionWatts(packageID)
		require.Equal(t, currPowerExp, currPowerOut)
		require.ErrorContains(t, err, "\"rapl\" is not initialized")
	})

	t.Run("FailedToGetCurrentPower", func(t *testing.T) {
		packageID := 0
		currPowerExp := 0.0

		mError := errors.New("mock error")
		mRapl := &raplMock{}
		mRapl.On("getCurrentPowerConsumptionWatts", packageID, dramDomain.String()).Return(currPowerExp, mError).Once()

		pt := &PowerTelemetry{
			rapl: mRapl,
		}

		currPowerOut, err := pt.GetCurrentDramPowerConsumptionWatts(packageID)
		require.Equal(t, currPowerExp, currPowerOut)
		require.ErrorContains(t, err, mError.Error())
		mRapl.AssertExpectations(t)
	})

	t.Run("Ok", func(t *testing.T) {
		packageID := 0
		currPowerExp := 30.0

		mRapl := &raplMock{}
		mRapl.On("getCurrentPowerConsumptionWatts", packageID, dramDomain.String()).Return(currPowerExp, nil).Once()

		pt := &PowerTelemetry{
			rapl: mRapl,
		}

		currPowerOut, err := pt.GetCurrentDramPowerConsumptionWatts(packageID)
		require.Equal(t, currPowerExp, currPowerOut)
		require.NoError(t, err)
		mRapl.AssertExpectations(t)
	})
}

func TestPower_GetPackageThermalDesignPowerWatts(t *testing.T) {
	t.Run("RaplIsNil", func(t *testing.T) {
		packageID := 0
		currPowerExp := 0.0

		pt := &PowerTelemetry{}
		currPowerOut, err := pt.GetPackageThermalDesignPowerWatts(packageID)
		require.Equal(t, currPowerExp, currPowerOut)
		require.ErrorContains(t, err, "\"rapl\" is not initialized")
	})

	t.Run("FailedToGetMaxPower", func(t *testing.T) {
		packageID := 0
		maxPowerExp := 0.0

		mError := errors.New("mock error")
		mRapl := &raplMock{}
		mRapl.On("getMaxPowerConstraintWatts", packageID).Return(maxPowerExp, mError).Once()

		pt := &PowerTelemetry{
			rapl: mRapl,
		}

		maxPowerOut, err := pt.GetPackageThermalDesignPowerWatts(packageID)
		require.Equal(t, maxPowerExp, maxPowerOut)
		require.ErrorContains(t, err, mError.Error())
		mRapl.AssertExpectations(t)
	})

	t.Run("Ok", func(t *testing.T) {
		packageID := 0
		maxPowerExp := 30.0

		mRapl := &raplMock{}
		mRapl.On("getMaxPowerConstraintWatts", packageID).Return(maxPowerExp, nil).Once()

		pt := &PowerTelemetry{
			rapl: mRapl,
		}

		maxPowerOut, err := pt.GetPackageThermalDesignPowerWatts(packageID)
		require.Equal(t, maxPowerExp, maxPowerOut)
		require.NoError(t, err)
		mRapl.AssertExpectations(t)
	})
}

type perfMock struct {
	mock.Mock
}

func (m *perfMock) activate(events []string, cores []int) error {
	args := m.Called(events, cores)
	return args.Error(0)
}

func (m *perfMock) read() ([]coreMetric, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]coreMetric), args.Error(1)
}

func (m *perfMock) deactivate() error {
	args := m.Called()
	return args.Error(0)
}

func (m *perfMock) initResolver(jsonFile string) error {
	args := m.Called(jsonFile)
	return args.Error(0)
}

func (m *perfMock) update() error {
	args := m.Called()
	return args.Error(0)
}

func (m *perfMock) getCoreMetrics(cpuID int) []coreMetric {
	args := m.Called(cpuID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]coreMetric)
}

func TestPower_GetCPUC0SubstateC01Percent(t *testing.T) {
	t.Run("PerfIsNil", func(t *testing.T) {
		cpuID := 0
		c01Exp := 0.0

		pt := &PowerTelemetry{}

		c01Out, err := pt.GetCPUC0SubstateC01Percent(cpuID)
		require.ErrorContains(t, err, "\"perf\" is not initialized")
		require.Equal(t, c01Exp, c01Out)
	})

	t.Run("GetCoreMetricsError", func(t *testing.T) {
		cpuID := 0
		c01Exp := 0.0

		mPerf := &perfMock{}
		mMetrics := []coreMetric{}
		mPerf.On("getCoreMetrics", cpuID).Return(mMetrics).Once()

		pt := &PowerTelemetry{
			perf: mPerf,
		}

		c01Out, err := pt.GetCPUC0SubstateC01Percent(cpuID)
		require.ErrorContains(t, err, fmt.Sprintf("no core metrics found for CPU ID: %v", cpuID))
		require.Equal(t, c01Exp, c01Out)
		mPerf.AssertExpectations(t)
	})

	t.Run("C01MetricNotFound", func(t *testing.T) {
		cpuID := 0
		c01Exp := 0.0

		mPerf := &perfMock{}
		mMetrics := []coreMetric{
			{
				name:   c02.String(),
				cpuID:  cpuID,
				scaled: 100,
			},
			{
				name:   thread.String(),
				cpuID:  cpuID,
				scaled: 5000,
			},
		}
		mPerf.On("getCoreMetrics", cpuID).Return(mMetrics).Once()

		pt := &PowerTelemetry{
			perf: mPerf,
		}

		c01Out, err := pt.GetCPUC0SubstateC01Percent(cpuID)
		require.ErrorContains(t, err, fmt.Sprintf("could not find metric: %q", c01.String()))
		require.Equal(t, c01Exp, c01Out)
		mPerf.AssertExpectations(t)
	})

	t.Run("ThreadMetricNotFound", func(t *testing.T) {
		cpuID := 0
		c01Exp := 0.0

		mPerf := &perfMock{}
		mMetrics := []coreMetric{
			{
				name:   c01.String(),
				cpuID:  cpuID,
				scaled: 100,
			},
		}
		mPerf.On("getCoreMetrics", cpuID).Return(mMetrics).Once()

		pt := &PowerTelemetry{
			perf: mPerf,
		}

		c01Out, err := pt.GetCPUC0SubstateC01Percent(cpuID)
		require.ErrorContains(t, err, fmt.Sprintf("could not find metric: %q", thread.String()))
		require.Equal(t, c01Exp, c01Out)
		mPerf.AssertExpectations(t)
	})

	t.Run("ThreadMetricIsZero", func(t *testing.T) {
		cpuID := 0
		c01Exp := 0.0

		mPerf := &perfMock{}
		mMetrics := []coreMetric{
			{
				name:   c01.String(),
				cpuID:  cpuID,
				scaled: 100,
			},
			{
				name:   thread.String(),
				cpuID:  cpuID,
				scaled: 0,
			},
		}
		mPerf.On("getCoreMetrics", cpuID).Return(mMetrics).Once()

		pt := &PowerTelemetry{
			perf: mPerf,
		}

		c01Out, err := pt.GetCPUC0SubstateC01Percent(cpuID)
		require.ErrorContains(t, err, fmt.Sprintf("zero scaled value for reference metric: %q", thread.String()))
		require.Equal(t, c01Exp, c01Out)
		mPerf.AssertExpectations(t)
	})

	t.Run("C01State2Per", func(t *testing.T) {
		cpuID := 0
		c01Exp := 2.0

		mPerf := &perfMock{}
		mMetrics := []coreMetric{
			{
				name:   c01.String(),
				cpuID:  cpuID,
				scaled: 100,
			},
			{
				name:   thread.String(),
				cpuID:  cpuID,
				scaled: 5000,
			},
		}
		mPerf.On("getCoreMetrics", cpuID).Return(mMetrics).Once()

		pt := &PowerTelemetry{
			perf: mPerf,
		}

		c01Out, err := pt.GetCPUC0SubstateC01Percent(cpuID)
		require.NoError(t, err)
		require.Equal(t, c01Exp, c01Out)
		mPerf.AssertExpectations(t)
	})
}

func TestPower_GetCPUC0SubstateC02Percent(t *testing.T) {
	t.Run("PerfIsNil", func(t *testing.T) {
		cpuID := 0
		c02Exp := 0.0

		pt := &PowerTelemetry{}

		c02Out, err := pt.GetCPUC0SubstateC02Percent(cpuID)
		require.ErrorContains(t, err, "\"perf\" is not initialized")
		require.Equal(t, c02Exp, c02Out)
	})

	t.Run("GetCoreMetricsError", func(t *testing.T) {
		cpuID := 0
		c02Exp := 0.0

		mPerf := &perfMock{}
		mMetrics := []coreMetric{}
		mPerf.On("getCoreMetrics", cpuID).Return(mMetrics).Once()

		pt := &PowerTelemetry{
			perf: mPerf,
		}

		c02Out, err := pt.GetCPUC0SubstateC02Percent(cpuID)
		require.ErrorContains(t, err, fmt.Sprintf("no core metrics found for CPU ID: %v", cpuID))
		require.Equal(t, c02Exp, c02Out)
		mPerf.AssertExpectations(t)
	})

	t.Run("C02MetricNotFound", func(t *testing.T) {
		cpuID := 0
		c02Exp := 0.0

		mPerf := &perfMock{}
		mMetrics := []coreMetric{
			{
				name:   c01.String(),
				cpuID:  cpuID,
				scaled: 100,
			},
			{
				name:   thread.String(),
				cpuID:  cpuID,
				scaled: 5000,
			},
		}
		mPerf.On("getCoreMetrics", cpuID).Return(mMetrics).Once()

		pt := &PowerTelemetry{
			perf: mPerf,
		}

		c02Out, err := pt.GetCPUC0SubstateC02Percent(cpuID)
		require.ErrorContains(t, err, fmt.Sprintf("could not find metric: %q", c02.String()))
		require.Equal(t, c02Exp, c02Out)
		mPerf.AssertExpectations(t)
	})

	t.Run("ThreadMetricNotFound", func(t *testing.T) {
		cpuID := 0
		c02Exp := 0.0

		mPerf := &perfMock{}
		mMetrics := []coreMetric{
			{
				name:   c02.String(),
				cpuID:  cpuID,
				scaled: 100,
			},
		}
		mPerf.On("getCoreMetrics", cpuID).Return(mMetrics).Once()

		pt := &PowerTelemetry{
			perf: mPerf,
		}

		c02Out, err := pt.GetCPUC0SubstateC02Percent(cpuID)
		require.ErrorContains(t, err, fmt.Sprintf("could not find metric: %q", thread.String()))
		require.Equal(t, c02Exp, c02Out)
		mPerf.AssertExpectations(t)
	})

	t.Run("ThreadMetricIsZero", func(t *testing.T) {
		cpuID := 0
		c02Exp := 0.0

		mPerf := &perfMock{}
		mMetrics := []coreMetric{
			{
				name:   c02.String(),
				cpuID:  cpuID,
				scaled: 100,
			},
			{
				name:   thread.String(),
				cpuID:  cpuID,
				scaled: 0,
			},
		}
		mPerf.On("getCoreMetrics", cpuID).Return(mMetrics).Once()

		pt := &PowerTelemetry{
			perf: mPerf,
		}

		c02Out, err := pt.GetCPUC0SubstateC02Percent(cpuID)
		require.ErrorContains(t, err, fmt.Sprintf("zero scaled value for reference metric: %q", thread.String()))
		require.Equal(t, c02Exp, c02Out)
		mPerf.AssertExpectations(t)
	})

	t.Run("C02State4Per", func(t *testing.T) {
		cpuID := 0
		c02Exp := 4.0

		mPerf := &perfMock{}
		mMetrics := []coreMetric{
			{
				name:   c02.String(),
				cpuID:  cpuID,
				scaled: 200,
			},
			{
				name:   thread.String(),
				cpuID:  cpuID,
				scaled: 5000,
			},
		}
		mPerf.On("getCoreMetrics", cpuID).Return(mMetrics).Once()

		pt := &PowerTelemetry{
			perf: mPerf,
		}

		c02Out, err := pt.GetCPUC0SubstateC02Percent(cpuID)
		require.NoError(t, err)
		require.Equal(t, c02Exp, c02Out)
		mPerf.AssertExpectations(t)
	})
}

func TestPower_GetCPUC0SubstateC0WaitPercent(t *testing.T) {
	t.Run("PerfIsNil", func(t *testing.T) {
		cpuID := 0
		c0WaitExp := 0.0

		pt := &PowerTelemetry{}

		c0WaitOut, err := pt.GetCPUC0SubstateC0WaitPercent(cpuID)
		require.ErrorContains(t, err, "\"perf\" is not initialized")
		require.Equal(t, c0WaitExp, c0WaitOut)
	})

	t.Run("GetCoreMetricsError", func(t *testing.T) {
		cpuID := 0
		c0WaitExp := 0.0

		mPerf := &perfMock{}
		mMetrics := []coreMetric{}
		mPerf.On("getCoreMetrics", cpuID).Return(mMetrics).Once()

		pt := &PowerTelemetry{
			perf: mPerf,
		}

		c0WaitOut, err := pt.GetCPUC0SubstateC0WaitPercent(cpuID)
		require.ErrorContains(t, err, fmt.Sprintf("no core metrics found for CPU ID: %v", cpuID))
		require.Equal(t, c0WaitExp, c0WaitOut)
		mPerf.AssertExpectations(t)
	})

	t.Run("C0WaitMetricNotFound", func(t *testing.T) {
		cpuID := 0
		c0WaitExp := 0.0

		mPerf := &perfMock{}
		mMetrics := []coreMetric{
			{
				name:   c01.String(),
				cpuID:  cpuID,
				scaled: 100,
			},
			{
				name:   thread.String(),
				cpuID:  cpuID,
				scaled: 5000,
			},
		}
		mPerf.On("getCoreMetrics", cpuID).Return(mMetrics).Once()

		pt := &PowerTelemetry{
			perf: mPerf,
		}

		c0WaitOut, err := pt.GetCPUC0SubstateC0WaitPercent(cpuID)
		require.ErrorContains(t, err, fmt.Sprintf("could not find metric: %q", c0Wait.String()))
		require.Equal(t, c0WaitExp, c0WaitOut)
		mPerf.AssertExpectations(t)
	})

	t.Run("ThreadMetricNotFound", func(t *testing.T) {
		cpuID := 0
		c0WaitExp := 0.0

		mPerf := &perfMock{}
		mMetrics := []coreMetric{
			{
				name:   c0Wait.String(),
				cpuID:  cpuID,
				scaled: 100,
			},
		}
		mPerf.On("getCoreMetrics", cpuID).Return(mMetrics).Once()

		pt := &PowerTelemetry{
			perf: mPerf,
		}

		c0WaitOut, err := pt.GetCPUC0SubstateC0WaitPercent(cpuID)
		require.ErrorContains(t, err, fmt.Sprintf("could not find metric: %q", thread.String()))
		require.Equal(t, c0WaitExp, c0WaitOut)
		mPerf.AssertExpectations(t)
	})

	t.Run("ThreadMetricIsZero", func(t *testing.T) {
		cpuID := 0
		c0WaitExp := 0.0

		mPerf := &perfMock{}
		mMetrics := []coreMetric{
			{
				name:   c0Wait.String(),
				cpuID:  cpuID,
				scaled: 100,
			},
			{
				name:   thread.String(),
				cpuID:  cpuID,
				scaled: 0,
			},
		}
		mPerf.On("getCoreMetrics", cpuID).Return(mMetrics).Once()

		pt := &PowerTelemetry{
			perf: mPerf,
		}

		c0WaitOut, err := pt.GetCPUC0SubstateC0WaitPercent(cpuID)
		require.ErrorContains(t, err, fmt.Sprintf("zero scaled value for reference metric: %q", thread.String()))
		require.Equal(t, c0WaitExp, c0WaitOut)
		mPerf.AssertExpectations(t)
	})

	t.Run("C0WaitState10Per", func(t *testing.T) {
		cpuID := 0
		c0WaitExp := 10.0

		mPerf := &perfMock{}
		mMetrics := []coreMetric{
			{
				name:   c0Wait.String(),
				cpuID:  cpuID,
				scaled: 500,
			},
			{
				name:   thread.String(),
				cpuID:  cpuID,
				scaled: 5000,
			},
		}
		mPerf.On("getCoreMetrics", cpuID).Return(mMetrics).Once()

		pt := &PowerTelemetry{
			perf: mPerf,
		}

		c0WaitOut, err := pt.GetCPUC0SubstateC0WaitPercent(cpuID)
		require.NoError(t, err)
		require.Equal(t, c0WaitExp, c0WaitOut)
		mPerf.AssertExpectations(t)
	})
}

func TestPower_ReadPerfEvents(t *testing.T) {
	t.Run("PerfIsNil", func(t *testing.T) {
		pt := &PowerTelemetry{}

		err := pt.ReadPerfEvents()
		require.ErrorContains(t, err, "\"perf\" is not initialized")
	})

	t.Run("FailedToRead", func(t *testing.T) {
		mError := errors.New("error while reading perf events")
		mPerf := &perfMock{}
		mPerf.On("update").Return(mError).Once()

		pt := &PowerTelemetry{
			perf: mPerf,
		}

		err := pt.ReadPerfEvents()
		require.ErrorContains(t, err, mError.Error())
		mPerf.AssertExpectations(t)
	})

	t.Run("SuccessfulRead", func(t *testing.T) {
		mPerf := &perfMock{}
		mPerf.On("update").Return(nil).Once()

		pt := &PowerTelemetry{
			perf: mPerf,
		}

		err := pt.ReadPerfEvents()
		require.NoError(t, err)
		mPerf.AssertExpectations(t)
	})
}

func TestPower_DeactivatePerfEvents(t *testing.T) {
	t.Run("PerfIsNil", func(t *testing.T) {
		pt := &PowerTelemetry{}

		err := pt.DeactivatePerfEvents()
		require.ErrorContains(t, err, "\"perf\" is not initialized")
	})

	t.Run("FailedToDeactivate", func(t *testing.T) {
		mError := errors.New("error while reading perf events")
		mPerf := &perfMock{}
		mPerf.On("deactivate").Return(mError).Once()

		pt := &PowerTelemetry{
			perf: mPerf,
		}

		err := pt.DeactivatePerfEvents()
		require.ErrorContains(t, err, mError.Error())
		mPerf.AssertExpectations(t)
	})

	t.Run("SuccessfulDeactivation", func(t *testing.T) {
		mPerf := &perfMock{}
		mPerf.On("deactivate").Return(nil).Once()

		pt := &PowerTelemetry{
			perf: mPerf,
		}

		err := pt.DeactivatePerfEvents()
		require.NoError(t, err)
		mPerf.AssertExpectations(t)
	})
}

func TestPower_GetPackageIDs(t *testing.T) {
	testCases := []struct {
		name       string
		packageIDs []int
	}{
		{
			name:       "Empty",
			packageIDs: []int{},
		},
		{
			name:       "NotEmpty",
			packageIDs: []int{0, 1, 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pt := &PowerTelemetry{
				topology: &topologyData{
					packageIDs: tc.packageIDs,
				},
			}

			require.Equal(t, tc.packageIDs, pt.GetPackageIDs())
		})
	}
}

func TestPower_GetRaplPackageIDs(t *testing.T) {
	testCases := []struct {
		name       string
		raplZones  map[int]powerZone
		packageIDs []int
	}{
		{
			name:       "MapZonesIsNil",
			raplZones:  nil,
			packageIDs: []int{},
		},
		{
			name:       "MapZonesIsEmpty",
			raplZones:  map[int]powerZone{},
			packageIDs: []int{},
		},
		{
			name: "MapZonesIsUnordered",
			raplZones: map[int]powerZone{
				1: &zone{},
				0: &zone{},
				4: &zone{},
				2: &zone{},
			},
			packageIDs: []int{0, 1, 2, 4},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pt := &PowerTelemetry{
				rapl: &raplData{
					zones: tc.raplZones,
				},
			}

			require.Equal(t, tc.packageIDs, pt.GetRaplPackageIDs())
		})
	}

	t.Run("RaplIsNil", func(t *testing.T) {
		pt := &PowerTelemetry{
			rapl: nil,
		}

		require.Nil(t, pt.GetRaplPackageIDs())
	})
}
