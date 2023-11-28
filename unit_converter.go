// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

const (
	fromKiloHertzToMegaHertzRatio = 1e-3
	fromMicrojoulesToJoulesRatio  = 1e-6
	fromMicrowattsToWatts         = 1e-6
	fromProcessorCyclesToHertz    = 1e-6
	fromNanosecondsToSecondsRatio = 1e-9
)
