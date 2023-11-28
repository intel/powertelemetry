// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// base path comprising the uncore frequency type files.
	defaultUncoreFreqBasePath = "/sys/devices/system/cpu/intel_uncore_frequency"

	// pattern string to identify the specific package ID of an uncore frequency path.
	packageIDPattern = "package_%s"

	// pattern string to identify the specific die ID of an uncore frequency path.
	dieIDPattern = "die_%s"

	// pattern to identify file name of the uncore frequency type path.
	freqTypeFilePattern = "%s_freq_khz"
)

// uncoreFreqType is an enum type to identify specific uncore frequency parameters.
type uncoreFreqType int

// uncoreFreqType enum defines supported uncore frequency parameters.
const (
	initialMax uncoreFreqType = iota
	initialMin
	customizedMax
	customizedMin
	current
)

// Helper function to return a string representation of uncoreFreqType.
func (f uncoreFreqType) String() string {
	switch f {
	case initialMax:
		return "initial_max"
	case initialMin:
		return "initial_min"
	case customizedMax:
		return "max"
	case customizedMin:
		return "min"
	case current:
		return "current"
	default:
		return ""
	}
}

// Helper function takes a string representation of uncoreFreqType and returns its equivalent
// uncoreFreqType. If the string is not a valid frequency type, an error is returned.
func toUncoreFreqType(freqType string) (uncoreFreqType, error) {
	switch freqType {
	case "initial_max":
		return initialMax, nil
	case "initial_min":
		return initialMin, nil
	case "max":
		return customizedMax, nil
	case "min":
		return customizedMin, nil
	case "current":
		return current, nil
	default:
		return 0, fmt.Errorf("unsupported uncore frequency type %q", freqType)
	}
}

// getUncoreFreqPath is a helper function that takes a package ID, die ID, a frequency type and returns
// a string with the corresponding file name of the frequency type parameter.
func getUncoreFreqPath(packageID, dieID int, freqType string) (string, error) {
	uFreqType, err := toUncoreFreqType(freqType)
	if err != nil {
		return "", err
	}
	freqTypeFile := fmt.Sprintf(freqTypeFilePattern, uFreqType)

	// format packageID and dieID as two-digit strings
	packageIDPrefix := fmt.Sprintf(packageIDPattern, fmt.Sprintf("%02d", packageID))
	dieIDPrefix := fmt.Sprintf(dieIDPattern, fmt.Sprintf("%02d", dieID))
	prefix := fmt.Sprintf("%s_%s", packageIDPrefix, dieIDPrefix)

	return filepath.Join(prefix, freqTypeFile), nil
}

// uncoreFreqReader represents a mechanism for reading uncore frequency values exposed via filesystem.
type uncoreFreqReader interface {
	init() error

	// getUncoreFrequencyMhz takes a package ID, die ID and a frequency type and returns its value.
	getUncoreFrequencyMhz(packageID, dieID int, freqType string) (float64, error)
}

// uncoreFreqData allows to get uncore frequency values exposed via filesystem. Implements uncoreFreqReader interface.
type uncoreFreqData struct {
	uncoreFreqBasePath string
}

// init checks if uncoreFreqBasePath is a valid path.
// TODO: Consider to remove this method.
func (u *uncoreFreqData) init() error {
	if len(u.uncoreFreqBasePath) == 0 {
		return errors.New("base path of uncore frequency cannot be empty")
	}
	if err := checkFile(u.uncoreFreqBasePath); err != nil {
		return fmt.Errorf("invalid base path of uncore frequency: %w", err)
	}
	return nil
}

// getUncoreFrequencyMhz retrieves the uncore frequency value, in MHz, for the given package ID and die ID
// and the specified frequency type.
func (u *uncoreFreqData) getUncoreFrequencyMhz(packageID, dieID int, freqType string) (float64, error) {
	// Create the path to the frequency file based on the input parameters.
	freqPath, err := getUncoreFreqPath(packageID, dieID, freqType)
	if err != nil {
		return 0, fmt.Errorf("failed to get frequency path: %w", err)
	}

	// Read the contents of the frequency file.
	path := filepath.Join(u.uncoreFreqBasePath, freqPath)
	content, err := readFile(path)
	if err != nil {
		return 0, fmt.Errorf("failed to read frequency file: %w", err)
	}

	// Convert the file contents to a float64 value.
	freqKhz, err := strconv.ParseFloat(strings.TrimRight(string(content), "\n"), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert frequency file content to float64: %w", err)
	}
	return freqKhz * fromKiloHertzToMegaHertzRatio, nil
}
