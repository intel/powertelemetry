// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"time"

	"github.com/jmhodges/clock"
)

var (
	// holds function definition to retrieve the current local time.
	timeNowFn func() time.Time

	// holds a fake clock used to test time-sensitive code.
	fakeClock clock.FakeClock
)

// setFakeClock gates the use of a fake clock for unit tests to retrieve
// the current local time.
func setFakeClock() {
	timeNowFn = fakeClock.Now
}

// unsetFakeClock restores timeNowFn function to retrieve the current time from the host.
func unsetFakeClock() {
	timeNowFn = time.Now
}

func init() {
	timeNowFn = time.Now
	fakeClock = clock.NewFake()
}
