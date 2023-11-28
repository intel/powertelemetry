// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain rolls out testdata in a temporary directory, runs all tests, and cleans up.
func TestMain(m *testing.M) {
	err := setupTestData()
	if err == nil {
		m.Run()
	}

	teardownTestData()

	if err != nil {
		fmt.Printf("TestData setup failed: %v\n", err)
		os.Exit(1)
	}
}

var tempTestDataDir string

// makeTestDataPath returns an absolute path created from the temporary directory and
// the relative path.
func makeTestDataPath(rel string) string {
	return filepath.Join(tempTestDataDir, rel)
}

// setupTestData copies all testdata directory structure and files to a temporary directory.
// While copying, files and directories are renamed to replace a colon placeholder "{colon}"
// with a regular colon character.
//
// E.g. the directory "intel-rapl{colon}0{colon}0" will be renamed to "intel-rapl:0:0".
//
// This workaround eliminates the issue with the ZIP archiver, which is used by the Go
// toolchain when importing a library. The ZIP archiver doesn't allow several characters
// including colon to be used in file or directory names.
//
// The workaround requires files and directories in the testdata directory to be named
// the way that colon characters are replaced with a "{colon}" placeholder.
//
// E.g. an "intel-rapl:0:0" directory should be named "intel-rapl{colon}0{colon}0".
//
// TODO: this whole implementation of the ZIP archiver workaround has to be moved under
// the rapl package along with rapl related unit tests when the powertelemetry library architecture
// is reworked to be package based.
func setupTestData() error {
	var err error
	tempTestDataDir, err = os.MkdirTemp("", "sampledir")
	if err != nil {
		return err
	}

	srcDir := "testdata"

	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, _ error) error {
		v := path
		colonPlaceholder := "{colon}"
		colonCharacter := ":"
		if strings.Contains(v, colonPlaceholder) {
			v = strings.ReplaceAll(v, colonPlaceholder, colonCharacter)
		}

		destPath := filepath.Join(tempTestDataDir, v)

		if d.IsDir() {
			return os.Mkdir(destPath, 0750)
		} else if d.Type() == fs.ModeSymlink {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, destPath)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(destPath, data, 0640)
	})
}

// teardownTestData removes the temporary directory with all its contents.
func teardownTestData() {
	_ = os.RemoveAll(tempTestDataDir)
}
