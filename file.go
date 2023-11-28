// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"
)

// readFile reads the contents of a file at the given path and returns them as a byte slice.
// If the file doesn't exist or can't be read, an error is returned.
func readFile(filePath string) ([]byte, error) {
	// Check if the file exists and can be read.
	if err := checkFile(filePath); err != nil {
		return nil, err
	}

	// Read the entire contents of the file into a byte slice.
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error while reading file from path %q: %w", filePath, err)
	}

	return fileContent, nil
}

// readFileWithTimestamp reads the content of the given file specified as argument.
// If no error occurred, it returns a slice of bytes with file content and a timestamp.
// Otherwise, returns an error.
func readFileWithTimestamp(filePath string) ([]byte, time.Time, error) {
	// Check if the file exists and can be read.
	if err := checkFile(filePath); err != nil {
		return nil, time.Time{}, err
	}

	// Read the entire contents of the file into a byte slice.
	timestamp := timeNowFn()
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("error while reading file from path %q: %w", filePath, err)
	}
	return fileContent, timestamp, nil
}

// checkFile is a helper function that returns nil if the given file path exists,
// and it is not a symlink. Otherwise, it returns an error.
func checkFile(path string) error {
	if len(path) == 0 {
		return errors.New("file path is empty")
	}
	fInfo, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("file %q does not exist", path)
		}
		return fmt.Errorf("could not get the info for file %q: %w", path, err)
	}
	if fMode := fInfo.Mode(); fMode&os.ModeSymlink != 0 {
		return fmt.Errorf("file %q is a symlink", path)
	}
	return nil
}

// fileExists checks if a file exists at the given filePath.
// It returns true if the file exists, and false otherwise.
func fileExists(filePath string) (bool, error) {
	if len(filePath) == 0 {
		return false, errors.New("file path is empty")
	}
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err), nil
}
