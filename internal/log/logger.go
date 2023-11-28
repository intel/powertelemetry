// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package log

// Logger defines an interface for logging.
type Logger interface {
	Errorf(format string, args ...interface{})
	Error(args ...interface{})
	Debugf(format string, args ...interface{})
	Debug(args ...interface{})
	Warnf(format string, args ...interface{})
	Warn(args ...interface{})
	Infof(format string, args ...interface{})
	Info(args ...interface{})
}

// log defines a variable that stores the actual logger pointer.
var log Logger = &defaultLogger{}

// SetLogger sets a user provided logger structure to be used to log messages.
// If the provided logger is a nil pointer, a default no-op logger will be set.
func SetLogger(l Logger) {
	if l != nil {
		log = l
	} else {
		log = &defaultLogger{}
	}
}

// Errorf logs an error message.
func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

// Error logs an error message.
func Error(args ...interface{}) {
	log.Error(args...)
}

// Debugf logs a debug message.
func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

// Debug logs a debug message.
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Warnf logs a warning message.
func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

// Warn logs a warning message.
func Warn(args ...interface{}) {
	log.Warn(args...)
}

// Infof logs an information message.
func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// Info logs an information message.
func Info(args ...interface{}) {
	log.Info(args...)
}
