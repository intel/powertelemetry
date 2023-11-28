// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package cpumodel

// Model definitions for CPU models are taken from the following kernel source link:
// https://github.com/torvalds/linux/blob/master/arch/x86/include/asm/intel-family.h

/*
 * "Big Core" Processors (Branded as Core, Xeon, etc...)
 *
 * While adding a new CPUID for a new microarchitecture, add a new
 * group to keep logically sorted out in chronological order. Within
 * that group keep the CPUID for the variants sorted by model number.
 *
 * The defined symbol names have the following form:
 * INTEL_FAM6{OPTFAMILY}_{MICROARCH}{OPTDIFF}
 * where:
 * OPTFAMILY Describes the family of CPUs that this belongs to. Default
 *  is assumed to be "_CORE" (and should be omitted). Other values
 *  currently in use are _ATOM and _XEON_PHI
 * MICROARCH Is the code name for the micro-architecture for this core.
 *  N.B. Not the platform name.
 * OPTDIFF If needed, a short string to differentiate by market segment.
 *
 *  Common OPTDIFFs:
 *
 *   - regular client parts
 *  _L - regular mobile parts
 *  _G - parts with extra graphics on
 *  _X - regular server parts
 *  _D - micro server parts
 *  _N,_P - other mobile parts
 *  _H - premium mobile parts
 *  _S - other client parts
 *
 *  Historical OPTDIFFs:
 *
 *  _EP - 2 socket server parts
 *  _EX - 4+ socket server parts
 *
 * Lines may optionally include a comment including platform or core
 * names. An exception is made for skylake/kabylake where steppings seem to have gotten
 * their own names :-(
 */

//nolint:revive,godot // this is to keep CPU model definitions same as in Linux kernel sources
const (
	INTEL_FAM6_CORE_YONAH = 0x0E

	INTEL_FAM6_CORE2_MEROM      = 0x0F
	INTEL_FAM6_CORE2_MEROM_L    = 0x16
	INTEL_FAM6_CORE2_PENRYN     = 0x17
	INTEL_FAM6_CORE2_DUNNINGTON = 0x1D

	INTEL_FAM6_NEHALEM    = 0x1E
	INTEL_FAM6_NEHALEM_G  = 0x1F /* Auburndale / Havendale */
	INTEL_FAM6_NEHALEM_EP = 0x1A
	INTEL_FAM6_NEHALEM_EX = 0x2E

	INTEL_FAM6_WESTMERE    = 0x25
	INTEL_FAM6_WESTMERE_EP = 0x2C
	INTEL_FAM6_WESTMERE_EX = 0x2F

	INTEL_FAM6_SANDYBRIDGE   = 0x2A
	INTEL_FAM6_SANDYBRIDGE_X = 0x2D
	INTEL_FAM6_IVYBRIDGE     = 0x3A
	INTEL_FAM6_IVYBRIDGE_X   = 0x3E

	INTEL_FAM6_HASWELL   = 0x3C
	INTEL_FAM6_HASWELL_X = 0x3F
	INTEL_FAM6_HASWELL_L = 0x45
	INTEL_FAM6_HASWELL_G = 0x46

	INTEL_FAM6_BROADWELL   = 0x3D
	INTEL_FAM6_BROADWELL_G = 0x47
	INTEL_FAM6_BROADWELL_X = 0x4F
	INTEL_FAM6_BROADWELL_D = 0x56

	INTEL_FAM6_SKYLAKE_L = 0x4E /* Sky Lake          */
	INTEL_FAM6_SKYLAKE   = 0x5E /* Sky Lake          */
	INTEL_FAM6_SKYLAKE_X = 0x55 /* Sky Lake          */
	/* CASCADELAKE_X     = 0x55    Sky Lake -- s: 7  */
	/* COOPERLAKE_X      = 0x55    Sky Lake -- s: 11 */

	INTEL_FAM6_KABYLAKE_L = 0x8E /* Sky Lake             */
	/* AMBERLAKE_L        = 0x8E    Sky Lake -- s: 9     */
	/* COFFEELAKE_L       = 0x8E    Sky Lake -- s: 10    */
	/* WHISKEYLAKE_L      = 0x8E    Sky Lake -- s: 11,12 */

	INTEL_FAM6_KABYLAKE = 0x9E /* Sky Lake              */
	/* COFFEELAKE       = 0x9E    Sky Lake -- s: 10-13  */

	INTEL_FAM6_COMETLAKE   = 0xA5 /* Sky Lake */
	INTEL_FAM6_COMETLAKE_L = 0xA6 /* Sky Lake */

	INTEL_FAM6_CANNONLAKE_L = 0x66 /* Palm Cove */

	INTEL_FAM6_ICELAKE_X    = 0x6A /* Sunny Cove */
	INTEL_FAM6_ICELAKE_D    = 0x6C /* Sunny Cove */
	INTEL_FAM6_ICELAKE      = 0x7D /* Sunny Cove */
	INTEL_FAM6_ICELAKE_L    = 0x7E /* Sunny Cove */
	INTEL_FAM6_ICELAKE_NNPI = 0x9D /* Sunny Cove */

	INTEL_FAM6_ROCKETLAKE = 0xA7 /* Cypress Cove */

	INTEL_FAM6_TIGERLAKE_L = 0x8C /* Willow Cove */
	INTEL_FAM6_TIGERLAKE   = 0x8D /* Willow Cove */

	INTEL_FAM6_SAPPHIRERAPIDS_X = 0x8F /* Golden Cove */

	INTEL_FAM6_EMERALDRAPIDS_X = 0xCF

	INTEL_FAM6_GRANITERAPIDS_X = 0xAD
	INTEL_FAM6_GRANITERAPIDS_D = 0xAE

	/* "Hybrid" Processors (P-Core/E-Core) */

	INTEL_FAM6_LAKEFIELD = 0x8A /* Sunny Cove / Tremont */

	INTEL_FAM6_ALDERLAKE   = 0x97 /* Golden Cove / Gracemont */
	INTEL_FAM6_ALDERLAKE_L = 0x9A /* Golden Cove / Gracemont */

	INTEL_FAM6_RAPTORLAKE   = 0xB7 /* Raptor Cove / Enhanced Gracemont */
	INTEL_FAM6_RAPTORLAKE_P = 0xBA
	INTEL_FAM6_RAPTORLAKE_S = 0xBF

	INTEL_FAM6_METEORLAKE   = 0xAC
	INTEL_FAM6_METEORLAKE_L = 0xAA

	INTEL_FAM6_ARROWLAKE_H = 0xC5
	INTEL_FAM6_ARROWLAKE   = 0xC6

	INTEL_FAM6_LUNARLAKE_M = 0xBD

	/* "Small Core" Processors (Atom/E-Core) */

	INTEL_FAM6_ATOM_BONNELL     = 0x1C /* Diamondville, Pineview */
	INTEL_FAM6_ATOM_BONNELL_MID = 0x26 /* Silverthorne, Lincroft */

	INTEL_FAM6_ATOM_SALTWELL        = 0x36 /* Cedarview */
	INTEL_FAM6_ATOM_SALTWELL_MID    = 0x27 /* Penwell */
	INTEL_FAM6_ATOM_SALTWELL_TABLET = 0x35 /* Cloverview */

	INTEL_FAM6_ATOM_SILVERMONT            = 0x37 /* Bay Trail, Valleyview */
	INTEL_FAM6_ATOM_SILVERMONT_D          = 0x4D /* Avaton, Rangely */
	INTEL_FAM6_ATOM_SILVERMONT_MID        = 0x4A /* Merriefield */
	INTEL_FAM6_ATOM_SILVERMONT_SMARTPHONE = 0x5A // INTEL_FAM6_ATOM_AIRMONT_MID in turbostat

	INTEL_FAM6_ATOM_AIRMONT    = 0x4C /* Cherry Trail, Braswell */
	INTEL_FAM6_ATOM_AIRMONT_NP = 0x75 /* Lightning Mountain */

	INTEL_FAM6_ATOM_GOLDMONT   = 0x5C /* Apollo Lake */
	INTEL_FAM6_ATOM_GOLDMONT_D = 0x5F /* Denverton */

	/* Note: the micro-architecture is "Goldmont Plus" */
	INTEL_FAM6_ATOM_GOLDMONT_PLUS = 0x7A /* Gemini Lake */

	INTEL_FAM6_ATOM_TREMONT_D = 0x86 /* Jacobsville */
	INTEL_FAM6_ATOM_TREMONT   = 0x96 /* Elkhart Lake */
	INTEL_FAM6_ATOM_TREMONT_L = 0x9C /* Jasper Lake */

	INTEL_FAM6_ATOM_GRACEMONT = 0xBE /* Alderlake N */

	INTEL_FAM6_ATOM_CRESTMONT_X = 0xAF /* Sierra Forest */
	INTEL_FAM6_ATOM_CRESTMONT   = 0xB6 /* Grand Ridge */

	/* Xeon Phi */

	INTEL_FAM6_XEON_PHI_KNL = 0x57 /* Knights Landing */
	INTEL_FAM6_XEON_PHI_KNM = 0x85 /* Knights Mill */
)
