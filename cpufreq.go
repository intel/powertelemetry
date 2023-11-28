// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// Path to folder with a collection of both global and individual CPU attributes.
	defaultCPUFreqBasePath = "/sys/devices/system/cpu"

	// Path to a file which frequency provides the current operating frequency of the CPU.
	cpuFrequencyPath = "cpufreq/scaling_cur_freq"
)

// cpuFreqReader represents a mechanism for reading core frequency values exposed via filesystem.
type cpuFreqReader interface {
	init() error

	getCPUFrequencyMhz(cpuID int) (float64, error)
}

// cpuFreqData allows to get core frequency values exposed via filesystem. Implements cpuFreqReader interface.
type cpuFreqData struct {
	cpuFrequencyFilePath string
}

// getCPUFrequencyMhz returns CPU's current frequency read from a file.
func (c *cpuFreqData) getCPUFrequencyMhz(cpuID int) (float64, error) {
	cpuFrequencyFile := c.getCPUFrequencyFilePath(cpuID)

	fileContent, err := readFile(cpuFrequencyFile)
	if err != nil {
		return 0, fmt.Errorf("error reading file %q: %w", cpuFrequencyFile, err)
	}

	cpuFrequency, err := strconv.ParseFloat(strings.TrimRight(string(fileContent), "\n"), 64)
	if err != nil {
		return 0, fmt.Errorf("error while converting value from file %q: %w", cpuFrequencyFile, err)
	}
	return cpuFrequency * fromKiloHertzToMegaHertzRatio, nil
}

// init checks if cpuFrequencyFilePath is a valid path.
// TODO: Consider to remove this method.
func (c *cpuFreqData) init() error {
	if len(c.cpuFrequencyFilePath) == 0 {
		return fmt.Errorf("base path of CPU core frequency cannot be empty")
	}
	if err := checkFile(c.cpuFrequencyFilePath); err != nil {
		return fmt.Errorf("invalid base path of CPU core frequency: %w", err)
	}
	return nil
}

// getCPUFrequencyFilePath returns the file path, from which the CPU's current frequency can be read.
func (c *cpuFreqData) getCPUFrequencyFilePath(cpuID int) string {
	cpuFrequencyFilePath := filepath.Join(c.cpuFrequencyFilePath, "cpu%d", cpuFrequencyPath)
	return fmt.Sprintf(cpuFrequencyFilePath, cpuID)
}
