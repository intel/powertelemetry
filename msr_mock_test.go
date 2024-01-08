// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"math/big"
	"time"

	"github.com/stretchr/testify/mock"
)

// msrRegMock represents a mock for msr type. Implements msrReg interface.
type msrRegMock struct {
	mock.Mock
}

func (m *msrRegMock) getPath() string {
	args := m.Called()
	return args.String(0)
}

func (m *msrRegMock) getCPUID() int {
	args := m.Called()
	return args.Int(0)
}

func (m *msrRegMock) read(offset uint32) (uint64, error) {
	args := m.Called(offset)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *msrRegMock) readAll(offsets []uint32) (map[uint32]uint64, error) {
	args := m.Called(offsets)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uint32]uint64), args.Error(1)
}

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

func (m *msrMock) scaleOffsetDeltas(cpuID int, offsets []uint32, f *big.Float) error {
	args := m.Called(cpuID, offsets, f)
	return args.Error(0)
}

func (m *msrMock) getOffsetDeltas(cpuID int) (map[uint32]uint64, error) {
	args := m.Called(cpuID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uint32]uint64), args.Error(1)
}

func (m *msrMock) getTimestampDelta(cpuID int) (time.Duration, error) {
	args := m.Called(cpuID)
	return args.Get(0).(time.Duration), args.Error(1)
}
