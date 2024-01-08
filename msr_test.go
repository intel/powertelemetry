// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestNewMsr(t *testing.T) {
	testCases := []struct {
		name string
		path string
		msr  msrReg
		err  error
	}{
		{
			name: "CPUIDNonNumeric",
			path: "testdata/cpu-msr-invalid-cpuID-directories/1invalid",
			msr:  nil,
			err:  errors.New("invalid format for CPU ID in path"),
		},
		{
			name: "CPUIDWithLeadingZeroes",
			path: "testdata/cpu-msr-invalid-cpuID-directories/01",
			msr:  nil,
			err:  errors.New("invalid format for CPU ID in path"),
		},
		{
			name: "CPUIDMsrFileNotExist",
			path: "testdata/cpu-msr-cpuID-msr-not-exist/1",
			msr:  nil,
			err:  fmt.Errorf("invalid MSR file for cpu ID 1: file \"testdata/cpu-msr-cpuID-msr-not-exist/1/msr\" does not exist"),
		},
		{
			name: "CPUIDMsrFileSymlink",
			path: "testdata/cpu-msr-cpuID-msr-softlink/1",
			msr:  nil,
			err:  fmt.Errorf("invalid MSR file for cpu ID 1: file \"testdata/cpu-msr-cpuID-msr-softlink/1/msr\" is a symlink"),
		},
		{
			name: "Valid",
			path: "testdata/cpu-msr/0",
			msr: &msr{
				path:  "testdata/cpu-msr/0/msr",
				cpuID: 0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := newMsr(tc.path, 0)
			require.Equal(t, tc.msr, m)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsrGetters(t *testing.T) {
	cpuID := 0
	cpuPath := "testdata/cpu-msr/0"
	m, err := newMsr(cpuPath, 0)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(cpuPath, msrFile), m.getPath())
	require.Equal(t, cpuID, m.getCPUID())
}

// TestMsrRead tests read a single offset value of the corresponding MSR register. Valid MSR registers have
// paths within "testdata/cpu-msr" directory. Each directory corresponds to CPU ID-specific MSR path, which
// comprises a binary file named "msr". All valid MSR binary files have the same 16-byte content, written byte
// by byte, using little endian.
//
// 0x01 0x23 0x45 0x67 0x89 0xab 0xcd 0xef
// 0xfe 0xdc 0xba 0x98 0x76 0x54 0x32 0x10
//
// Example:
// The result of reading operation given offset 0, would result in an 8-byte value corresponding to the first row:
//
// 0x01 0x23 0x45 0x67 0x89 0xab 0xcd 0xef
//
// Since values were written byte by byte using little endian, the result of the operation is:
//
// 0xef 0xcd 0xab 0x89 0x67 0x45 0x23 0x01 -> 0xefcdab8967452301.
func TestMsrRead(t *testing.T) {
	testCases := []struct {
		name       string
		cpuMsrPath string
		offset     uint32
		timeout    time.Duration
		expected   uint64
		err        error
	}{
		{
			name:       "MsrFileNotExists",
			cpuMsrPath: "testdata/cpu-msr-cpuID-msr-not-exist/0/msr",
			offset:     0x0,
			expected:   0,
			err:        errors.New("open testdata/cpu-msr-cpuID-msr-not-exist/0/msr: no such file or directory"),
		},
		{
			name:       "ZeroBytesOffset",
			cpuMsrPath: "testdata/cpu-msr/0/msr",
			offset:     0x0,
			expected:   0xefcdab8967452301,
			err:        nil,
		},
		{
			name:       "4BytesOffset",
			cpuMsrPath: "testdata/cpu-msr/0/msr",
			offset:     0x4,
			expected:   0x98badcfeefcdab89,
			err:        nil,
		},
		{
			name:       "8BytesOffset",
			cpuMsrPath: "testdata/cpu-msr/0/msr",
			offset:     0x8,
			expected:   0x1032547698badcfe,
			err:        nil,
		},
		{
			name:       "OutOfBoundsOffset",
			cpuMsrPath: "testdata/cpu-msr/0/msr",
			offset:     0xb,
			expected:   0,
			err:        errors.New("offset 0xb is out-of-bounds"),
		},
		{
			name:       "MsrFileNotExistsWithExtremelyLargeTimeout",
			cpuMsrPath: "testdata/cpu-msr-cpuID-msr-not-exist/0/msr",
			offset:     0x0,
			timeout:    time.Hour,
			expected:   0,
			err:        errors.New("open testdata/cpu-msr-cpuID-msr-not-exist/0/msr: no such file or directory"),
		},
		{
			name:       "ZeroBytesOffsetWithExtremelyLargeTimeout",
			cpuMsrPath: "testdata/cpu-msr/0/msr",
			offset:     0x0,
			timeout:    time.Hour,
			expected:   0xefcdab8967452301,
			err:        nil,
		},
		{
			name:       "4BytesOffsetWithExtremelyLargeTimeout",
			cpuMsrPath: "testdata/cpu-msr/0/msr",
			offset:     0x4,
			timeout:    time.Hour,
			expected:   0x98badcfeefcdab89,
			err:        nil,
		},
		{
			name:       "8BytesOffsetWithExtremelyLargeTimeout",
			cpuMsrPath: "testdata/cpu-msr/0/msr",
			offset:     0x8,
			timeout:    time.Hour,
			expected:   0x1032547698badcfe,
			err:        nil,
		},
		{
			name:       "OutOfBoundsOffsetWithExtremelyLargeTimeout",
			cpuMsrPath: "testdata/cpu-msr/0/msr",
			offset:     0xb,
			timeout:    time.Hour,
			expected:   0,
			err:        errors.New("offset 0xb is out-of-bounds"),
		},
		{
			name:       "MsrFileNotExistsWithExtremelySmallTimeout",
			cpuMsrPath: "testdata/cpu-msr-cpuID-msr-not-exist/0/msr",
			offset:     0x0,
			timeout:    time.Nanosecond,
			expected:   0,
			err:        errors.New("open testdata/cpu-msr-cpuID-msr-not-exist/0/msr: no such file or directory"),
		},
		{
			name:       "4BytesOffsetWithExtremelySmallTimeout",
			cpuMsrPath: "testdata/cpu-msr/0/msr",
			offset:     0x4,
			timeout:    time.Nanosecond,
			expected:   0,
			err:        errors.New("timeout when reading file at offset 0x4"),
		},
		{
			name:       "8BytesOffsetWithExtremelySmallTimeout",
			cpuMsrPath: "testdata/cpu-msr/0/msr",
			offset:     0x8,
			timeout:    time.Nanosecond,
			expected:   0,
			err:        errors.New("timeout when reading file at offset 0x8"),
		},
		{
			name:       "OutOfBoundsOffsetWithExtremelySmallTimeout",
			cpuMsrPath: "testdata/cpu-msr/0/msr",
			offset:     0xb,
			timeout:    time.Nanosecond,
			expected:   0,
			err:        errors.New("timeout when reading file at offset 0xb"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := &msr{
				path:    tc.cpuMsrPath,
				timeout: tc.timeout,
			}

			out, err := m.read(tc.offset)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.expected, out)
		})
	}
}

func TestMsrReadAll(t *testing.T) {
	testCases := []struct {
		name       string
		cpuMsrPath string
		offsets    []uint32
		timeout    time.Duration
		expected   map[uint32]uint64
		err        error
	}{
		{
			name:       "NoError",
			cpuMsrPath: "testdata/cpu-msr/0/msr",
			offsets:    []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x08},
			expected: map[uint32]uint64{
				0x00: 0xefcdab8967452301,
				0x02: 0xdcfeefcdab896745,
				0x04: 0x98badcfeefcdab89,
				0x05: 0x7698badcfeefcdab,
				0x06: 0x547698badcfeefcd,
				0x08: 0x1032547698badcfe,
			},
			err: nil,
		},
		{
			name:       "OutOfBoundsMsrError",
			cpuMsrPath: "testdata/cpu-msr/0/msr",
			offsets:    []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x0b},
			expected:   nil,
			err:        errors.New("is out-of-bounds"),
		},
		{
			name:       "MsrFileNotExist",
			cpuMsrPath: "testdata/cpu-msr-cpuID-msr-not-exist/0/msr",
			offsets:    []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x0b},
			expected:   nil,
			err:        errors.New("open testdata/cpu-msr-cpuID-msr-not-exist/0/msr: no such file or directory"),
		},
		{
			name:       "NoErrorWithExtremelyLargeTimeout",
			cpuMsrPath: "testdata/cpu-msr/0/msr",
			offsets:    []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x08},
			timeout:    time.Hour,
			expected: map[uint32]uint64{
				0x00: 0xefcdab8967452301,
				0x02: 0xdcfeefcdab896745,
				0x04: 0x98badcfeefcdab89,
				0x05: 0x7698badcfeefcdab,
				0x06: 0x547698badcfeefcd,
				0x08: 0x1032547698badcfe,
			},
			err: nil,
		},
		{
			name:       "OutOfBoundsMsrErrorWithExtremelyLargeTimeout",
			cpuMsrPath: "testdata/cpu-msr/0/msr",
			offsets:    []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x0b},
			timeout:    time.Hour,
			expected:   nil,
			err:        errors.New("is out-of-bounds"),
		},
		{
			name:       "NoError",
			cpuMsrPath: "testdata/cpu-msr/0/msr",
			offsets:    []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x08},
			timeout:    time.Nanosecond,
			expected:   nil,
			err:        errors.New("error reading MSR offsets: timeout when reading file at offset"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := &msr{
				path:    tc.cpuMsrPath,
				timeout: tc.timeout,
			}

			out, err := m.readAll(tc.offsets)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.expected, out)
		})
	}
}

func TestNewMsrWithStorage(t *testing.T) {
	t.Run("WithoutOffsets", func(t *testing.T) {
		reqOffsets := []uint32{}
		cpuMsrDir := "testdata/cpu-msr/0"

		m, err := newMsrWithStorage(cpuMsrDir, reqOffsets, 0)
		require.Nil(t, m)
		require.ErrorContains(t, err, "no offsets were provided")
	})

	t.Run("WithOffsets", func(t *testing.T) {
		reqOffsets := []uint32{c3Residency, c6Residency, c7Residency, maxFreqClockCount, actualFreqClockCount, timestampCounter}
		cpuMsrDir := "testdata/cpu-msr/0"
		cpuMsrFile := filepath.Join(cpuMsrDir, msrFile)

		expected := &msrWithStorage{
			msrReg: &msr{
				path:  cpuMsrFile,
				cpuID: 0,
			},
			offsets:      reqOffsets,
			offsetValues: map[uint32]uint64{},
			offsetDeltas: map[uint32]uint64{},
		}

		m, err := newMsrWithStorage(cpuMsrDir, reqOffsets, 0)
		require.NoError(t, err)
		require.Equal(t, expected, m)
	})
}

func TestMsrWithStorageGetters(t *testing.T) {
	mReg, err := newMsr("testdata/cpu-msr/0", 0)
	require.NoError(t, err)

	expectedValues := map[uint32]uint64{
		0x00: 2,
		0x01: 2,
		0x02: 2,
		0x03: 2,
		0x04: 2,
		0x05: 2,
	}

	expectedDeltas := map[uint32]uint64{
		0x00: 1,
		0x01: 1,
		0x02: 1,
		0x03: 1,
		0x04: 1,
		0x05: 1,
	}

	expectedTimestampDelta := 10 * time.Second

	m := &msrWithStorage{
		msrReg:         mReg,
		timestampDelta: expectedTimestampDelta,
		offsetValues:   expectedValues,
		offsetDeltas:   expectedDeltas,
	}
	require.Equal(t, expectedValues, m.getOffsetValues())
	require.Equal(t, expectedDeltas, m.getOffsetDeltas())
	require.Equal(t, expectedTimestampDelta, m.getTimestampDelta())
}

func TestMsrWithStorageSetters(t *testing.T) {
	mReg, err := newMsr("testdata/cpu-msr/0", 0)
	require.NoError(t, err)

	m := &msrWithStorage{
		msrReg:       mReg,
		offsetValues: nil,
		offsetDeltas: nil,
	}

	expectedValues := map[uint32]uint64{
		0x00: 2,
		0x01: 2,
		0x02: 2,
		0x03: 2,
		0x04: 2,
		0x05: 2,
	}

	expectedDeltas := map[uint32]uint64{
		0x00: 1,
		0x01: 1,
		0x02: 1,
		0x03: 1,
		0x04: 1,
		0x05: 1,
	}

	m.setOffsetValues(expectedValues)
	m.setOffsetDeltas(expectedDeltas)
	require.Equal(t, expectedValues, m.getOffsetValues())
	require.Equal(t, expectedDeltas, m.getOffsetDeltas())
}

type msrTimeSensitiveSuite struct {
	suite.Suite
}

func (s *msrTimeSensitiveSuite) SetupTest() {
	setFakeClock()
	fakeClock.Set(time.Now())
}

func (s *msrTimeSensitiveSuite) TearDownTest() {
	unsetFakeClock()
}

func (s *msrTimeSensitiveSuite) TestMsrWithStorageUpdate() {
	s.Run("NegativeOffsetDelta", func() {
		msrOffsets := []uint32{0x00, 0x02, 0x04}

		offsetValuesT1 := map[uint32]uint64{
			msrOffsets[0]: uint64(0xefcdab8967452301),
			msrOffsets[1]: uint64(0xdcfeefcdab896745),
			msrOffsets[2]: uint64(0x98badcfeefcdab89),
		}

		offsetValuesT2 := map[uint32]uint64{
			msrOffsets[0]: uint64(0xefcdab8967452300),
			msrOffsets[1]: uint64(0xdcfeefcdab895745),
			msrOffsets[2]: uint64(0x98badcfeefcdab89),
		}

		expectedDeltas := map[uint32]uint64{
			msrOffsets[0]: 0,
			msrOffsets[1]: 0,
			msrOffsets[2]: 0,
		}

		mMsrReg := &msrRegMock{}

		// mock getting CPU ID for msr register
		mMsrReg.On("getCPUID").Return(1).Twice()

		// mock reading all msr offset values at t1
		mMsrReg.On("readAll", msrOffsets).Return(offsetValuesT1, nil).Once()

		// mock reading all msr offset values at t2
		mMsrReg.On("readAll", msrOffsets).Return(offsetValuesT2, nil).Once()

		m := &msrWithStorage{
			msrReg: mMsrReg,

			offsets:      msrOffsets,
			offsetValues: map[uint32]uint64{},
			offsetDeltas: map[uint32]uint64{},
			timestamp:    fakeClock.Now(),
		}

		// update at t1
		d1 := 10 * time.Second
		fakeClock.Add(d1)

		s.Require().NoError(m.update())
		s.Require().Equal(offsetValuesT1, m.getOffsetValues())
		s.Require().Equal(offsetValuesT1, m.getOffsetDeltas())
		s.Require().Equal(d1, m.getTimestampDelta())

		// update at t2
		d2 := 5 * time.Second
		fakeClock.Add(d2)

		s.Require().NoError(m.update())
		s.Require().Equal(offsetValuesT2, m.getOffsetValues())
		s.Require().Equal(expectedDeltas, m.getOffsetDeltas())
		s.Require().Equal(d2, m.getTimestampDelta())
	})

	s.Run("PositiveOffsetDeltas", func() {
		mReg, err := newMsr("testdata/cpu-msr/0", 0)
		s.Require().NoError(err)

		m := &msrWithStorage{
			msrReg:       mReg,
			offsets:      []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x08},
			offsetValues: map[uint32]uint64{},
			offsetDeltas: map[uint32]uint64{},
			timestamp:    fakeClock.Now(),
		}

		expectedValues := map[uint32]uint64{
			0x00: uint64(0xefcdab8967452301),
			0x02: uint64(0xdcfeefcdab896745),
			0x04: uint64(0x98badcfeefcdab89),
			0x05: uint64(0x7698badcfeefcdab),
			0x06: uint64(0x547698badcfeefcd),
			0x08: uint64(0x1032547698badcfe),
		}

		d1 := 10 * time.Second
		fakeClock.Add(d1)

		s.Require().NoError(m.update())
		s.Require().Equal(expectedValues, m.getOffsetValues())
		s.Require().Equal(expectedValues, m.getOffsetDeltas())
		s.Require().Equal(d1, m.getTimestampDelta())

		expectedDeltas := map[uint32]uint64{
			0x00: 0,
			0x02: 0,
			0x04: 0,
			0x05: 0,
			0x06: 0,
			0x08: 0,
		}

		d2 := 5 * time.Second
		fakeClock.Add(d2)

		s.Require().NoError(m.update())
		s.Require().Equal(expectedValues, m.getOffsetValues())
		s.Require().Equal(expectedDeltas, m.getOffsetDeltas())
		s.Require().Equal(d2, m.getTimestampDelta())
	})
}

func TestMsrTimeSensitive(t *testing.T) {
	suite.Run(t, new(msrTimeSensitiveSuite))
}

func (s *msrTimeSensitiveSuite) TestInitMsrMap() {
	testCases := []struct {
		name    string
		msrPath string
		offsets []uint32
		cpuIDs  []int
		msrMap  map[int]msrRegWithStorage
		err     error
	}{
		{
			name:    "OffsetsNotProvided",
			offsets: []uint32{},
			msrMap:  nil,
			err:     errors.New("MSR offsets argument cannot be empty"),
		},
		{
			name:    "InvalidMsrBasePath",
			msrPath: "",
			offsets: []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x08},
			msrMap:  nil,
			err:     errors.New("base path for MSR files cannot be an empty string"),
		},
		{
			name:    "MsrBasePathNotExist",
			msrPath: "/dummy/path",
			offsets: []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x08},
			msrMap:  nil,
			err:     errors.New("invalid MSR base path \"/dummy/path\""),
		},
		{
			name:    "MsrBasePathCPUDirectoriesNotExist",
			msrPath: "testdata/cpu-msr-directories-not-exist",
			offsets: []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x08},
			msrMap:  nil,
			err:     errors.New("could not find valid CPU MSR files for path: \"testdata/cpu-msr-directories-not-exist\""),
		},
		{
			name:    "MsrBasePathIsNotADir",
			msrPath: "testdata/cpu-msr/0/msr",
			offsets: []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x08},
			msrMap:  nil,
			err:     errors.New("error reading directory \"testdata/cpu-msr/0/msr\""),
		},
		{
			name:    "MsrBasePathDirectoriesCPUIDNotFound",
			msrPath: "testdata/cpu-msr-invalid-cpuID-directories",
			offsets: []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x08},
			msrMap:  nil,
			err:     errors.New("could not find valid CPU MSR files for path: \"testdata/cpu-msr-invalid-cpuID-directories\""),
		},
		{
			name:    "MsrPathWithMissingCPUIDMsrFile",
			msrPath: "testdata/cpu-msr-cpuID-msr-not-exist",
			offsets: []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x08},
			msrMap:  nil,
			err: errors.New("error creating MSR register for CPU path \"testdata/cpu-msr-cpuID-msr-not-exist/0\": invalid MSR file for cpu ID 0: " +
				"file \"testdata/cpu-msr-cpuID-msr-not-exist/0/msr\" does not exist"),
		},
		{
			name:    "MsrPathWithCPUIDMsrSymlink",
			msrPath: "testdata/cpu-msr-cpuID-msr-softlink",
			offsets: []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x08},
			msrMap:  nil,
			err: errors.New("error creating MSR register for CPU path \"testdata/cpu-msr-cpuID-msr-softlink/1\": invalid MSR file for cpu ID 1: " +
				"file \"testdata/cpu-msr-cpuID-msr-softlink/1/msr\" is a symlink"),
		},
		{
			name:    "OffsetoutOfBounds",
			msrPath: "testdata/cpu-msr",
			offsets: []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x0b},
			msrMap:  nil,
			err: errors.New("error initializing the MSR register storage for CPU ID 0: " +
				"error reading MSR offsets: offset 0xb is out-of-bounds"),
		},
		{
			name:    "MsrPathValid",
			msrPath: "testdata/cpu-msr",
			offsets: []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x08},
			msrMap: map[int]msrRegWithStorage{
				0: &msrWithStorage{
					msrReg: &msr{
						cpuID: 0,
						path:  "testdata/cpu-msr/0/msr",
					},
					offsets:        []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x08},
					timestamp:      fakeClock.Now(),
					timestampDelta: fakeClock.Now().Sub(time.Time{}),
					offsetValues: map[uint32]uint64{
						0x00: uint64(0xefcdab8967452301),
						0x02: uint64(0xdcfeefcdab896745),
						0x04: uint64(0x98badcfeefcdab89),
						0x05: uint64(0x7698badcfeefcdab),
						0x06: uint64(0x547698badcfeefcd),
						0x08: uint64(0x1032547698badcfe),
					},
					offsetDeltas: map[uint32]uint64{
						0x00: uint64(0xefcdab8967452301),
						0x02: uint64(0xdcfeefcdab896745),
						0x04: uint64(0x98badcfeefcdab89),
						0x05: uint64(0x7698badcfeefcdab),
						0x06: uint64(0x547698badcfeefcd),
						0x08: uint64(0x1032547698badcfe),
					},
				},
				1: &msrWithStorage{
					msrReg: &msr{
						cpuID: 1,
						path:  "testdata/cpu-msr/1/msr",
					},
					offsets: []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x08},
					offsetValues: map[uint32]uint64{
						0x00: uint64(0xefcdab8967452301),
						0x02: uint64(0xdcfeefcdab896745),
						0x04: uint64(0x98badcfeefcdab89),
						0x05: uint64(0x7698badcfeefcdab),
						0x06: uint64(0x547698badcfeefcd),
						0x08: uint64(0x1032547698badcfe),
					},
					offsetDeltas: map[uint32]uint64{
						0x00: uint64(0xefcdab8967452301),
						0x02: uint64(0xdcfeefcdab896745),
						0x04: uint64(0x98badcfeefcdab89),
						0x05: uint64(0x7698badcfeefcdab),
						0x06: uint64(0x547698badcfeefcd),
						0x08: uint64(0x1032547698badcfe),
					},
					timestamp:      fakeClock.Now(),
					timestampDelta: fakeClock.Now().Sub(time.Time{}),
				},
				10: &msrWithStorage{
					msrReg: &msr{
						cpuID: 10,
						path:  "testdata/cpu-msr/10/msr",
					},
					offsets: []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x08},
					offsetValues: map[uint32]uint64{
						0x00: uint64(0xefcdab8967452301),
						0x02: uint64(0xdcfeefcdab896745),
						0x04: uint64(0x98badcfeefcdab89),
						0x05: uint64(0x7698badcfeefcdab),
						0x06: uint64(0x547698badcfeefcd),
						0x08: uint64(0x1032547698badcfe),
					},
					offsetDeltas: map[uint32]uint64{
						0x00: uint64(0xefcdab8967452301),
						0x02: uint64(0xdcfeefcdab896745),
						0x04: uint64(0x98badcfeefcdab89),
						0x05: uint64(0x7698badcfeefcdab),
						0x06: uint64(0x547698badcfeefcd),
						0x08: uint64(0x1032547698badcfe),
					},
					timestamp:      fakeClock.Now(),
					timestampDelta: fakeClock.Now().Sub(time.Time{}),
				},
				100: &msrWithStorage{
					msrReg: &msr{
						cpuID: 100,
						path:  "testdata/cpu-msr/100/msr",
					},
					offsets: []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x08},
					offsetValues: map[uint32]uint64{
						0x00: uint64(0xefcdab8967452301),
						0x02: uint64(0xdcfeefcdab896745),
						0x04: uint64(0x98badcfeefcdab89),
						0x05: uint64(0x7698badcfeefcdab),
						0x06: uint64(0x547698badcfeefcd),
						0x08: uint64(0x1032547698badcfe),
					},
					offsetDeltas: map[uint32]uint64{
						0x00: uint64(0xefcdab8967452301),
						0x02: uint64(0xdcfeefcdab896745),
						0x04: uint64(0x98badcfeefcdab89),
						0x05: uint64(0x7698badcfeefcdab),
						0x06: uint64(0x547698badcfeefcd),
						0x08: uint64(0x1032547698badcfe),
					},
					timestamp:      fakeClock.Now(),
					timestampDelta: fakeClock.Now().Sub(time.Time{}),
				},
			},
			err: nil,
		},
		{
			name:    "MsrPathValidSpecified_0_10",
			msrPath: "testdata/cpu-msr",
			offsets: []uint32{0x00, 0x02},
			cpuIDs:  []int{0, 10},
			msrMap: map[int]msrRegWithStorage{
				0: &msrWithStorage{
					msrReg: &msr{
						cpuID: 0,
						path:  "testdata/cpu-msr/0/msr",
					},
					offsets: []uint32{0x00, 0x02},
					offsetValues: map[uint32]uint64{
						0x00: uint64(0xefcdab8967452301),
						0x02: uint64(0xdcfeefcdab896745),
					},
					offsetDeltas: map[uint32]uint64{
						0x00: uint64(0xefcdab8967452301),
						0x02: uint64(0xdcfeefcdab896745),
					},
					timestamp:      fakeClock.Now(),
					timestampDelta: fakeClock.Now().Sub(time.Time{}),
				},
				10: &msrWithStorage{
					msrReg: &msr{
						cpuID: 10,
						path:  "testdata/cpu-msr/10/msr",
					},
					offsets: []uint32{0x00, 0x02},
					offsetValues: map[uint32]uint64{
						0x00: uint64(0xefcdab8967452301),
						0x02: uint64(0xdcfeefcdab896745),
					},
					offsetDeltas: map[uint32]uint64{
						0x00: uint64(0xefcdab8967452301),
						0x02: uint64(0xdcfeefcdab896745),
					},
					timestamp:      fakeClock.Now(),
					timestampDelta: fakeClock.Now().Sub(time.Time{}),
				},
			},
			err: nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			m := &msrDataWithStorage{
				msrPath:    tc.msrPath,
				msrOffsets: tc.offsets,
			}

			err := m.initMsrMap(tc.cpuIDs, 0)
			s.Require().Equal(tc.msrMap, m.msrMap)
			if tc.err != nil {
				s.Require().ErrorContains(err, tc.err.Error())
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func TestMsrDataWithStorageRead(t *testing.T) {
	testCases := []struct {
		name     string
		msrPath  string
		cpuID    int
		offset   uint32
		expected uint64
		err      error
	}{
		{
			name:     "InvalidCPUID",
			msrPath:  "testdata/cpu-msr",
			cpuID:    2,
			expected: 0,
			err:      errors.New("could not find MSR register for CPU ID: 2"),
		},
		{
			name:     "ZeroBytesOffset",
			msrPath:  "testdata/cpu-msr",
			cpuID:    0,
			offset:   0x0,
			expected: uint64(0xefcdab8967452301),
			err:      nil,
		},
		{
			name:     "4BytesOffset",
			msrPath:  "testdata/cpu-msr",
			cpuID:    0,
			offset:   0x4,
			expected: uint64(0x98badcfeefcdab89),
			err:      nil,
		},
		{
			name:     "8BytesOffset",
			msrPath:  "testdata/cpu-msr",
			cpuID:    0,
			offset:   0x8,
			expected: uint64(0x1032547698badcfe),
			err:      nil,
		},
		{
			name:     "OutOfBoundsOffset",
			msrPath:  "testdata/cpu-msr",
			cpuID:    0,
			offset:   0xb,
			expected: uint64(0),
			err:      errors.New("offset 0xb is out-of-bounds"),
		},
	}

	mReg, err := newMsr("testdata/cpu-msr/0", 0)
	require.NoError(t, err)
	m := &msrDataWithStorage{
		msrMap: map[int]msrRegWithStorage{
			0: &msrWithStorage{
				msrReg: mReg,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, errRead := m.read(tc.offset, tc.cpuID)
			require.Equal(t, tc.expected, out)
			if tc.err != nil {
				require.ErrorContains(t, errRead, tc.err.Error())
			} else {
				require.NoError(t, errRead)
			}
		})
	}
}

func TestMsrDataWithStorageIsMsrLoaded(t *testing.T) {
	testCases := []struct {
		desc     string
		filePath string
		expected bool
		err      error
	}{
		{
			desc:     "EmptyFilename",
			filePath: "",
			expected: false,
			err:      errors.New("file path is empty"),
		},
		{
			desc:     "FileNotExist",
			filePath: "dummy_proc_modules_file",
			expected: false,
			err:      errors.New("file \"dummy_proc_modules_file\" does not exist"),
		},
		{
			desc:     "PathIsADir",
			filePath: "testdata",
			expected: false,
			err:      errors.New("could not read file \"testdata\": read testdata: is a directory"),
		},
		{
			desc:     "Symlink",
			filePath: "testdata/symlink",
			expected: false,
			err:      errors.New("file \"testdata/symlink\" is a symlink"),
		},
		{
			desc:     "NotLoaded",
			filePath: "testdata/proc_modules_msr_not_loaded",
			expected: false,
			err:      nil,
		},
		{
			desc:     "Loaded",
			filePath: "testdata/proc_modules_msr_loaded",
			expected: true,
			err:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			m := &msrDataWithStorage{}
			out, err := m.isMsrLoaded(tc.filePath)
			require.Equal(t, tc.expected, out)
			if tc.err != nil {
				require.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsrDataWithStorageUpdate(t *testing.T) {
	testCases := []struct {
		name                 string
		msrPath              string
		cpuID                int
		offsets              []uint32
		expectedMsrInfo      map[uint32]uint64
		expectedMsrInfoDelta map[uint32]uint64
		err                  error
	}{
		{
			name:    "InvalidCPUID",
			msrPath: "testdata/cpu-msr",
			cpuID:   2,
			err:     errors.New("could not find MSR register for CPU ID: 2"),
		},
		{
			name:    "Valid",
			msrPath: "testdata/cpu-msr",
			cpuID:   0,
			offsets: []uint32{0x00, 0x02, 0x04, 0x05, 0x06, 0x08},
			expectedMsrInfo: map[uint32]uint64{
				0x00: uint64(0xefcdab8967452301),
				0x02: uint64(0xdcfeefcdab896745),
				0x04: uint64(0x98badcfeefcdab89),
				0x05: uint64(0x7698badcfeefcdab),
				0x06: uint64(0x547698badcfeefcd),
				0x08: uint64(0x1032547698badcfe),
			},
			expectedMsrInfoDelta: map[uint32]uint64{
				0x00: uint64(0xefcdab8967452301),
				0x02: uint64(0xdcfeefcdab896745),
				0x04: uint64(0x98badcfeefcdab89),
				0x05: uint64(0x7698badcfeefcdab),
				0x06: uint64(0x547698badcfeefcd),
				0x08: uint64(0x1032547698badcfe),
			},
			err: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := &msrDataWithStorage{
				msrMap: map[int]msrRegWithStorage{
					0: &msrWithStorage{
						msrReg: &msr{
							path:  "testdata/cpu-msr/0/msr",
							cpuID: 0,
						},
						offsets:      tc.offsets,
						offsetValues: map[uint32]uint64{},
						offsetDeltas: map[uint32]uint64{},
					},
				},
			}

			err := m.update(tc.cpuID)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedMsrInfo, m.msrMap[tc.cpuID].getOffsetValues())
				require.Equal(t, tc.expectedMsrInfoDelta, m.msrMap[tc.cpuID].getOffsetDeltas())
			}
		})
	}
}

func TestMsrDataWithStorageGetOffsetDeltas(t *testing.T) {
	m := &msrDataWithStorage{
		msrMap: map[int]msrRegWithStorage{
			0: &msrWithStorage{
				msrReg: &msr{
					path:  "testdata/cpu-msr/0/msr",
					cpuID: 0,
				},
				offsets:      []uint32{0x00, 0x02, 0x04},
				offsetValues: map[uint32]uint64{},
				offsetDeltas: map[uint32]uint64{},
			},
		},
	}

	t.Run("InvalidCPUID", func(t *testing.T) {
		cpuID := 2
		deltas, err := m.getOffsetDeltas(cpuID)
		require.Nil(t, deltas)
		require.ErrorContains(t, err, fmt.Sprintf("could not find MSR register for CPU ID: %v", cpuID))
	})

	t.Run("WithoutUpdate", func(t *testing.T) {
		cpuID := 0
		deltasExp := map[uint32]uint64{}
		deltasOut, err := m.getOffsetDeltas(cpuID)
		require.NoError(t, err)
		require.Equal(t, deltasExp, deltasOut)
	})

	t.Run("WithUpdate", func(t *testing.T) {
		cpuID := 0

		require.NoError(t, m.update(cpuID))

		deltasExp := map[uint32]uint64{
			0x00: uint64(0xefcdab8967452301),
			0x02: uint64(0xdcfeefcdab896745),
			0x04: uint64(0x98badcfeefcdab89),
		}
		deltasOut, err := m.getOffsetDeltas(cpuID)
		require.NoError(t, err)
		require.Equal(t, deltasExp, deltasOut)

		require.NoError(t, m.update(cpuID))
		deltasExp = map[uint32]uint64{
			0x00: 0,
			0x02: 0,
			0x04: 0,
		}
		deltasOut, err = m.getOffsetDeltas(cpuID)
		require.NoError(t, err)
		require.Equal(t, deltasExp, deltasOut)
	})
}

func TestMsrDataWithStorageScaleOffsetDeltas(t *testing.T) {
	testCases := []struct {
		name            string
		cpuID           int
		offsetDeltasMap map[uint32]uint64
		num             uint64
		denom           uint64

		scaledOffsetDeltasMap map[uint32]uint64
		err                   error
	}{
		{
			name:  "CPUIDNotFound",
			cpuID: 100,
			offsetDeltasMap: map[uint32]uint64{
				0x00: 15000,
			},
			num:   10,
			denom: 100,
			err:   errors.New("could not find MSR register for CPU ID: 100"),
		},
		{
			name: "ScalingFactorCloseToOne",
			offsetDeltasMap: map[uint32]uint64{
				0x01: 100000,
			},
			num:   math.MaxUint64 - 1000,
			denom: math.MaxUint64,
			scaledOffsetDeltasMap: map[uint32]uint64{
				0x01: 99999,
			},
		},
		{
			name: "ScalingFactorCloseToOneAndBigOffsetDeltas",
			offsetDeltasMap: map[uint32]uint64{
				0x01: math.MaxUint64,
			},
			num:   math.MaxUint64 - 1,
			denom: math.MaxUint64,
			scaledOffsetDeltasMap: map[uint32]uint64{
				0x01: math.MaxUint64 - 1,
			},
		},
		{
			name: "SmallScalingFactor",
			offsetDeltasMap: map[uint32]uint64{
				0x01: 10000000000,
			},
			num:   1,
			denom: 100000000,
			scaledOffsetDeltasMap: map[uint32]uint64{
				0x01: 100,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msrOffsets := make([]uint32, 0, len(tc.offsetDeltasMap))
			for offset := range tc.offsetDeltasMap {
				msrOffsets = append(msrOffsets, offset)
			}

			m := &msrDataWithStorage{
				msrMap: map[int]msrRegWithStorage{
					0: &msrWithStorage{
						offsets:      msrOffsets,
						offsetDeltas: tc.offsetDeltasMap,
					},
				},
			}

			numBig := new(big.Float).SetUint64(tc.num)
			denomBig := new(big.Float).SetUint64(tc.denom)
			fBig := new(big.Float).Quo(numBig, denomBig)

			err := m.scaleOffsetDeltas(tc.cpuID, msrOffsets, fBig)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)

				offsetDeltasOutMap, err := m.getOffsetDeltas(tc.cpuID)
				require.NoError(t, err)
				require.Equal(t, tc.scaledOffsetDeltasMap, offsetDeltasOutMap)
			}
		})
	}
}

func (s *msrTimeSensitiveSuite) TestMsrDataWithStorageGetTimestampDelta() {
	m := &msrDataWithStorage{
		msrMap: map[int]msrRegWithStorage{
			0: &msrWithStorage{
				msrReg: &msr{
					path:  "testdata/cpu-msr/0/msr",
					cpuID: 0,
				},
				offsets:      []uint32{0x00, 0x02, 0x04},
				offsetValues: map[uint32]uint64{},
				offsetDeltas: map[uint32]uint64{},
			},
		},
	}

	s.Run("InvalidCPUID", func() {
		cpuID := 2
		tsDeltaExp := time.Duration(0)
		tsDeltaOut, err := m.getTimestampDelta(cpuID)
		s.Require().Equal(tsDeltaExp, tsDeltaOut)
		s.Require().ErrorContains(err, fmt.Sprintf("could not find MSR register for CPU ID: %v", cpuID))
	})

	s.Run("WithoutUpdate", func() {
		cpuID := 0
		tsDeltaExp := time.Duration(0)
		tsDeltaOut, err := m.getTimestampDelta(cpuID)
		s.Require().NoError(err)
		s.Require().Equal(tsDeltaExp, tsDeltaOut)
	})

	s.Run("WithUpdate", func() {
		cpuID := 0

		s.Require().NoError(m.update(cpuID))

		d := 10 * time.Second
		fakeClock.Add(d)

		s.Require().NoError(m.update(cpuID))

		tsDeltaOut, err := m.getTimestampDelta(cpuID)
		s.Require().NoError(err)
		s.Require().Equal(d, tsDeltaOut)
	})
}
