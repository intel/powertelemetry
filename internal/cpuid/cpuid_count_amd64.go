// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package cpuid

//nolint:revive // This is to keep the function name same as in Linux kernel sources
func cpuid_count(level, count uint32) (eax, ebx, ecx, edx uint32) // implemented in cpuid_count_amd64.s
