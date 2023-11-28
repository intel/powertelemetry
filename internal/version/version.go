// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package version

import (
	"fmt"
	"strings"
)

// Set via LDFLAGS -X.
var (
	LibName = "powertelemetry"
	Version = "unknown"
	Branch  = ""
	Commit  = ""
)

func GetFullVersion() string {
	var parts = []string{LibName}

	if Version != "" {
		parts = append(parts, Version)
	} else {
		parts = append(parts, "unknown")
	}

	if Branch != "" || Commit != "" {
		if Branch == "" {
			Branch = "unknown"
		}
		if Commit == "" {
			Commit = "unknown"
		}
		git := fmt.Sprintf("(git: %s@%s)", Branch, Commit)
		parts = append(parts, git)
	}

	return strings.Join(parts, " ")
}
