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

func withTopologyMock(m *topologyMock) Option {
	return func(b *powerBuilder) {
		b.topology = &topologyBuilder{
			topologyReader: m,
		}
	}
}

func withMsrMock(m *msrMock) Option {
	return func(b *powerBuilder) {
		b.msr = &msrBuilder{
			msrReaderWithStorage: m,
		}
	}
}

func withRaplMock(m *raplMock) Option {
	return func(b *powerBuilder) {
		b.rapl = &raplBuilder{
			raplReader: m,
		}
	}
}

func withCoreFrequencyMock(m *coreFreqMock) Option {
	return func(b *powerBuilder) {
		b.coreFreq = &coreFreqBuilder{
			cpuFreqReader: m,
		}
	}
}

func withUncoreFrequencyMock(m *uncoreFreqMock) Option {
	return func(b *powerBuilder) {
		b.uncoreFreq = &uncoreFreqBuilder{
			uncoreFreqReader: m,
		}
	}
}

func withPerfMock(m *perfMock) Option {
	return func(b *powerBuilder) {
		b.perf = &perfBuilder{
			perfReaderWithStorage: m,
		}
	}
}

func TestWithExcludedCPUs(t *testing.T) {
	cpus := []int{0, 1, 2, 3}
	exp := &powerBuilder{
		excludedCPUs: cpus,
	}

	b := &powerBuilder{}
	f := WithExcludedCPUs(cpus)
	f(b)

	require.Equal(t, exp, b)
}

func TestWithIncludedCPUs(t *testing.T) {
	cpus := []int{0, 1, 2, 3}
	exp := &powerBuilder{
		includedCPUs: cpus,
	}

	b := &powerBuilder{}
	f := WithIncludedCPUs(cpus)
	f(b)

	require.Equal(t, exp, b)
}

func TestWithMsr(t *testing.T) {
	exp := &powerBuilder{
		msr: &msrBuilder{
			msrReaderWithStorage: &msrDataWithStorage{
				msrOffsets: cStateOffsets,
				msrPath:    defaultMsrBasePath,
			},
		},
	}

	b := &powerBuilder{}
	f := WithMsr()
	f(b)

	require.Equal(t, exp, b)
}

func TestWithMsrTimeout(t *testing.T) {
	exp := &powerBuilder{
		msr: &msrBuilder{
			msrReaderWithStorage: &msrDataWithStorage{
				msrOffsets: cStateOffsets,
				msrPath:    defaultMsrBasePath,
			},
			timeout: time.Minute,
		},
	}

	b := &powerBuilder{}
	f := WithMsrTimeout(time.Minute)
	f(b)

	require.Equal(t, exp, b)
}

func TestWithRapl(t *testing.T) {
	t.Run("DefaultBasePath", func(t *testing.T) {
		exp := &powerBuilder{
			rapl: &raplBuilder{
				raplReader: &raplData{
					basePath: defaultRaplBasePath,
				},
			},
		}

		b := &powerBuilder{}
		f := WithRapl()
		f(b)

		require.Equal(t, exp, b)
	})

	t.Run("CustomBasePath", func(t *testing.T) {
		customPath := "custom/rapl"
		exp := &powerBuilder{
			rapl: &raplBuilder{
				raplReader: &raplData{
					basePath: customPath,
				},
			},
		}

		b := &powerBuilder{}
		f := WithRapl(customPath)
		f(b)

		require.Equal(t, exp, b)
	})
}

func TestWithCoreFrequency(t *testing.T) {
	t.Run("DefaultBasePath", func(t *testing.T) {
		exp := &powerBuilder{
			coreFreq: &coreFreqBuilder{
				cpuFreqReader: &cpuFreqData{
					cpuFrequencyFilePath: defaultCPUFreqBasePath,
				},
			},
		}

		b := &powerBuilder{}
		f := WithCoreFrequency()
		f(b)

		require.Equal(t, exp, b)
	})

	t.Run("CustomBasePath", func(t *testing.T) {
		customPath := "custom/core_freq"
		exp := &powerBuilder{
			coreFreq: &coreFreqBuilder{
				cpuFreqReader: &cpuFreqData{
					cpuFrequencyFilePath: customPath,
				},
			},
		}

		b := &powerBuilder{}
		f := WithCoreFrequency(customPath)
		f(b)

		require.Equal(t, exp, b)
	})
}

func TestWithUncoreFrequency(t *testing.T) {
	t.Run("DefaultBasePath", func(t *testing.T) {
		exp := &powerBuilder{
			uncoreFreq: &uncoreFreqBuilder{
				uncoreFreqReader: &uncoreFreqData{
					uncoreFreqBasePath: defaultUncoreFreqBasePath,
				},
			},
		}

		b := &powerBuilder{}
		f := WithUncoreFrequency()
		f(b)

		require.Equal(t, exp, b)
	})

	t.Run("CustomBasePath", func(t *testing.T) {
		customPath := "custom/uncore_freq"
		exp := &powerBuilder{
			uncoreFreq: &uncoreFreqBuilder{
				uncoreFreqReader: &uncoreFreqData{
					uncoreFreqBasePath: customPath,
				},
			},
		}

		b := &powerBuilder{}
		f := WithUncoreFrequency(customPath)
		f(b)

		require.Equal(t, exp, b)
	})
}

func TestWithPerf(t *testing.T) {
	jsonFile := "testdata/sapphirerapids_core.json"

	b := &powerBuilder{}
	f := WithPerf(jsonFile)
	f(b)

	require.NotNil(t, b.perf)
	require.NotNil(t, b.perf.perfReaderWithStorage)
	require.Equal(t, jsonFile, b.perf.jsonPath)
	require.Equal(t, cStatePerfEvents, b.perf.events)
}

func TestGetAvailableCPUs(t *testing.T) {
	t.Run("WithAllCPUs", func(t *testing.T) {
		mTopo := &topologyMock{}

		// mock getting number of CPUs from powerBuilder.getAvailableCPUs
		mTopo.On("getCPUsNumber").Return(10).Once()

		cpusExp := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

		b := &powerBuilder{
			topology: &topologyBuilder{
				topologyReader: mTopo,
			},
		}

		cpusOut, err := b.getAvailableCPUs()
		require.NoError(t, err)
		require.Equal(t, cpusExp, cpusOut)
	})

	t.Run("WithIncludedCPUs", func(t *testing.T) {
		mTopo := &topologyMock{}

		// mock getting number of CPUs from powerBuilder.getAvailableCPUs
		// valid CPU IDs within the range [0, 19]
		mTopo.On("getCPUsNumber").Return(20)

		t.Run("OutOfBounds", func(t *testing.T) {
			includedCPUs := []int{15, 16, 17, 18, 19, 20}

			b := &powerBuilder{
				topology: &topologyBuilder{
					topologyReader: mTopo,
				},
			}
			f := WithIncludedCPUs(includedCPUs)
			f(b)

			cpusOut, err := b.getAvailableCPUs()
			require.ErrorContains(t, err, "failed to validate included CPU IDs")
			require.ErrorContains(t, err, "20 is out of bounds [0, 19]")
			require.Nil(t, cpusOut)
			mTopo.AssertExpectations(t)
		})

		t.Run("NilCPUsSlice", func(t *testing.T) {
			b := &powerBuilder{
				topology: &topologyBuilder{
					topologyReader: mTopo,
				},
			}
			f := WithIncludedCPUs(nil)
			f(b)

			cpusOut, err := b.getAvailableCPUs()
			require.NoError(t, err)
			require.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}, cpusOut)
			mTopo.AssertExpectations(t)
		})

		t.Run("EmptyCPUsSlice", func(t *testing.T) {
			b := &powerBuilder{
				topology: &topologyBuilder{
					topologyReader: mTopo,
				},
			}
			f := WithIncludedCPUs([]int{})
			f(b)

			cpusOut, err := b.getAvailableCPUs()
			require.NoError(t, err)
			require.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}, cpusOut)
			mTopo.AssertExpectations(t)
		})

		t.Run("PartialCPUsIncluded", func(t *testing.T) {
			cpusExp := []int{0, 1, 2, 3}

			b := &powerBuilder{
				topology: &topologyBuilder{
					topologyReader: mTopo,
				},
			}
			f := WithIncludedCPUs(cpusExp)
			f(b)

			cpusOut, err := b.getAvailableCPUs()
			require.NoError(t, err)
			require.Equal(t, cpusExp, cpusOut)
			mTopo.AssertExpectations(t)
		})
	})

	t.Run("WithExcludedCPUs", func(t *testing.T) {
		mTopo := &topologyMock{}

		// mock getting number of CPUs from powerBuilder.getAvailableCPUs
		mTopo.On("getCPUsNumber").Return(10)

		t.Run("OutOfBounds", func(t *testing.T) {
			excludedCPUs := []int{2, 1, 0, 12, 13}

			b := &powerBuilder{
				topology: &topologyBuilder{
					topologyReader: mTopo,
				},
			}
			f := WithExcludedCPUs(excludedCPUs)
			f(b)

			cpusOut, err := b.getAvailableCPUs()
			require.ErrorContains(t, err, "failed to validate excluded CPU IDs")
			require.ErrorContains(t, err, "12 is out of bounds [0, 9]")
			require.Nil(t, cpusOut)
			mTopo.AssertExpectations(t)
		})

		t.Run("AllCPUsExcluded", func(t *testing.T) {
			excludedCPUs := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

			b := &powerBuilder{
				topology: &topologyBuilder{
					topologyReader: mTopo,
				},
			}
			f := WithExcludedCPUs(excludedCPUs)
			f(b)

			cpusOut, err := b.getAvailableCPUs()
			require.NoError(t, err)
			require.Empty(t, cpusOut)
			mTopo.AssertExpectations(t)
		})

		t.Run("NilCPUsSlice", func(t *testing.T) {
			b := &powerBuilder{
				topology: &topologyBuilder{
					topologyReader: mTopo,
				},
			}
			f := WithExcludedCPUs(nil)
			f(b)

			cpusOut, err := b.getAvailableCPUs()
			require.NoError(t, err)
			require.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, cpusOut)
			mTopo.AssertExpectations(t)
		})

		t.Run("EmptyCPUsSlice", func(t *testing.T) {
			b := &powerBuilder{
				topology: &topologyBuilder{
					topologyReader: mTopo,
				},
			}
			f := WithExcludedCPUs([]int{})
			f(b)

			cpusOut, err := b.getAvailableCPUs()
			require.NoError(t, err)
			require.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, cpusOut)
			mTopo.AssertExpectations(t)
		})

		t.Run("PartialCPUsExcluded", func(t *testing.T) {
			cpusExcluded := []int{0, 1, 2, 3}
			cpusExp := []int{4, 5, 6, 7, 8, 9}

			b := &powerBuilder{
				topology: &topologyBuilder{
					topologyReader: mTopo,
				},
			}
			f := WithExcludedCPUs(cpusExcluded)
			f(b)

			cpusOut, err := b.getAvailableCPUs()
			require.NoError(t, err)
			require.Equal(t, cpusExp, cpusOut)
			mTopo.AssertExpectations(t)
		})
	})

	t.Run("WithExcludedAndIncludedCPUs", func(t *testing.T) {
		b := &powerBuilder{}
		withIncluded := WithIncludedCPUs([]int{0, 1, 2, 3})
		withExcluded := WithExcludedCPUs([]int{0, 2})

		withIncluded(b)
		withExcluded(b)

		cpusOut, err := b.getAvailableCPUs()
		require.ErrorContains(t, err, "invalid CPU ID configuration, only one of both included or excluded modes allowed")
		require.Nil(t, cpusOut)
	})
}

func TestNew(t *testing.T) {
	mError := errors.New("mock error")

	t.Run("FailedToInitTopology", func(t *testing.T) {
		mTopology := &topologyMock{}

		// mock initializing topology map
		mTopology.On("initTopology").Return(mError).Once()

		pt, err := New(
			withTopologyMock(mTopology),
		)

		require.ErrorContains(t, err, mError.Error())
		require.Nil(t, pt)

		mTopology.AssertExpectations(t)
	})

	t.Run("FailedToCheckCPU", func(t *testing.T) {
		mTopology := &topologyMock{}

		// mock initializing topology map
		mTopology.On("initTopology").Return(nil).Once()

		// TODO: Consider to make isCPUSupported as topology method
		// mock getting CPU family and vendor from isCPUSupported
		mTopology.On("getCPUFamily", 0).Return("6", nil).Once()
		mTopology.On("getCPUVendor", 0).Return("", mError).Once()

		// mock getting topology CPU data from logTopologyDetails
		mTopology.On("getCPUFamily", 0).Return("6", nil).Once()
		mTopology.On("getCPUVendor", 0).Return("", mError).Once()
		mTopology.On("getCPUModel").Return(cpumodel.INTEL_FAM6_ALDERLAKE, nil).Once()
		mTopology.On("getCPUsNumber").Return(1, nil).Once()
		mTopology.On("getCPUCoreID", 0).Return(0, nil).Once()
		mTopology.On("getCPUPackageID", 0).Return(0, nil).Once()
		mTopology.On("getCPUDieID", 0).Return(0, nil).Once()

		pt, err := New(
			withTopologyMock(mTopology),
		)

		require.ErrorContains(t, err, "error retrieving host processor")
		require.Nil(t, pt)

		mTopology.AssertExpectations(t)
	})

	t.Run("CPUNotSupported", func(t *testing.T) {
		mTopology := &topologyMock{}

		// mock initializing topology map
		mTopology.On("initTopology").Return(nil).Once()

		// mock getting CPU family and vendor from isCPUSupported
		mTopology.On("getCPUFamily", 0).Return("6", nil).Once()
		mTopology.On("getCPUVendor", 0).Return("AuthenticAMD", nil).Once()

		// mock getting topology CPU data from logTopologyDetails
		mTopology.On("getCPUFamily", 0).Return("6", nil).Once()
		mTopology.On("getCPUVendor", 0).Return("AuthenticAMD", nil).Once()
		mTopology.On("getCPUModel").Return(cpumodel.INTEL_FAM6_ALDERLAKE, nil).Once()
		mTopology.On("getCPUsNumber").Return(1, nil).Once()
		mTopology.On("getCPUCoreID", 0).Return(0, nil).Once()
		mTopology.On("getCPUPackageID", 0).Return(0, nil).Once()
		mTopology.On("getCPUDieID", 0).Return(0, nil).Once()

		pt, err := New(
			withTopologyMock(mTopology),
		)

		require.ErrorContains(t, err, "host processor is not supported")
		require.Nil(t, pt)

		mTopology.AssertExpectations(t)
	})

	t.Run("FailedToGetAvailableCPUs", func(t *testing.T) {
		mTopology := &topologyMock{}

		// mock initializing topology map
		mTopology.On("initTopology").Return(nil)

		// mock getting CPU family and vendor from isCPUSupported
		mTopology.On("getCPUFamily", 0).Return("6", nil)
		mTopology.On("getCPUVendor", 0).Return("GenuineIntel", nil)

		// mock getting topology CPU data from logTopologyDetails
		mTopology.On("getCPUModel").Return(cpumodel.INTEL_FAM6_ALDERLAKE, nil).Once()
		mTopology.On("getCPUsNumber").Return(1, nil).Once()
		mTopology.On("getCPUCoreID", 0).Return(0, nil).Once()
		mTopology.On("getCPUPackageID", 0).Return(0, nil).Once()
		mTopology.On("getCPUDieID", 0).Return(0, nil).Once()

		pt, err := New(
			withTopologyMock(mTopology),

			WithExcludedCPUs([]int{5, 6, 7, 8, 9}),
			WithIncludedCPUs([]int{0, 1, 2, 3, 4}),
		)

		require.ErrorContains(t, err, "failed to get available CPUs")
		require.Nil(t, pt)

		mTopology.AssertExpectations(t)
	})

	t.Run("AllCPUsExcluded", func(t *testing.T) {
		cpus := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

		mTopology := &topologyMock{}

		// mock initializing topology map
		mTopology.On("initTopology").Return(nil)

		// mock getting CPU family and vendor from isCPUSupported
		mTopology.On("getCPUFamily", 0).Return("6", nil)
		mTopology.On("getCPUVendor", 0).Return("GenuineIntel", nil)

		// mock getting topology CPU data from logTopologyDetails
		mTopology.On("getCPUModel").Return(cpumodel.INTEL_FAM6_ALDERLAKE, nil).Once()
		mTopology.On("getCPUsNumber").Return(1, nil).Once()
		mTopology.On("getCPUCoreID", 0).Return(0, nil).Once()
		mTopology.On("getCPUPackageID", 0).Return(0, nil).Once()
		mTopology.On("getCPUDieID", 0).Return(0, nil).Once()

		// mock getting number of CPU IDs from powerBuilder.getAvailableCPUs
		mTopology.On("getCPUsNumber").Return(len(cpus))

		pt, err := New(
			withTopologyMock(mTopology),

			WithExcludedCPUs(cpus),
		)

		require.ErrorContains(t, err, "no available CPUs were found")
		require.Nil(t, pt)

		mTopology.AssertExpectations(t)
	})

	t.Run("With", func(t *testing.T) {
		cpus := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

		mTopology := &topologyMock{}

		// mock initializing topology map
		mTopology.On("initTopology").Return(nil)

		// mock getting CPU family and vendor from isCPUSupported
		mTopology.On("getCPUFamily", 0).Return("6", nil)
		mTopology.On("getCPUVendor", 0).Return("GenuineIntel", nil)

		// mock getting model to calculate bus clock
		mTopology.On("getCPUModel").Return(cpumodel.INTEL_FAM6_ICELAKE)

		// mock getting number of CPU IDs from powerBuilder.getAvailableCPUs
		mTopology.On("getCPUsNumber").Return(len(cpus))

		// mock getting topology CPU data from logTopologyDetails
		mTopology.On("getCPUCoreID", mock.AnythingOfType("int")).Return(0, nil)
		mTopology.On("getCPUPackageID", mock.AnythingOfType("int")).Return(0, nil)
		mTopology.On("getCPUDieID", mock.AnythingOfType("int")).Return(0, nil)

		t.Run("Msr", func(t *testing.T) {
			t.Run("FailedToInitMsrMap", func(t *testing.T) {
				mMsr := &msrMock{}

				// mock initializing msr map from powerBuilder.initMsr
				mMsr.On("initMsrMap", cpus, time.Duration(0)).Return(mError).Once()

				pt, err := New(
					withTopologyMock(mTopology),
					withMsrMock(mMsr),
				)

				require.ErrorContains(t, err, "failed to initialize msr")
				require.NotNil(t, pt)
				require.Nil(t, pt.msr)
				require.Nil(t, pt.GetRaplPackageIDs())
				require.Equal(t, cpus, pt.GetMsrCPUIDs())

				mTopology.AssertExpectations(t)
				mMsr.AssertExpectations(t)
			})

			t.Run("IncludedCPUs", func(t *testing.T) {
				includedCPUs := []int{0, 1, 2, 3, 4}

				mMsr := &msrMock{}

				// mock initializing msr map from powerBuilder.initMsr
				mMsr.On("initMsrMap", includedCPUs, time.Duration(0)).Return(nil).Once()

				pt, err := New(
					withTopologyMock(mTopology),
					withMsrMock(mMsr),

					WithIncludedCPUs(includedCPUs),
				)

				require.NoError(t, err)
				require.NotNil(t, pt)
				require.NotNil(t, pt.msr)
				require.Nil(t, pt.GetRaplPackageIDs())
				require.Equal(t, includedCPUs, pt.GetMsrCPUIDs())

				mTopology.AssertExpectations(t)
				mMsr.AssertExpectations(t)
			})

			t.Run("ExcludedCPUs", func(t *testing.T) {
				excludedCPUs := []int{0, 1, 2, 3, 4}
				availableCPUs := []int{5, 6, 7, 8, 9}

				mMsr := &msrMock{}

				// mock initializing msr map from powerBuilder.initMsr
				mMsr.On("initMsrMap", availableCPUs, time.Duration(0)).Return(nil).Once()

				pt, err := New(
					withTopologyMock(mTopology),
					withMsrMock(mMsr),

					WithExcludedCPUs(excludedCPUs),
				)

				require.NoError(t, err)
				require.NotNil(t, pt)
				require.NotNil(t, pt.msr)
				require.Nil(t, pt.GetRaplPackageIDs())
				require.Equal(t, availableCPUs, pt.GetMsrCPUIDs())

				mTopology.AssertExpectations(t)
				mMsr.AssertExpectations(t)
			})
		})

		t.Run("Rapl", func(t *testing.T) {
			t.Run("FailedToInitZoneMap", func(t *testing.T) {
				mRapl := &raplMock{}

				// mock initializing rapl zone map from powerBuilder.initRapl
				mRapl.On("initZoneMap").Return(mError).Once()

				pt, err := New(
					withTopologyMock(mTopology),
					withRaplMock(mRapl),
				)

				require.ErrorContains(t, err, "failed to initialize rapl")
				require.NotNil(t, pt)
				require.Nil(t, pt.rapl)
				require.Nil(t, pt.GetRaplPackageIDs())
				require.Equal(t, cpus, pt.GetMsrCPUIDs())

				mTopology.AssertExpectations(t)
				mRapl.AssertExpectations(t)
			})

			t.Run("Ok", func(t *testing.T) {
				pt, err := New(
					withTopologyMock(mTopology),
					WithRapl(makeTestDataPath("testdata/intel-rapl")),
				)

				require.NoError(t, err)
				require.NotNil(t, pt)
				require.NotNil(t, pt.rapl)
				require.Equal(t, []int{0, 1, 2, 3}, pt.GetRaplPackageIDs())
				require.Equal(t, cpus, pt.GetMsrCPUIDs())

				mTopology.AssertExpectations(t)
			})
		})

		t.Run("CoreFrequency", func(t *testing.T) {
			// TODO: Consider to remove
			t.Run("FailedToInit", func(t *testing.T) {
				mCoreFreq := &coreFreqMock{}

				// mock initializing core frequency from powerBuilder.initCoreFreq
				mCoreFreq.On("init").Return(mError).Once()

				pt, err := New(
					withTopologyMock(mTopology),
					withCoreFrequencyMock(mCoreFreq),
				)

				require.ErrorContains(t, err, "failed to initialize core freq")
				require.NotNil(t, pt)
				require.Nil(t, pt.cpuFreq)
				require.Nil(t, pt.GetRaplPackageIDs())
				require.Equal(t, cpus, pt.GetMsrCPUIDs())

				mTopology.AssertExpectations(t)
				mCoreFreq.AssertExpectations(t)
			})

			t.Run("Ok", func(t *testing.T) {
				pt, err := New(
					withTopologyMock(mTopology),
					WithCoreFrequency("testdata/cpu-freq"),
				)

				require.NoError(t, err)
				require.NotNil(t, pt)
				require.NotNil(t, pt.cpuFreq)
				require.Nil(t, pt.GetRaplPackageIDs())
				require.Equal(t, cpus, pt.GetMsrCPUIDs())

				mTopology.AssertExpectations(t)
			})
		})

		t.Run("UncoreFrequency", func(t *testing.T) {
			// TODO: Consider to remove
			t.Run("FailedToInit", func(t *testing.T) {
				mUncoreFreq := &uncoreFreqMock{}

				// mock initializing uncore frequency from powerBuilder.initUncoreFreq
				mUncoreFreq.On("init").Return(mError).Once()

				pt, err := New(
					withTopologyMock(mTopology),
					withUncoreFrequencyMock(mUncoreFreq),
				)

				require.ErrorContains(t, err, "failed to initialize uncore freq")
				require.NotNil(t, pt)
				require.Nil(t, pt.cpuFreq)
				require.Nil(t, pt.GetRaplPackageIDs())
				require.Equal(t, cpus, pt.GetMsrCPUIDs())

				mTopology.AssertExpectations(t)
				mUncoreFreq.AssertExpectations(t)
			})

			t.Run("Ok", func(t *testing.T) {
				pt, err := New(
					withTopologyMock(mTopology),
					WithUncoreFrequency("testdata/intel_uncore_frequency"),
				)

				require.NoError(t, err)
				require.NotNil(t, pt)
				require.NotNil(t, pt.uncoreFreq)
				require.Nil(t, pt.GetRaplPackageIDs())
				require.Equal(t, cpus, pt.GetMsrCPUIDs())

				mTopology.AssertExpectations(t)
			})
		})

		t.Run("Perf", func(t *testing.T) {
			// Reset mock object for perf
			mTopology = &topologyMock{}

			// mock initializing topology map
			mTopology.On("initTopology").Return(nil)

			// mock getting CPU family and vendor from isCPUSupported
			mTopology.On("getCPUFamily", 0).Return("6", nil)
			mTopology.On("getCPUVendor", 0).Return("GenuineIntel", nil)

			// mock getting number of CPUs from powerBuilder.getAvailableCPUs
			mTopology.On("getCPUsNumber").Return(len(cpus))

			// mock getting topology CPU data from logTopologyDetails
			mTopology.On("getCPUCoreID", mock.AnythingOfType("int")).Return(0, nil)
			mTopology.On("getCPUPackageID", mock.AnythingOfType("int")).Return(0, nil)
			mTopology.On("getCPUDieID", mock.AnythingOfType("int")).Return(0, nil)

			t.Run("FailedToGetCPUModel", func(t *testing.T) {
				model := cpumodel.INTEL_FAM6_ICELAKE

				// mock getting model from powerBuilder.initPerf and logTopologyDetails
				mTopology.On("getCPUModel").Return(0).Twice()

				// mock getting model to calculate bus clock
				mTopology.On("getCPUModel").Return(model).Once()

				pt, err := New(
					withTopologyMock(mTopology),
					WithPerf("events.json"),
				)

				require.NotNil(t, pt)
				require.Nil(t, pt.perf)
				require.ErrorContains(t, err, "failed to initialize perf: perf based metrics are not supported for processor model: 0x0")
				require.Nil(t, pt.GetRaplPackageIDs())
				require.Equal(t, cpus, pt.GetPerfCPUIDs())

				mTopology.AssertExpectations(t)
			})

			t.Run("CPUModelNotAllowed", func(t *testing.T) {
				model := cpumodel.INTEL_FAM6_ICELAKE

				// mock getting model from isCPUSupported, powerBuilder.initPerf and logTopologyDetails
				mTopology.On("getCPUModel").Return(model).Times(3)

				pt, err := New(
					withTopologyMock(mTopology),
					WithPerf("events.json"),
				)

				require.NotNil(t, pt)
				require.Nil(t, pt.perf)
				require.ErrorContains(t, err, "failed to initialize perf")
				require.ErrorContains(t, err, fmt.Sprintf("perf based metrics are not supported for processor model: 0x%X", model))
				require.Nil(t, pt.GetRaplPackageIDs())
				require.Equal(t, cpus, pt.GetPerfCPUIDs())

				mTopology.AssertExpectations(t)
			})

			t.Run("FailedToInitResolver", func(t *testing.T) {
				// mock getting model from isCPUSupported, powerBuilder.initPerf and logTopologyDetails
				mTopology.On("getCPUModel").Return(cpumodel.INTEL_FAM6_EMERALDRAPIDS_X).Times(3)

				mPerf := &perfMock{}

				// mock initialize perf resolver from powerBuilder.initPerf
				mPerf.On("initResolver", mock.AnythingOfType("string")).Return(mError).Once()

				pt, err := New(
					withTopologyMock(mTopology),
					withPerfMock(mPerf),
				)

				require.NotNil(t, pt)
				require.Nil(t, pt.perf)
				require.ErrorContains(t, err, "failed to initialize perf")
				require.ErrorContains(t, err, "failed to init resolver")
				require.Nil(t, pt.GetRaplPackageIDs())
				require.Equal(t, cpus, pt.GetPerfCPUIDs())

				mTopology.AssertExpectations(t)
				mPerf.AssertExpectations(t)
			})

			t.Run("FailedToActivate", func(t *testing.T) {
				// mock getting model from isCPUSupported, powerBuilder.initPerf and logTopologyDetails
				mTopology.On("getCPUModel").Return(cpumodel.INTEL_FAM6_EMERALDRAPIDS_X).Times(3)

				mPerf := &perfMock{}

				// mock initializing perf resolver and event activation from powerBuilder.initPerf
				mPerf.On("initResolver", mock.AnythingOfType("string")).Return(nil).Once()
				mPerf.On("activate", cStatePerfEvents, cpus).Return(mError).Once()

				pt, err := New(
					withTopologyMock(mTopology),
					withPerfMock(mPerf),
				)

				require.NotNil(t, pt)
				require.Nil(t, pt.perf)
				require.ErrorContains(t, err, "failed to initialize perf")
				require.ErrorContains(t, err, "failed to activate events")
				require.Nil(t, pt.GetRaplPackageIDs())
				require.Equal(t, cpus, pt.GetPerfCPUIDs())

				mTopology.AssertExpectations(t)
				mPerf.AssertExpectations(t)
			})

			t.Run("IncludedCPUs", func(t *testing.T) {
				includedCPUs := []int{1, 2, 3, 4}

				// mock getting model from isCPUSupported, powerBuilder.initPerf and logTopologyDetails
				mTopology.On("getCPUModel").Return(cpumodel.INTEL_FAM6_SAPPHIRERAPIDS_X).Times(3)

				mPerf := &perfMock{}

				// mock initializing perf resolver and event activation from powerBuilder.initPerf
				mPerf.On("initResolver", mock.AnythingOfType("string")).Return(nil).Once()
				mPerf.On("activate", cStatePerfEvents, includedCPUs).Return(nil).Once()

				pt, err := New(
					withTopologyMock(mTopology),
					withPerfMock(mPerf),

					WithIncludedCPUs(includedCPUs),
				)

				require.NotNil(t, pt)
				require.NotNil(t, pt.perf)
				require.NoError(t, err)
				require.Nil(t, pt.GetRaplPackageIDs())
				require.Equal(t, includedCPUs, pt.GetPerfCPUIDs())

				mTopology.AssertExpectations(t)
				mPerf.AssertExpectations(t)
			})

			t.Run("ExcludedCPUs", func(t *testing.T) {
				excludedCPUs := []int{0, 1, 8, 9}
				availableCPUs := []int{2, 3, 4, 5, 6, 7}

				// mock getting model from isCPUSupported, powerBuilder.initPerf and logTopologyDetails
				mTopology.On("getCPUModel").Return(cpumodel.INTEL_FAM6_SAPPHIRERAPIDS_X).Times(3)

				mPerf := &perfMock{}

				// mock initializing perf resolver and event activation from powerBuilder.initPerf
				mPerf.On("initResolver", mock.AnythingOfType("string")).Return(nil).Once()
				mPerf.On("activate", cStatePerfEvents, availableCPUs).Return(nil).Once()

				pt, err := New(
					withTopologyMock(mTopology),
					withPerfMock(mPerf),

					WithExcludedCPUs(excludedCPUs),
				)

				require.NotNil(t, pt)
				require.NotNil(t, pt.perf)
				require.NoError(t, err)
				require.Nil(t, pt.GetRaplPackageIDs())
				require.Equal(t, availableCPUs, pt.GetPerfCPUIDs())

				mTopology.AssertExpectations(t)
				mPerf.AssertExpectations(t)
			})
		})

		t.Run("Multiple", func(t *testing.T) {
			t.Run("FailedMsrAndPerf", func(t *testing.T) {
				includedCPUs := []int{0, 2, 4, 6, 8}

				// mock getting model from isCPUSupported, powerBuilder.initPerf and logTopologyDetails
				mTopology.On("getCPUModel").Return(cpumodel.INTEL_FAM6_EMERALDRAPIDS_X).Times(3)

				mMsr := &msrMock{}

				// mock initializing msr map from powerBuilder.initMsr
				mMsr.On("initMsrMap", includedCPUs, time.Duration(0)).Return(mError).Once()

				mPerf := &perfMock{}

				// mock initializing perf resolver and event activation from powerBuilder.initPerf
				mPerf.On("initResolver", mock.AnythingOfType("string")).Return(nil).Once()
				mPerf.On("activate", cStatePerfEvents, includedCPUs).Return(mError).Once()

				pt, err := New(
					withTopologyMock(mTopology),
					withMsrMock(mMsr),
					withPerfMock(mPerf),
					WithRapl(makeTestDataPath("testdata/intel-rapl")),
					WithCoreFrequency("testdata/cpu-freq"),
					WithUncoreFrequency("testdata/intel_uncore_frequency"),

					WithIncludedCPUs(includedCPUs),
				)

				require.ErrorContains(t, err, "failed to initialize msr")
				require.ErrorContains(t, err, "failed to initialize perf")
				require.Nil(t, pt.msr)
				require.Nil(t, pt.perf)
				require.NotNil(t, pt.rapl)
				require.NotNil(t, pt.cpuFreq)
				require.NotNil(t, pt.uncoreFreq)
				require.Equal(t, []int{0, 1, 2, 3}, pt.GetRaplPackageIDs())
				require.Equal(t, includedCPUs, pt.GetPerfCPUIDs())

				mTopology.AssertExpectations(t)
				mMsr.AssertExpectations(t)
				mPerf.AssertExpectations(t)
			})

			t.Run("Ok", func(t *testing.T) {
				excludedCPUs := []int{1, 3, 5, 7, 9}

				// mock getting model from isCPUSupported and logTopologyDetails
				mTopology.On("getCPUModel").Return(cpumodel.INTEL_FAM6_ALDERLAKE).Twice()

				pt, err := New(
					withTopologyMock(mTopology),
					WithCoreFrequency("testdata/cpu-freq"),
					WithUncoreFrequency("testdata/intel_uncore_frequency"),

					WithExcludedCPUs(excludedCPUs),
				)

				require.NoError(t, err)
				require.NotNil(t, pt)
				require.NotNil(t, pt.cpuFreq)
				require.NotNil(t, pt.uncoreFreq)
				require.Nil(t, pt.GetRaplPackageIDs())
				require.Equal(t, []int{0, 2, 4, 6, 8}, pt.GetPerfCPUIDs())

				mTopology.AssertExpectations(t)
			})
		})
	})

	t.Run("FailedToGetBusClock", func(t *testing.T) {
		cpus := []int{2, 3, 4, 5, 6, 7, 8, 9}
		model := cpumodel.INTEL_FAM6_ATOM_SILVERMONT

		mTopology := &topologyMock{}

		// mock initializing topology map
		mTopology.On("initTopology").Return(nil)

		// mock getting CPU family and vendor from isCPUSupported
		mTopology.On("getCPUFamily", 0).Return("6", nil)
		mTopology.On("getCPUVendor", 0).Return("GenuineIntel", nil)

		// mock getting model to calculate bus clock
		mTopology.On("getCPUModel").Return(model)

		// mock getting number of CPUs from powerBuilder.getAvailableCPUs
		mTopology.On("getCPUsNumber").Return(10)

		// mock getting topology CPU data from logTopologyDetails
		mTopology.On("getCPUCoreID", mock.AnythingOfType("int")).Return(0, nil)
		mTopology.On("getCPUPackageID", mock.AnythingOfType("int")).Return(0, nil)
		mTopology.On("getCPUDieID", mock.AnythingOfType("int")).Return(0, nil)

		t.Run("MsrIsNil", func(t *testing.T) {
			pt, err := New(
				withTopologyMock(mTopology),
				WithIncludedCPUs(cpus),
			)

			require.ErrorContains(t, err, fmt.Sprintf("failed to get bus clock for model: 0x%X", model))
			require.ErrorContains(t, err, "\"msr\" is not initialized")
			require.NotNil(t, pt)
			require.Equal(t, 0.0, pt.busClock)
			require.Nil(t, pt.GetRaplPackageIDs())
			require.Equal(t, cpus, pt.GetMsrCPUIDs())

			mTopology.AssertExpectations(t)
		})

		t.Run("WithMsr", func(t *testing.T) {
			t.Run("ExcludedCPUs", func(t *testing.T) {
				excludedCPUs := []int{0, 1, 2, 3, 4}
				availableCPUs := []int{5, 6, 7, 8, 9}

				mMsr := &msrMock{}

				// mock initializing msr map from powerBuilder.initMsr
				mMsr.On("initMsrMap", availableCPUs, time.Duration(0)).Return(nil).Once()

				// mock reading msr offset MSR_FSB_FREQ
				mMsr.On("read", uint32(fsbFreq), availableCPUs[0]).Return(uint64(0), mError).Once()

				// TODO: Call to WithIncludedCPUs, for instance {1,2,3, ...} check that the mock still works
				pt, err := New(
					withTopologyMock(mTopology),
					withMsrMock(mMsr),

					WithExcludedCPUs(excludedCPUs),
				)

				require.ErrorContains(t, err, fmt.Sprintf("failed to get bus clock for model: 0x%X", model))
				require.NotNil(t, pt)
				require.Equal(t, 0.0, pt.busClock)
				require.Nil(t, pt.GetRaplPackageIDs())
				require.Equal(t, availableCPUs, pt.GetMsrCPUIDs())

				mTopology.AssertExpectations(t)
				mMsr.AssertExpectations(t)
			})

			t.Run("IncludedCPUs", func(t *testing.T) {
				includedCPUs := []int{5, 6, 7, 8, 9}

				mMsr := &msrMock{}

				// mock initializing msr map from powerBuilder.initMsr
				mMsr.On("initMsrMap", includedCPUs, time.Duration(0)).Return(nil).Once()

				// mock reading msr offset MSR_FSB_FREQ
				mMsr.On("read", uint32(fsbFreq), includedCPUs[0]).Return(uint64(0), mError).Once()

				pt, err := New(
					withTopologyMock(mTopology),
					withMsrMock(mMsr),

					WithIncludedCPUs(includedCPUs),
				)

				require.ErrorContains(t, err, fmt.Sprintf("failed to get bus clock for model: 0x%X", model))
				require.NotNil(t, pt)
				require.Equal(t, 0.0, pt.busClock)
				require.Nil(t, pt.GetRaplPackageIDs())
				require.Equal(t, includedCPUs, pt.GetMsrCPUIDs())

				mTopology.AssertExpectations(t)
				mMsr.AssertExpectations(t)
			})
		})
	})
}

func Test_IsPerfAllowed(t *testing.T) {
	models := []int{
		0xCF, //INTEL_FAM6_EMERALDRAPIDS_X
		0x8F, //INTEL_FAM6_SAPPHIRERAPIDS_X
		//TODO: Hybrid models are not supported right now
		//0x97, //INTEL_FAM6_ALDERLAKE
		//0x9A, //INTEL_FAM6_ALDERLAKE_L
		//0xB7, //INTEL_FAM6_RAPTORLAKE
		//0xBA, //INTEL_FAM6_RAPTORLAKE_P
		//0xBF, //INTEL_FAM6_RAPTORLAKE_S
		//0xAC, //INTEL_FAM6_METEORLAKE
		//0xAA, //INTEL_FAM6_METEORLAKE_L
	}

	m := map[int]interface{}{}
	for _, model := range models {
		m[model] = struct{}{}
	}

	for model := 0; model < 0xFF; model++ {
		isAllowed := isPerfAllowed(model)
		require.Equalf(t, m[model] != nil, isAllowed, "Model 0x%X")
	}
}
