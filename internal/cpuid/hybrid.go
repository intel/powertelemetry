// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package cpuid

var (
	cpuIsHybrid            bool
	cpuIsHybridCheckedOnce bool
)

// IsCPUHybrid checks if CPU is hybrid (Alder Lake, Raptor Lake, Meteor Lake, etc.)
// The function ensures the actual cpuid reading is done only once.
// The cpuid value handling is aligned with the process_cpuid() function of turbostat.
func IsCPUHybrid() bool {
	if !cpuIsHybridCheckedOnce {
		cpuIsHybridCheckedOnce = true
		maxLevel, _, _, _ := cpuid_count(0, 0)
		if maxLevel >= 0x7 {
			_, _, _, edx := cpuid_count(7, 0)
			if (edx & (1 << 15)) != 0 {
				cpuIsHybrid = true
			}
		}
	}
	return cpuIsHybrid
}
