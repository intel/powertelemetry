// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	cpuUtil "github.com/shirou/gopsutil/v3/cpu"
)

const (
	// base path which holds global and individual CPU attributes.
	defaultDieBasePath = "/sys/devices/system/cpu"

	// name for die ID attribute file, corresponding to a specific CPU ID.
	dieFilename = "topology/die_id"
)

// topologyGetter gets topology information of the host.
type topologyGetter interface {
	getCPUVendor(cpuID int) (string, error)
	getCPUFamily(cpuID int) (string, error)
	getCPUDieID(cpuID int) (int, error)
	getCPUPackageID(cpuID int) (int, error)
	getCPUCoreID(cpuID int) (int, error)
	getCPUModel() int
	getCPUFlags(cpuID int) ([]string, error)
	getCPUsNumber() int
	getPackageDieIDs(packageID int) ([]int, error)
	getPackageIDs() []int
}

// topologyReader provides per-CPU ID attribute information of the host.
type topologyReader interface {
	// initTopology parses topology information from the host.
	initTopology() error

	topologyGetter
}

// cpuInfo represents attribute information of a CPU.
type cpuInfo struct {
	vendorID  string
	family    string
	dieID     int
	packageID int
	coreID    int
	flags     []string
}

// topologyData provides information about the processor of the host, including the number of CPUs present,
// CPUs details, CPU model, mapping between packages and dies and all package IDs.
// Implements topologyReader interface.
type topologyData struct {
	dieIDPath string

	topologyMap map[int]*cpuInfo
	packageDies map[int][]int
	packageIDs  []int
	model       int
}

// initTopology initializes information about the processor of the host, including the number of CPUs present,
// CPUs details, CPU model, mapping between packages and dies and all package IDs.
func (t *topologyData) initTopology() error {
	cpus, err := cpuUtil.Info()
	if err != nil {
		return fmt.Errorf("error occurred while parsing CPU information: %w", err)
	}
	if len(cpus) == 0 {
		return fmt.Errorf("no CPUs were found")
	}

	modelParsed := false
	cpuInfoMap := make(map[int]*cpuInfo, len(cpus))
	for _, singleCPUInfo := range cpus {
		info, err := parseCPUInfo(singleCPUInfo)
		if err != nil {
			return err
		}

		if !modelParsed {
			t.model, err = strconv.Atoi(singleCPUInfo.Model)
			if err != nil {
				return fmt.Errorf("error parsing model: %w", err)
			}
			modelParsed = true
		}

		cpuInfoMap[int(singleCPUInfo.CPU)] = info
	}

	t.packageDies = make(map[int][]int)
	// Attempt to retrieve die ID for each CPU ID from sysfs
	// If not retrieved, default value is zero, as in turbostat.
	for cpuID, cInfo := range cpuInfoMap {
		cpuDir := "cpu" + strconv.Itoa(cpuID)
		dieName := filepath.Join(t.dieIDPath, cpuDir, dieFilename)
		dieID, err := extractDieID(dieName)
		if err != nil {
			continue
		}
		cpuInfoMap[cpuID].dieID = dieID
		t.addDieToPackageDies(cInfo.packageID, dieID)
	}
	t.topologyMap = cpuInfoMap

	// slices.Compact replaces consecutive runs of equal elements with a single copy
	// (therefore, the slice must be sorted earlier to remove duplicates)
	for packageID, dies := range t.packageDies {
		slices.Sort(dies)
		t.packageDies[packageID] = slices.Compact(dies)
	}

	// Get ordered slice of unique package IDs.
	t.packageIDs = packageIDs(t.topologyMap)

	return nil
}

// parseCPUInfo parses information from single CPU.
func parseCPUInfo(infoStat cpuUtil.InfoStat) (*cpuInfo, error) {
	physicalID, err := strconv.Atoi(infoStat.PhysicalID)
	if err != nil {
		return nil, fmt.Errorf("error parsing physical ID: %w", err)
	}

	coreID, err := strconv.Atoi(infoStat.CoreID)
	if err != nil {
		return nil, fmt.Errorf("error parsing core ID: %w", err)
	}

	return &cpuInfo{
		vendorID:  infoStat.VendorID,
		family:    infoStat.Family,
		packageID: physicalID,
		coreID:    coreID,
		flags:     infoStat.Flags,
	}, nil
}

// packageIDs takes a topology map and returns a sorted slice with unique package IDs.
func packageIDs(topologyMap map[int]*cpuInfo) []int {
	pkgIDs := make([]int, 0, len(topologyMap))
	for _, info := range topologyMap {
		pkgIDs = append(pkgIDs, info.packageID)
	}
	slices.Sort(pkgIDs)
	return slices.Compact(pkgIDs)
}

// extractDieID extracts id of die from dieFile.
func extractDieID(dieFile string) (dieID int, err error) {
	// Return 0 in case die_id does not exist
	exists, err := fileExists(dieFile)
	if err != nil {
		return 0, fmt.Errorf("error opening file %q: %w", dieFile, err)
	}
	if !exists {
		return 0, nil
	}

	fileContent, err := readFile(dieFile)
	if err != nil {
		return 0, fmt.Errorf("error reading file %q: %w", dieFile, err)
	}

	dieID, err = strconv.Atoi(strings.TrimSpace(string(fileContent)))
	if err != nil {
		return 0, fmt.Errorf("error converting die ID value from the file %q to int: %w", dieFile, err)
	}

	return dieID, nil
}

func (t *topologyData) addDieToPackageDies(packageID int, dieID int) {
	dies, ok := t.packageDies[packageID]
	if ok {
		dies = append(dies, dieID)
	} else {
		dies = []int{dieID}
	}
	t.packageDies[packageID] = dies
}

// getCPUVendor gets cpu's vendorID value. If no cpu is found for the corresponding cpuID
// an error is returned.
func (t *topologyData) getCPUVendor(cpuID int) (string, error) {
	if info, ok := t.topologyMap[cpuID]; ok {
		return info.vendorID, nil
	}
	return "", fmt.Errorf("cpu: %d doesn't exist", cpuID)
}

// getCPUFamily gets cpu's family value. If no cpu is found for the corresponding cpuID
// an error is returned.
func (t *topologyData) getCPUFamily(cpuID int) (string, error) {
	if info, ok := t.topologyMap[cpuID]; ok {
		return info.family, nil
	}
	return "", fmt.Errorf("cpu: %d doesn't exist", cpuID)
}

// getCPUDieID gets cpu's dieID value. If no cpu is found for the corresponding cpuID
// an error is returned.
func (t *topologyData) getCPUDieID(cpuID int) (int, error) {
	if info, ok := t.topologyMap[cpuID]; ok {
		return info.dieID, nil
	}
	return 0, fmt.Errorf("cpu: %d doesn't exist", cpuID)
}

// getCPUPackageID gets cpu's package ID value. If no cpu is found for the corresponding cpuID
// an error is returned.
func (t *topologyData) getCPUPackageID(cpuID int) (int, error) {
	if info, ok := t.topologyMap[cpuID]; ok {
		return info.packageID, nil
	}
	return 0, fmt.Errorf("cpu: %d doesn't exist", cpuID)
}

// getCPUCoreID gets cpu's core ID value. If no cpu is found for the corresponding cpuID
// an error is returned.
func (t *topologyData) getCPUCoreID(cpuID int) (int, error) {
	if info, ok := t.topologyMap[cpuID]; ok {
		return info.coreID, nil
	}
	return 0, fmt.Errorf("cpu: %d doesn't exist", cpuID)
}

// getCPUModel gets model value of CPU.
func (t *topologyData) getCPUModel() int {
	return t.model
}

// getCPUFlags gets cpu's flags' values. If no cpu is found for the corresponding cpuID
// an error is returned.
func (t *topologyData) getCPUFlags(cpuID int) ([]string, error) {
	if info, ok := t.topologyMap[cpuID]; ok {
		return info.flags, nil
	}
	return nil, fmt.Errorf("cpu: %d doesn't exist", cpuID)
}

// getCPUsNumber returns the number of logical CPUs on a server.
func (t *topologyData) getCPUsNumber() int {
	return len(t.topologyMap)
}

// getCPUDieID gets cpu's dieID value. If no cpu is found for the corresponding cpuID
// an error is returned.
func (t *topologyData) getPackageDieIDs(packageID int) ([]int, error) {
	if dies, ok := t.packageDies[packageID]; ok {
		return dies, nil
	}
	return nil, fmt.Errorf("package: %d doesn't exist", packageID)
}

// getPackageIDs returns a slice with ordered package IDs of the host topology.
func (t *topologyData) getPackageIDs() []int {
	return t.packageIDs
}
