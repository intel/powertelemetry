// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

#include "textflag.h"

// func cpuid_count(level, count uint32) (eax, ebx, ecx, edx uint32)
TEXT ·cpuid_count(SB), NOSPLIT, $0-24
	MOVL level+0(FP), AX
	MOVL count+4(FP), CX
	CPUID
	MOVL AX, eax+8(FP)
	MOVL BX, ebx+12(FP)
	MOVL CX, ecx+16(FP)
	MOVL DX, edx+20(FP)
	RET
