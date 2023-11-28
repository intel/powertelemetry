// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"fmt"
	"strings"
)

// MultiError holds a slice of error descriptions. Implements Error interface.
// It is used to mark errors that happened during the initialization of PowerTelemetry dependencies.
type MultiError struct {
	errs []string
}

// add takes an error message and appends it to the receiver's slice of error descriptions.
func (e *MultiError) add(errMsg string) {
	e.errs = append(e.errs, errMsg)
}

// Error returns a string with all error descriptions. Implements error.Error.
func (e *MultiError) Error() string {
	return strings.Join(e.errs, "; ")
}

// ModuleNotInitializedError indicates that a module has not been initialized.
type ModuleNotInitializedError struct {
	Name string //holds name of not initialized module
}

// Error returns a reason of this error.
func (e *ModuleNotInitializedError) Error() string {
	return fmt.Sprintf("module %q is not initialized", e.Name)
}

// MetricNotSupportedError indicates that a metric is not supported.
type MetricNotSupportedError struct {
	reason string
}

// Error returns a reason of this error.
func (e *MetricNotSupportedError) Error() string {
	return e.reason
}
