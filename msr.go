// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/intel/powertelemetry/internal/log"
)

const (
	// base path comprising all per-CPU ID MSR files.
	defaultMsrBasePath = "/dev/cpu"

	// file name of the binary MSR file specific for each CPU ID.
	msrFile = "msr"
)

var (
	// regex used to check CPU ID format as numeric value without leading zeroes.
	cpuIDRegex = regexp.MustCompile("^(0|[1-9][0-9]*)$")

	// regex used to check MSR module within the loaded kernel modules list.
	msrModuleRegex = regexp.MustCompile(`\bmsr\b`)
)

// msrReg represents a CPU ID specific MSR register with the ability to read offset
// values.
type msrReg interface {
	// getPath gets the absolute path of the MSR file.
	getPath() string

	// getCPUID gets the CPU ID corresponding to the MSR register.
	getCPUID() int

	// read returns the MSR value of the given offset.
	read(offset uint32) (uint64, error)

	// readAll takes a slice of offsets and returns a map with offset key
	// and content of the MSR offset value.
	readAll(offsets []uint32) (map[uint32]uint64, error)
}

// msr represents a CPU ID specific MSR register. Implements msrReg interface.
type msr struct {
	path    string
	cpuID   int
	timeout time.Duration
}

// resultError is used for transmitting a value or an err through the channel.
type resultError struct {
	value uint64
	err   error
}

// newMsr creates a new MSR register, initializing the CPU ID for this specific
// register and the path where to find the MSR file.
func newMsr(path string, timeout time.Duration) (msrReg, error) {
	cpuIDStr := filepath.Base(path)
	if !cpuIDRegex.MatchString(cpuIDStr) {
		return nil, fmt.Errorf("invalid format for CPU ID in path %q", path)
	}
	cpuID, err := strconv.Atoi(cpuIDStr)
	if err != nil {
		return nil, fmt.Errorf("error converting parsed CPU ID from path to numeric")
	}
	cpuMsr := filepath.Join(path, msrFile)
	if err := checkFile(cpuMsr); err != nil {
		return nil, fmt.Errorf("invalid MSR file for cpu ID %v: %w", cpuID, err)
	}
	return &msr{
		path:    cpuMsr,
		cpuID:   cpuID,
		timeout: timeout,
	}, nil
}

// getPath returns the MSR file path of the receiver.
func (m *msr) getPath() string {
	return m.path
}

// getCPUID returns the CPU ID corresponding to the receiver.
func (m *msr) getCPUID() int {
	return m.cpuID
}

// read takes an address, specified as offset, and returns an 8-byte value with
// the address content of the given CPU ID's MSR.
func (m *msr) read(offset uint32) (uint64, error) {
	f, err := os.OpenFile(m.path, os.O_RDONLY, 0400)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return readOffset(offset, f, m.timeout)
}

// readAll takes a slice of addresses, specified as offsets, and returns a map
// with offset key and the offset content of the given CPU ID's MSR as value.
// Each read offset operation is performed in a separate goroutine. In case an
// error occurs, the function returns a nil map with the corresponding error.
func (m *msr) readAll(offsets []uint32) (map[uint32]uint64, error) {
	f, err := os.OpenFile(m.path, os.O_RDONLY, 0400)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	errCh := make(chan error)
	msrOffsetChannels := make(map[uint32]chan uint64)
	for _, offset := range offsets {
		msrOffsetChannels[offset] = make(chan uint64)
	}

	for offset, ch := range msrOffsetChannels {
		go func(off uint32, ch chan uint64, errCh chan error) {
			v, err := readOffset(off, f, m.timeout)
			if err != nil {
				errCh <- err
				return
			}
			ch <- v
		}(offset, ch, errCh)
	}

	valuesMap := make(map[uint32]uint64)
	for offset, ch := range msrOffsetChannels {
		select {
		case err := <-errCh:
			return nil, fmt.Errorf("error reading MSR offsets: %w", err)
		case v := <-ch:
			valuesMap[offset] = v
		}
	}
	return valuesMap, nil
}

// readOffset is a helper function that takes an address, specified as offset, an io.ReaderAt interface, and timeout.
// It returns an 8-byte value with the address content of the given reader argument.
func readOffset(offset uint32, reader io.ReaderAt, timeout time.Duration) (uint64, error) {
	// read without timeout
	if timeout <= 0 {
		return readOffsetWithoutTimeout(offset, reader)
	}

	// read with timeout
	resultCh := make(chan resultError, 1)
	go func(resCh chan<- resultError) {
		buf := make([]byte, 8)
		if _, err := reader.ReadAt(buf, int64(offset)); err != nil {
			if errors.Is(err, io.EOF) {
				resCh <- resultError{value: 0, err: fmt.Errorf("offset 0x%x is out-of-bounds", offset)}
			} else {
				resCh <- resultError{value: 0, err: fmt.Errorf("error when reading file at offset 0x%x: %w", offset, err)}
			}
		} else {
			resCh <- resultError{value: binary.LittleEndian.Uint64(buf), err: nil}
		}

		close(resCh)
	}(resultCh)

	t := time.NewTimer(timeout)
	select {
	case <-t.C:
		return 0, fmt.Errorf("timeout when reading file at offset 0x%x", offset)
	case result := <-resultCh:
		if !t.Stop() {
			<-t.C
		}
		if result.err != nil {
			return 0, result.err
		}

		return result.value, nil
	}
}

// readOffsetWithoutTimeout is a helper function that takes an address, specified as offset and an io.ReaderAt interface.
// It returns an 8-byte value with the address content of the given reader argument.
func readOffsetWithoutTimeout(offset uint32, reader io.ReaderAt) (uint64, error) {
	buf := make([]byte, 8)
	if _, err := reader.ReadAt(buf, int64(offset)); err != nil {
		if errors.Is(err, io.EOF) {
			return 0, fmt.Errorf("offset 0x%x is out-of-bounds", offset)
		}
		return 0, fmt.Errorf("error when reading file at offset 0x%x: %w", offset, err)
	}

	return binary.LittleEndian.Uint64(buf), nil
}

// msrRegWithStorage represents a CPU ID specific MSR register with the ability to read and
// store offset values. Two types of stored offset values are supported:
//   - offset values from the last read operation.
//   - delta offset values defined as the subtraction between offset values from last read
//     and offset values from the previous read operation.
type msrRegWithStorage interface {
	msrReg

	// getOffsetValues gets a map with offset key and offset value of the latest read operation.
	getOffsetValues() map[uint32]uint64

	// getOffsetDeltas gets a map with offset key and delta offset value between the latest and
	// the previous read operation.
	getOffsetDeltas() map[uint32]uint64

	// setOffsetDeltas sets the given map of offset key and delta offset values to the receiver.
	setOffsetDeltas(offsets map[uint32]uint64)

	// getTimestampDelta gets the timestamp delta between the last offset values reading operation
	// and its previous reading operation.
	getTimestampDelta() time.Duration

	// update gets MSR values and updates the storage.
	update() error
}

// msrWithStorage represents a CPU ID specific MSR register with the ability to read and
// store offset values. Implements msrRegWithStorage interface.
// The offset values in the storage correspond to values for offsets specified in offsets
// field.
type msrWithStorage struct {
	msrReg
	offsets        []uint32
	offsetValues   map[uint32]uint64 // offset values from the last read operation
	offsetDeltas   map[uint32]uint64 // delta offset values between the latest and its previous reading operation
	timestamp      time.Time         // timestamp of the last reading operation
	timestampDelta time.Duration     // timestamp delta between the last read and its previous reading operation
}

// newMsrWithStorage creates a new MSR register with the ability to read and store multiple MSR
// offset values, provided as argument. First creates an MSR register, then decorates it adding
// storage for both offset values from the last read operation and delta offset values between
// the latest and its previous reading operation.
func newMsrWithStorage(path string, offsets []uint32, timeout time.Duration) (msrRegWithStorage, error) {
	if len(offsets) == 0 {
		return nil, errors.New("no offsets were provided")
	}

	msr, err := newMsr(path, timeout)
	if err != nil {
		return nil, fmt.Errorf("error creating MSR register for CPU path %q: %w", path, err)
	}

	return &msrWithStorage{
		msrReg:       msr,
		offsets:      offsets,
		offsetValues: make(map[uint32]uint64),
		offsetDeltas: make(map[uint32]uint64),
	}, nil
}

// getOffsetValues returns a map with offset key and offset values from the last read operation
// of the receiver.
func (m *msrWithStorage) getOffsetValues() map[uint32]uint64 {
	return m.offsetValues
}

// setOffsetValues sets the given map of offset key and offset values to the receiver.
func (m *msrWithStorage) setOffsetValues(offsetValues map[uint32]uint64) {
	m.offsetValues = offsetValues
}

// getOffsetDeltas returns a map with offset key and delta offset values from the last read operation
// of the receiver.
func (m *msrWithStorage) getOffsetDeltas() map[uint32]uint64 {
	return m.offsetDeltas
}

// setOffsetDeltas sets the given map of offset key and delta offset values to the receiver.
func (m *msrWithStorage) setOffsetDeltas(offsetDeltas map[uint32]uint64) {
	m.offsetDeltas = offsetDeltas
}

// update performs reading operations along the offsets specified by the receiver. It updates
// last read offset values and delta offset values of the receiver.
func (m *msrWithStorage) update() error {
	latest, err := m.msrReg.readAll(m.offsets)
	if err != nil {
		return err
	}

	// Get time interval between offset MSR read and its previous reading operation
	newTimestamp := timeNowFn()
	m.timestampDelta, m.timestamp = newTimestamp.Sub(m.timestamp), newTimestamp

	prev := m.getOffsetValues()
	deltasMap := make(map[uint32]uint64, len(latest))
	for offset := range latest {
		if latest[offset] < prev[offset] {
			deltasMap[offset] = 0
			log.Warnf("A negative delta for the offset 0x%X and CPU ID %v", offset, m.msrReg.getCPUID())
		} else {
			deltasMap[offset] = latest[offset] - prev[offset]
		}
	}

	m.setOffsetDeltas(deltasMap)
	m.setOffsetValues(latest)
	return nil
}

// getTimestampDelta returns the timestamp delta between the offset values last reading operations
// and its previous reading operation.
func (m *msrWithStorage) getTimestampDelta() time.Duration {
	return m.timestampDelta
}

// msrReaderWithStorage represents per-CPU ID MSR registers of the host with the ability to read single
// MSR offset values, read and store multiple MSR offset values, and eventually provide the MSR delta
// offset values between latest and its previous reading operation.
type msrReaderWithStorage interface {
	// initMsrMap initializes a map of CPU ID key and MSR register value with storage.
	initMsrMap(cpuIDs []int, timeout time.Duration) error

	// isMsrLoaded check if MSR kernel module is loaded.
	isMsrLoaded(modulesPath string) (bool, error)

	// read returns the MSR value for a given offset and CPU ID.
	read(offset uint32, cpuID int) (uint64, error)

	// update takes a CPU ID, reads multiple MSR offset values and updates the storage.
	update(cpuID int) error

	// getOffsetDeltas takes a CPU ID and returns MSR delta offset values between latest and its previous reading operation.
	getOffsetDeltas(cpuID int) (map[uint32]uint64, error)

	// scaleOffsetDeltas takes a CPU ID and a slice of msr offsets. It scales all offset deltas of the msr storage by multiplying
	// each offset delta by the given factor f.
	scaleOffsetDeltas(cpuID int, offsets []uint32, f *big.Float) error

	// getTimestampDelta takes a CPU ID and returns the time interval between the last offset value reading operation
	// and its previous reading operation.
	getTimestampDelta(cpuID int) (time.Duration, error)
}

// msrDataWithStorage represents per-CPU ID MSR registers of the host with offset values storage capabilities.
//
// It represents the hierarchy tree for /dev/cpu directory:
//
// /dev/cpu
// ├── 0       (CPU ID)
// │   └── msr (MSR binary file)
// ├── 1
// │   └── msr
// ├── 2
// │   └── msr
// ∙
// └── n
//
// Each map entry corresponds to an MSR register, which allows reading operations
// to specific addresses (offsets), as well as storage of the offset values read.
type msrDataWithStorage struct {
	msrPath    string
	msrOffsets []uint32

	msrMap map[int]msrRegWithStorage
}

// initMsrMap initializes a map of CPU ID key and MSR register value with storage. Each MSR register is able to update
// the storage values and deltas of provided offsets in the offsets slice field. Field msrCPUIDs holds values of CPU IDs
// for which an MSR register is initialized. In case msrCPUIDs is nil, MSR registers for all CPU IDs found in system
// file will be initialized. It ensures that each CPU ID directory is properly formatted and binary MSR file exists.
// In case of malformed base path tree, an error is returned.
func (m *msrDataWithStorage) initMsrMap(cpuIDs []int, timeout time.Duration) error {
	if len(m.msrOffsets) == 0 {
		return errors.New("MSR offsets argument cannot be empty")
	}

	if len(m.msrPath) == 0 {
		return errors.New("base path for MSR files cannot be an empty string")
	}
	if err := checkFile(m.msrPath); err != nil {
		return fmt.Errorf("invalid MSR base path %q: %w", m.msrPath, err)
	}

	cpuDirs, err := os.ReadDir(m.msrPath)
	if err != nil {
		return fmt.Errorf("error reading directory %q: %w", m.msrPath, err)
	}

	// Declaring map for constant time search
	filterCPUIDsMap := make(map[string]struct{})
	for _, cpuID := range cpuIDs {
		cpuIDStr := strconv.FormatUint(uint64(cpuID), 10)
		filterCPUIDsMap[cpuIDStr] = struct{}{}
	}
	isFilterEmpty := len(filterCPUIDsMap) == 0

	msrMap := make(map[int]msrRegWithStorage)
	for _, cpuDirEntry := range cpuDirs {
		cpuDir := cpuDirEntry.Name()
		if !cpuDirEntry.IsDir() || !cpuIDRegex.MatchString(cpuDir) {
			continue
		}

		// Skip only if filterCPUIDs are not empty and there is no corresponding value in the map to cpuDir.
		// if filterCPUIDs are empty, then we read all values from the directory.
		if _, isCPUIDFiltered := filterCPUIDsMap[cpuDir]; !isFilterEmpty && !isCPUIDFiltered {
			continue
		}

		cpuPath := filepath.Join(m.msrPath, cpuDir)
		cpuMsrWithStorage, err := newMsrWithStorage(cpuPath, m.msrOffsets, timeout)
		if err != nil {
			return fmt.Errorf("error creating MSR register with storage for CPU path %q: %w", cpuPath, err)
		}

		err = cpuMsrWithStorage.update()
		if err != nil {
			return fmt.Errorf("error initializing the MSR register storage for CPU ID %v: %w", cpuMsrWithStorage.getCPUID(), err)
		}
		msrMap[cpuMsrWithStorage.getCPUID()] = cpuMsrWithStorage
	}

	if len(msrMap) == 0 {
		return fmt.Errorf("could not find valid CPU MSR files for path: %q", m.msrPath)
	}

	m.msrMap = msrMap
	return nil
}

// isMsrLoaded returns true if MSR kernel module is loaded, otherwise returns false.
func (m *msrDataWithStorage) isMsrLoaded(modulesPath string) (bool, error) {
	if err := checkFile(modulesPath); err != nil {
		return false, err
	}
	data, err := os.ReadFile(modulesPath)
	if err != nil {
		return false, fmt.Errorf("could not read file %q: %w", modulesPath, err)
	}
	matches := msrModuleRegex.FindAll(data, -1)
	return len(matches) > 0, nil
}

// read takes a CPU ID and offset and returns an 8-byte value with the contents
// of the associated MSR register.
func (m *msrDataWithStorage) read(offset uint32, cpuID int) (uint64, error) {
	reg, ok := m.msrMap[cpuID]
	if !ok {
		return 0, fmt.Errorf("could not find MSR register for CPU ID: %v", cpuID)
	}
	return reg.read(offset)
}

// update takes a CPU ID, performs reading operations along the offsets, storing the results
// within the storage.
func (m *msrDataWithStorage) update(cpuID int) error {
	reg, ok := m.msrMap[cpuID]
	if !ok {
		return fmt.Errorf("could not find MSR register for CPU ID: %v", cpuID)
	}
	return reg.update()
}

// getOffsetDeltas takes a CPU ID and returns a map with offset keys and delta offset values between
// latest and its previous reading offsets operation.
func (m *msrDataWithStorage) getOffsetDeltas(cpuID int) (map[uint32]uint64, error) {
	reg, ok := m.msrMap[cpuID]
	if !ok {
		return nil, fmt.Errorf("could not find MSR register for CPU ID: %v", cpuID)
	}
	return reg.getOffsetDeltas(), nil
}

// scaleOffsetDeltas takes a CPU ID and a slice of msr offsets. It scales all offset deltas of the msr storage by multiplying
// each offset delta by the given factor f.
func (m *msrDataWithStorage) scaleOffsetDeltas(cpuID int, offsets []uint32, f *big.Float) error {
	reg, ok := m.msrMap[cpuID]
	if !ok {
		return fmt.Errorf("could not find MSR register for CPU ID: %v", cpuID)
	}

	deltas := reg.getOffsetDeltas()
	for _, offset := range offsets {
		if v, ok := deltas[offset]; ok {
			deltaBig := new(big.Float).SetUint64(v)
			scaledDeltaBig := new(big.Float).Mul(deltaBig, f)
			scaledDelta, _ := scaledDeltaBig.Uint64()
			deltas[offset] = scaledDelta
		}
	}

	reg.setOffsetDeltas(deltas)
	return nil
}

// getTimestampDelta takes a CPU ID and returns the time interval between the last offset value reading
// operation and its previous reading operation.
func (m *msrDataWithStorage) getTimestampDelta(cpuID int) (time.Duration, error) {
	reg, ok := m.msrMap[cpuID]
	if !ok {
		return time.Duration(0), fmt.Errorf("could not find MSR register for CPU ID: %v", cpuID)
	}
	return reg.getTimestampDelta(), nil
}
