// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	// control zone path where rapl exposes power capping capabilities to userspace.
	defaultRaplBasePath = "/sys/devices/virtual/powercap/intel-rapl"

	// pattern string to identify the path of a package domain zone.
	zonePattern = "intel-rapl\\:[0-9]*"

	// pattern string to identify the path of a package domain subzone, i.e. dram domain.
	subzonePattern = "intel-rapl\\:[0-9]*\\:[0-9]*"

	// file name of the maximum energy attribute supported by power capping.
	maxEnergyAttrFile = "max_energy_range_uj"

	// file name of the current energy attribute supported by power capping.
	currEnergyAttrFile = "energy_uj"

	// file name of the maximum allowed power constraint attribute supported by power capping.
	maxPowerConstraintAttrFile = "constraint_0_max_power_uw"
)

var (
	// regex used to check name format of a package domain zone.
	packageNameRegex = regexp.MustCompile("^package-(0|[1-9][0-9]*)$")

	// regex used to check the path format of a package domain zone.
	zoneRegex = regexp.MustCompile(zonePattern)

	// regex used to check the path format of a package domain subzone.
	subzoneRegex = regexp.MustCompile(subzonePattern)
)

// domainType is an enum type to identify specific intel rapl control domains.
type domainType int

// domainType enum defines supported intel rapl control domains.
const (
	packageDomain domainType = iota
	dramDomain
)

// Helper function to return a string representation of domainType.
func (d domainType) String() string {
	switch d {
	case packageDomain:
		return "package"
	case dramDomain:
		return "dram"
	default:
		return ""
	}
}

// attrType is an enum type to identify specific zone attributes.
type attrType int

// attrType enum defines supported zone attributes.
const (
	currEnergyAttr attrType = iota
	maxEnergyAttr
	maxPowerConstraintAttr
)

// Helper function to return a string representation of attrType.
func (e attrType) String() string {
	switch e {
	case currEnergyAttr:
		return "currEnergy"
	case maxEnergyAttr:
		return "maxEnergy"
	case maxPowerConstraintAttr:
		return "maxPower"
	default:
		return ""
	}
}

// attrSample represents a timestamped zone attribute sample.
type attrSample struct {
	value     float64
	timestamp time.Time
}

// powerZone represents a generic power zone accessible by power capping interface.
type powerZone interface {
	// getName gets the name of a zone.
	getName() string

	// getPath gets the absolute path of a zone.
	getPath() string

	// addSubzone takes a powerZone and adds it as a child zone.
	addSubzone(subzone powerZone)

	// getSubzones returns all the child zones of the given zone.
	getSubzones() []powerZone

	// getDomainSubzone returns the child zone with specified domain.
	getDomainSubzone(domain string) powerZone

	// getEnergySample returns the last energy sample stored in the given zone.
	getEnergySample() attrSample

	// setEnergySample sets the given energy sample to the given zone.
	setEnergySample(s attrSample)

	// readAttribute gets a timestamped sample of the specified attribute.
	readAttribute(attribute string) (attrSample, error)
}

// zone represents a generic power zone that can be monitored using power capping
// method determined by the control type the given zone belongs to.
// zones are hierarchical, meaning a parent zone can have multiple child zones
// or subzones, representing different parts of the system.
//
// An example of hierarchical directory tree for a zone is as follows:
//
// root-zone
// ├── package-zone:0
// │   ├── package-zone:0:0 (device subzone)
// │   ├── package-zone:0:1
// │   ∙
// |   └── package-zone:0:n
// ├── package-zone:1
// │   ├── package-zone:1:0 (device subzone)
// │   ├── package-zone:1:1
// │   ∙
// │   └── package-zone:1:m
// ∙
// └── package-zone:l
//
// Each zone has capabilities to retrieve timestamped values of zone attributes
// from files.
//
// package-zone:0
// ├── constraint_0_max_power_uw (maximum allowed power attribute)
// ├── energy_uj                 (current energy attribute)
// └── max_energy_range_uj       (maximum energy attribute)
//
// energy field of a zone stores the last measured current energy attribute.
type zone struct {
	name     string
	path     string
	energy   attrSample
	subzones []powerZone
}

// newZoneFromPath creates a new zone, initializing its name with
// contents of name file located at the specified path. If name file
// does not exist, or it is empty an error is returned.
func newZoneFromPath(path string) (powerZone, error) {
	f := filepath.Join(path, "name")
	data, err := readFile(f)
	if err != nil {
		return nil, err
	}

	name := strings.TrimRight(string(data), "\n")
	if len(name) == 0 {
		return nil, fmt.Errorf("zone domain cannot be empty")
	}

	return &zone{
		name: name,
		path: path,
		energy: attrSample{
			value:     0.0,
			timestamp: time.Time{},
		},
		subzones: make([]powerZone, 0),
	}, nil
}

// addSubzone takes a zone as an argument and adds it as a child
// of the receiver zone.
func (z *zone) addSubzone(subzone powerZone) {
	z.subzones = append(z.subzones, subzone)
}

// getName returns the name of a zone.
func (z *zone) getName() string {
	return z.name
}

// getPath returns the path at which a zone is located.
func (z *zone) getPath() string {
	return z.path
}

// getDomainSubzone loops through the subzones of the receiver zone,
// and returns the zone matching the name given as argument.
// If there are no matches it returns nil.
func (z *zone) getDomainSubzone(domain string) powerZone {
	for _, subzone := range z.subzones {
		if strings.Contains(subzone.getName(), domain) {
			return subzone
		}
	}
	return nil
}

// getSubzones returns a slice with subzones of the receiver zone.
func (z *zone) getSubzones() []powerZone {
	return z.subzones
}

// getEnergySample returns the last energy sample collected for the
// receiver zone.
func (z *zone) getEnergySample() attrSample {
	return z.energy
}

// setEnergySample sets the given energy sample to the receiver zone.
func (z *zone) setEnergySample(s attrSample) {
	z.energy = s
}

// readAttribute returns a timestamped sample of the specified attribute
// for a given zone.
func (z *zone) readAttribute(attribute string) (attrSample, error) {
	var attrFilePath string
	switch attribute {
	case currEnergyAttr.String():
		attrFilePath = filepath.Join(z.path, currEnergyAttrFile)
	case maxEnergyAttr.String():
		attrFilePath = filepath.Join(z.path, maxEnergyAttrFile)
	case maxPowerConstraintAttr.String():
		attrFilePath = filepath.Join(z.path, maxPowerConstraintAttrFile)
	default:
		return attrSample{}, fmt.Errorf("unsupported attribute %q", attribute)
	}

	data, timestamp, err := readFileWithTimestamp(attrFilePath)
	if err != nil {
		return attrSample{}, fmt.Errorf("error reading file %q: %w", attrFilePath, err)
	}
	val, err := strconv.ParseFloat(strings.TrimRight(string(data), "\n"), 64)
	if err != nil {
		return attrSample{}, fmt.Errorf("error converting attribute file content to float64: %w", err)
	}
	return attrSample{
		value:     val,
		timestamp: timestamp,
	}, nil
}

// packageZone is a specialized case of powerZone. It extends functionality
// of a generic zone adding validation for fields specific to package domain zones.
type packageZone struct {
	powerZone
}

// getPackageID returns the package ID of the package domain zone. It performs
// validation between the content name file of the package zone and its path.
func (p *packageZone) getPackageID() (int, error) {
	name := p.getName()
	path := p.getPath()

	if !packageNameRegex.MatchString(name) {
		return 0, fmt.Errorf("invalid package domain name for zone at path %q", path)
	}
	packageIDFromName := strings.Split(name, "-")[1]

	if !zoneRegex.MatchString(filepath.Base(path)) {
		return 0, fmt.Errorf("invalid package domain zone path %q", path)
	}
	packageIDFromPath := strings.Split(path, ":")[1]

	if packageIDFromPath != packageIDFromName {
		return 0, fmt.Errorf("package ID mismatch between zone path %q and zone name %q", path, name)
	}
	return strconv.Atoi(packageIDFromName)
}

// isPackageZone is a helper function that returns true if the power zone provided
// as argument is a package zone. Otherwise, it returns false.
func isPackageZone(z powerZone) bool {
	return packageNameRegex.MatchString(z.getName())
}

// raplReader checks if rapl kernel module is loaded and exposes power metrics supported by
// power capping interface.
//
// Exposed metric are:
// - Per-package ID current power consumption.
// - Per-dram current power consumption.
// - Per-package ID maximum allowed power.
type raplReader interface {
	// initZoneMap initializes a map of zones that represents the hierarchy tree for intel-rapl
	// control zones of the host.
	initZoneMap() error

	// getPackageIDs returns an ordered slice with package IDs within the map of zones.
	getPackageIDs() []int

	// isRaplLoaded check if intel-rapl kernel module is loaded.
	isRaplLoaded(modulesPath string) (bool, error)

	// getCurrentPowerConsumptionWatts takes a package ID and domain, and returns the current power consumption.
	getCurrentPowerConsumptionWatts(packageID int, domain string) (float64, error)

	// getMaxPowerConstraintWatts takes a package ID and returns the maximum allowed power.
	getMaxPowerConstraintWatts(packageID int) (float64, error)
}

// raplData represents per-package ID power zone tree of the intel rapl control zone
// of the host.
//
// It represents the hierarchy tree for intel-rapl control zone:
//
// /sys/devices/virtual/powercap/intel-rapl/
// ├── intel-rapl:0         (package zone)
// │   ├── intel-rapl:0:0   (device subzone)
// │   ├── intel-rapl:0:1
// │   ∙
// |   └── intel-rapl:0:n
// ├── intel-rapl:1         (package zone)
// │   ├── intel-rapl:1:0   (device subzone)
// │   ├── intel-rapl:1:1
// │   ∙
// │   └── intel-rapl:1:m
// ∙
// └── intel-rapl:l         (package zone)
//
// Each entry map corresponds to a package zone, which in turns has subzones corresponding
// to specific devices.
type raplData struct {
	basePath string
	zones    map[int]powerZone
}

// initZoneMap initializes the zone map of the receiver with the power zone tree corresponding
// to the host configuration. It validates that the root zone is a valid package domain
// zone. In case of malformed power zone trees, an error is returned.
func (r *raplData) initZoneMap() error {
	if len(r.basePath) == 0 {
		return errors.New("base path of rapl control zone cannot be empty")
	}
	if err := checkFile(r.basePath); err != nil {
		return fmt.Errorf("invalid base path of rapl control zone: %w", err)
	}

	zoneDirs, err := os.ReadDir(r.basePath)
	if err != nil {
		return fmt.Errorf("error reading path %q: %w", r.basePath, err)
	}

	// initialize package domain zones
	zones := make(map[int]powerZone, len(zoneDirs))
	for _, zoneDir := range zoneDirs {
		zoneName := zoneDir.Name()
		if !zoneDir.IsDir() || !zoneRegex.MatchString(zoneName) {
			continue
		}
		zonePath := filepath.Join(r.basePath, zoneName)
		newZone, err := newZoneFromPath(zonePath)
		if err != nil {
			return fmt.Errorf("error creating zone for path %q: %w", zonePath, err)
		}

		// skip if zone is not a package zone
		if !isPackageZone(newZone) {
			continue
		}

		// validate fields for the package zone
		pkgZone := &packageZone{newZone}
		packageID, err := pkgZone.getPackageID()
		if err != nil {
			return fmt.Errorf("error validating package domain zone: %w", err)
		}

		// initialize per package domain subzones
		subzoneDirs, err := os.ReadDir(zonePath)
		if err != nil {
			return fmt.Errorf("error reading directory %q: %w", zonePath, err)
		}

		for _, subzoneDir := range subzoneDirs {
			subzoneName := subzoneDir.Name()
			if !subzoneDir.IsDir() || !subzoneRegex.MatchString(subzoneName) {
				continue
			}
			subzonePath := filepath.Join(zonePath, subzoneName)
			subzone, err := newZoneFromPath(subzonePath)
			if err != nil {
				return fmt.Errorf("error creating subzone for path %q: %w", subzonePath, err)
			}
			newZone.addSubzone(subzone)
		}
		zones[packageID] = newZone
	}

	if len(zones) == 0 {
		return fmt.Errorf("no package zones found for base path %q", r.basePath)
	}

	// read and store a timestamped value of current energy attribute
	// for each package and dram domain zones.
	for _, pkgZone := range zones {
		s, err := pkgZone.readAttribute(currEnergyAttr.String())
		if err != nil {
			return fmt.Errorf("error initializing current energy attribute for package domain zone %q: %w", pkgZone.getPath(), err)
		}
		pkgZone.setEnergySample(s)
		if dramZone := pkgZone.getDomainSubzone(dramDomain.String()); dramZone != nil {
			s, err = dramZone.readAttribute(currEnergyAttr.String())
			if err != nil {
				return fmt.Errorf("error initializing current energy attribute for dram domain zone %q: %w", dramZone.getPath(), err)
			}
			dramZone.setEnergySample(s)
		}
	}
	r.zones = zones
	return nil
}

// getPackageIDs returns an ordered slice with package IDs within the map of zones.
func (r *raplData) getPackageIDs() []int {
	pkgIDs := make([]int, 0, len(r.zones))

	for packageID := range r.zones {
		pkgIDs = append(pkgIDs, packageID)
	}
	slices.Sort(pkgIDs)
	return pkgIDs
}

// isRaplLoaded returns true if intel rapl kernel module and its dependencies are
// loaded, otherwise returns false.
// TODO: Review implementation of this function to cover older kernel versions.
func (r *raplData) isRaplLoaded(modulesPath string) (bool, error) {
	if err := checkFile(modulesPath); err != nil {
		return false, err
	}

	f, err := os.Open(modulesPath)
	if err != nil {
		return false, fmt.Errorf("error opening file %q: %w", modulesPath, err)
	}
	defer f.Close()

	raplModules := map[string]bool{
		"rapl":              false,
		"intel_rapl_msr":    false,
		"intel_rapl_common": false,
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		mod := strings.Split(scanner.Text(), " ")[0]
		if _, ok := raplModules[mod]; ok {
			raplModules[mod] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("could not read file %q: %w", modulesPath, err)
	}

	res := raplModules["rapl"] && raplModules["intel_rapl_msr"] &&
		raplModules["intel_rapl_common"]
	return res, nil
}

// getEnergyAttributeWithTimestamp returns per-domain energy attribute, in Microjoules, for
// a specific package ID, and the timestamp of the operation.
func (r *raplData) getEnergyAttributeWithTimestamp(packageID int, domain, energyAttribute string) (attrSample, error) {
	z, ok := r.zones[packageID]
	if !ok {
		return attrSample{}, fmt.Errorf("could not find zone for package ID: %v", packageID)
	}

	switch domain {
	case packageDomain.String():
	case dramDomain.String():
		z = z.getDomainSubzone(domain)
		if z == nil {
			return attrSample{}, fmt.Errorf("could not find dram subzone for package ID: %v", packageID)
		}
	default:
		return attrSample{}, fmt.Errorf("unsupported rapl domain %q", domain)
	}

	sample, err := z.readAttribute(energyAttribute)
	if err != nil {
		return attrSample{}, fmt.Errorf("error reading energy attribute %q: %w", energyAttribute, err)
	}
	return sample, nil
}

// getLastMeasuredEnergyAttribute gets the per-domain last measured current energy attribute for a specific
// package ID.
func (r *raplData) getLastMeasuredEnergyAttribute(packageID int, domain string) (attrSample, error) {
	z, ok := r.zones[packageID]
	if !ok {
		return attrSample{}, fmt.Errorf("could not find zone for package ID: %v", packageID)
	}

	switch domain {
	case packageDomain.String():
	case dramDomain.String():
		z = z.getDomainSubzone(domain)
		if z == nil {
			return attrSample{}, fmt.Errorf("could not find dram subzone for package ID: %v", packageID)
		}
	default:
		return attrSample{}, fmt.Errorf("unsupported rapl domain %q", domain)
	}
	return z.getEnergySample(), nil
}

// setLastMeasuredEnergyAttribute sets the per-domain last measured current energy attribute to the one
// provided as argument, for a specific package ID.
func (r *raplData) setLastMeasuredEnergyAttribute(packageID int, domain string, sample attrSample) error {
	z, ok := r.zones[packageID]
	if !ok {
		return fmt.Errorf("could not find zone for package ID: %v", packageID)
	}

	switch domain {
	case packageDomain.String():
	case dramDomain.String():
		z = z.getDomainSubzone(domain)
		if z == nil {
			return fmt.Errorf("could not find dram subzone for package ID: %v", packageID)
		}
	default:
		return fmt.Errorf("unsupported rapl domain %q", domain)
	}

	z.setEnergySample(sample)
	return nil
}

// getCurrentPowerConsumptionWatts returns per-domain current power consumption, in Watts, for a
// specific package ID.
func (r *raplData) getCurrentPowerConsumptionWatts(packageID int, domain string) (float64, error) {
	var power float64

	// Get last measured current energy attribute for the corresponding zone
	s1, err := r.getLastMeasuredEnergyAttribute(packageID, domain)
	if err != nil {
		return 0.0, fmt.Errorf("error getting last measured current energy attribute for %q domain: %w", domain, err)
	}

	// Get current energy attribute for the corresponding zone
	s2, err := r.getEnergyAttributeWithTimestamp(packageID, domain, currEnergyAttr.String())
	if err != nil {
		return 0.0, fmt.Errorf("error reading current energy attribute for %q domain: %w", domain, err)
	}

	// Calculate power as the ratio between the delta of energy samples and time delta
	timeDelta := s2.timestamp.Sub(s1.timestamp).Seconds()
	if s2.value > s1.value {
		power = fromMicrojoulesToJoulesRatio * (s2.value - s1.value) / timeDelta
	} else {
		// The value of current energy attribute is reset to zero when it reaches the value of maximum
		// energy attribute. In this case the value of maximum energy attribute is used to calculate the
		// energy delta.
		sMax, err := r.getEnergyAttributeWithTimestamp(packageID, domain, maxEnergyAttr.String())
		if err != nil {
			return 0.0, fmt.Errorf("error reading current energy attribute for %q domain: %w", domain, err)
		}
		power = fromMicrojoulesToJoulesRatio * (sMax.value + s2.value - s1.value) / timeDelta
	}

	// Set current energy attribute of the corresponding zone as last measured
	if err = r.setLastMeasuredEnergyAttribute(packageID, domain, s2); err != nil {
		return 0.0, fmt.Errorf("error setting current energy attribute for %q domain: %w", domain, err)
	}
	return power, nil
}

// getMaxPowerConstraintWatts returns the maximum allowed power, in Watts, for a specific package ID.
func (r *raplData) getMaxPowerConstraintWatts(packageID int) (float64, error) {
	z, ok := r.zones[packageID]
	if !ok {
		return 0.0, fmt.Errorf("could not find zone for package ID: %v", packageID)
	}
	s, err := z.readAttribute(maxPowerConstraintAttr.String())
	if err != nil {
		return 0.0, fmt.Errorf("error reading max power constraint attribute for package ID: %v: %w", packageID, err)
	}
	return s.value * fromMicrowattsToWatts, nil
}
