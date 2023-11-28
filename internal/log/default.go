// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package log

// defaultLogger defines a default no-op logging structure.
type defaultLogger struct{}

func (l *defaultLogger) Errorf(_ string, _ ...interface{}) {}
func (l *defaultLogger) Error(_ ...interface{})            {}
func (l *defaultLogger) Debugf(_ string, _ ...interface{}) {}
func (l *defaultLogger) Debug(_ ...interface{})            {}
func (l *defaultLogger) Warnf(_ string, _ ...interface{})  {}
func (l *defaultLogger) Warn(_ ...interface{})             {}
func (l *defaultLogger) Infof(_ string, _ ...interface{})  {}
func (l *defaultLogger) Info(_ ...interface{})             {}
