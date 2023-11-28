// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/intel/powertelemetry/internal/cpumodel"
	"github.com/intel/powertelemetry/internal/log"
)

var (
	cStateOffsets    = []uint32{c3Residency, c6Residency, c7Residency, maxFreqClockCount, actualFreqClockCount, timestampCounter}
	cStatePerfEvents = []string{c01.String(), c02.String(), c0Wait.String(), thread.String()}
)

// PowerTelemetry enables monitoring platform metrics.
type PowerTelemetry struct {
	topology   topologyReader
	msr        msrReaderWithStorage
	uncoreFreq uncoreFreqReader
	rapl       raplReader
	cpuFreq    cpuFreqReader
	perf       perfReaderWithStorage

	busClock float64
	cpus     []int
}

// powerBuilder enables piecewise builds of PowerTelemetry instances. Implements functional options pattern.
// TODO: User should be able to specify custom base path for powertelemetry subsystems.
type powerBuilder struct {
	topology   *topologyBuilder
	msr        *msrBuilder
	rapl       *raplBuilder
	coreFreq   *coreFreqBuilder
	uncoreFreq *uncoreFreqBuilder
	perf       *perfBuilder

	includedCPUs []int
	excludedCPUs []int
}

type Option func(*powerBuilder)

// WithExcludedCPUs returns a function closure that sets a slice with excluded CPU IDs
// of a builder.
func WithExcludedCPUs(cpuIDs []int) Option {
	return func(b *powerBuilder) {
		b.excludedCPUs = cpuIDs
	}
}

// WithIncludedCPUs returns a function closure that sets a slice with included CPU IDs
// of a builder.
func WithIncludedCPUs(cpuIDs []int) Option {
	return func(b *powerBuilder) {
		b.includedCPUs = cpuIDs
	}
}

// WithMsr returns a function closure that initializes the msrBuilder struct of a builder with the default configuration.
func WithMsr() Option {
	return func(b *powerBuilder) {
		if b.msr == nil {
			b.msr = &msrBuilder{
				msrReaderWithStorage: &msrDataWithStorage{
					msrOffsets: cStateOffsets,
					msrPath:    defaultMsrBasePath,
				},
			}
		}
	}
}

// WithMsrTimeout returns a function closure that initializes the msrBuilder struct of a builder with the default configuration
// and given msr read timeout.
func WithMsrTimeout(timeout time.Duration) Option {
	return func(b *powerBuilder) {
		if b.msr == nil {
			b.msr = &msrBuilder{
				msrReaderWithStorage: &msrDataWithStorage{
					msrOffsets: cStateOffsets,
					msrPath:    defaultMsrBasePath,
				},
			}
		}
		b.msr.timeout = timeout
	}
}

// WithRapl returns a function closure that initializes the raplBuilder struct of a builder with the default configuration.
func WithRapl(basePath ...string) Option {
	var path string
	if len(basePath) != 0 {
		path = basePath[0]
	} else {
		path = defaultRaplBasePath
	}
	return func(b *powerBuilder) {
		b.rapl = &raplBuilder{
			raplReader: &raplData{
				basePath: path,
			},
		}
	}
}

// WithCoreFrequency returns a function closure that initializes the coreFreqBuilder struct of a builder with the default configuration.
func WithCoreFrequency(basePath ...string) Option {
	var path string
	if len(basePath) != 0 {
		path = basePath[0]
	} else {
		path = defaultCPUFreqBasePath
	}
	return func(b *powerBuilder) {
		b.coreFreq = &coreFreqBuilder{
			cpuFreqReader: &cpuFreqData{
				cpuFrequencyFilePath: path,
			},
		}
	}
}

// WithUncoreFrequency returns a function closure that initializes the uncoreFreqBuilder struct of a builder with the default configuration.
func WithUncoreFrequency(basePath ...string) Option {
	var path string
	if len(basePath) != 0 {
		path = basePath[0]
	} else {
		path = defaultUncoreFreqBasePath
	}
	return func(b *powerBuilder) {
		b.uncoreFreq = &uncoreFreqBuilder{
			uncoreFreqReader: &uncoreFreqData{
				uncoreFreqBasePath: path,
			},
		}
	}
}

// WithPerf takes a file path with perf event definition in JSON format. It returns a function closure
// that initializes a perfBuilder struct with the given JSON event definition file.
func WithPerf(jsonFile string) Option {
	return func(b *powerBuilder) {
		b.perf = &perfBuilder{
			perfReaderWithStorage: &perfWithStorage{
				perfReader: newPerf(),
			},
			jsonPath: jsonFile,
			events:   cStatePerfEvents,
		}
	}
}

// WithLogger returns a function closure that sets a user provided logger structure to be used to log messages.
// Note: this option is supposed to go first in the list of arguments passed to New() when creating a PowerTelemetry instance.
func WithLogger(l log.Logger) Option {
	return func(b *powerBuilder) {
		log.SetLogger(l)
	}
}

// New returns a PowerTelemetry instance that allows to gather power-related metrics from the host.
// An error is returned if either topology could not be initialized, or if the CPU
// model is not supported. Otherwise, a MultiError is returned if one or more user requested
// subsystems (pieces) fail to initialize.
func New(opts ...Option) (*PowerTelemetry, error) {
	b := &powerBuilder{
		topology: &topologyBuilder{
			topologyReader: &topologyData{
				dieIDPath: defaultDieBasePath,
			},
		},
	}

	for _, opt := range opts {
		opt(b)
	}

	pt := &PowerTelemetry{}

	// initialize topology
	if err := b.topology.initTopology(); err != nil {
		return nil, err
	}
	pt.topology = b.topology
	logTopologyDetails(pt.topology)

	// check if processor is supported
	isSupported, err := isCPUSupported(b.topology)
	if err != nil {
		return nil, fmt.Errorf("error retrieving host processor: %w", err)
	}
	if !isSupported {
		return nil, errors.New("host processor is not supported")
	}

	// get available CPU IDs which can be accessed to get metrics from
	// (and check if no calls to both WithIncludedCPUs and WithExcludedCPUs have been done)
	cpus, err := b.getAvailableCPUs()
	if err != nil {
		return nil, fmt.Errorf("failed to get available CPUs: %w", err)
	}

	// check that not all CPU IDs are excluded.
	if len(cpus) == 0 {
		return nil, errors.New("no available CPUs were found")
	}
	pt.cpus = cpus
	log.Debugf("CPU metrics related to MSR and coreFreq can be gathered for following CPUs: %v", pt.cpus)

	// custom error to mark non-critical initialization errors
	multiErr := &MultiError{}

	// initialize msr
	pt.msr, err = b.initMsr(cpus)
	if err != nil {
		multiErr.add(fmt.Sprintf("failed to initialize msr: %v", err))
	}

	// initialize rapl
	pt.rapl, err = b.initRapl()
	if err != nil {
		multiErr.add(fmt.Sprintf("failed to initialize rapl: %v", err))
	}

	// initialize cpu frequency
	// TODO: Add argument with enabled CPU IDs.
	pt.cpuFreq, err = b.initCoreFreq()
	if err != nil {
		multiErr.add(fmt.Sprintf("failed to initialize core freq: %v", err))
	}

	// TODO: Consider to remove init method
	// initialize uncore frequency
	pt.uncoreFreq, err = b.initUncoreFreq()
	if err != nil {
		multiErr.add(fmt.Sprintf("failed to initialize uncore freq: %v", err))
	}

	// initialize perf
	pt.perf, err = b.initPerf(cStatePerfEvents, cpus)
	if err != nil {
		multiErr.add(fmt.Sprintf("failed to initialize perf: %v", err))
	}

	// TODO: Remove this optimization. Call to bus clock should be done only when needed.
	// TODO: Getting model can be done inside getBusClock, pt.topology.getModel()
	model := b.topology.getCPUModel()
	pt.busClock, err = pt.getBusClock(model)
	if err != nil {
		multiErr.add(fmt.Sprintf("failed to get bus clock for model: 0x%X: %v", model, err))
	}

	if len(multiErr.errs) > 0 {
		return pt, fmt.Errorf("PowerTelemetry instance initialized with errors: %w", multiErr)
	}
	return pt, nil
}

// topologyBuilder enables initialization of topology subsystem for PowerTelemetry instances.
type topologyBuilder struct {
	topologyReader
}

// msrBuilder enables configuration and initialization of msr subsystem for PowerTelemetry instances.
type msrBuilder struct {
	msrReaderWithStorage

	timeout time.Duration
}

// raplBuilder enables configuration and initialization of rapl subsystem for PowerTelemetry instances.
type raplBuilder struct {
	raplReader
}

// coreFreqBuilder enables configuration and initialization of coreFreq subsystem for PowerTelemetry instances.
type coreFreqBuilder struct {
	cpuFreqReader
}

// uncoreFreqBuilder enables configuration and initialization of uncoreFreq subsystem for PowerTelemetry instances.
type uncoreFreqBuilder struct {
	uncoreFreqReader
}

// perfBuilder enables configuration and initialization of perf subsystem for PowerTelemetry instances.
type perfBuilder struct {
	perfReaderWithStorage

	jsonPath string
	events   []string
}

// getAvailableCPUs returns a slice with available CPU IDs which can be accessed to get metrics from.
func (b *powerBuilder) getAvailableCPUs() ([]int, error) {
	if len(b.excludedCPUs) != 0 && len(b.includedCPUs) != 0 {
		return nil, errors.New("invalid CPU ID configuration, only one of both included or excluded modes allowed")
	}

	numCPUs := b.topology.getCPUsNumber()
	if err := validateFromRange(b.includedCPUs, 0, numCPUs-1); err != nil {
		return nil, fmt.Errorf("failed to validate included CPU IDs: %w", err)
	}

	if len(b.includedCPUs) != 0 {
		return b.includedCPUs, nil
	}

	if err := validateFromRange(b.excludedCPUs, 0, numCPUs-1); err != nil {
		return nil, fmt.Errorf("failed to validate excluded CPU IDs: %w", err)
	}

	cpus := make([]int, 0, numCPUs-len(b.excludedCPUs))
	for i := 0; i < numCPUs; i++ {
		if slices.Contains(b.excludedCPUs, i) {
			continue
		}
		cpus = append(cpus, i)
	}

	return cpus, nil
}

// validateFromRange takes a slice of ints, a low, and a high bound. It returns an error
// in case any element of the slice is not within the interval [low, high].
func validateFromRange(nums []int, low, high int) error {
	for _, n := range nums {
		if n > high || n < low {
			return fmt.Errorf("%v is out of bounds [%v, %v]", n, low, high)
		}
	}
	return nil
}

// initMsr takes a slice of CPU IDs and initializes the msrReaderWithStorage from the receiver's msrBuilder configuration.
// If successfully initialized, it returns an msrReaderWithStorage. Otherwise, returns
// an error.
func (b *powerBuilder) initMsr(cpus []int) (msrReaderWithStorage, error) {
	if b.msr != nil {
		if err := b.msr.initMsrMap(cpus, b.msr.timeout); err != nil {
			return nil, err
		}
		return b.msr.msrReaderWithStorage, nil
	}
	return nil, nil
}

// initRapl initializes the raplReader from the receiver's raplBuilder configuration.
// If successfully initialized, it returns an raplReader. Otherwise, returns an error.
func (b *powerBuilder) initRapl() (raplReader, error) {
	if b.rapl != nil {
		if err := b.rapl.initZoneMap(); err != nil {
			return nil, err
		}
		return b.rapl.raplReader, nil
	}
	return nil, nil
}

// initCoreFreq initializes the cpuFreqReader from the receiver's coreFreqBuilder configuration.
// If successfully initialized, it returns a cpuFreqReader. Otherwise, returns an error.
func (b *powerBuilder) initCoreFreq() (cpuFreqReader, error) {
	if b.coreFreq != nil {
		if err := b.coreFreq.init(); err != nil {
			return nil, err
		}
		return b.coreFreq.cpuFreqReader, nil
	}
	return nil, nil
}

// initUncoreFreq initializes the uncoreFreqReader from the receiver's uncoreFreqBuilder configuration.
// If successfully initialized, it returns a cpuFreqReader. Otherwise, returns an error.
func (b *powerBuilder) initUncoreFreq() (uncoreFreqReader, error) {
	if b.uncoreFreq != nil {
		if err := b.uncoreFreq.init(); err != nil {
			return nil, err
		}
		return b.uncoreFreq.uncoreFreqReader, nil
	}
	return nil, nil
}

// initPerf takes a slice of perf events and a slice of CPU IDs. It initializes the perfReaderWithStorage
// from the receiver's perfBuilder configuration. If successfully initialized, it returns an perfReaderWithStorage.
// Otherwise, returns an error.
func (b *powerBuilder) initPerf(events []string, cpus []int) (perfReaderWithStorage, error) {
	if b.perf != nil {
		// check if processor supports perf hardware events for cstates
		model := b.topology.getCPUModel()
		if !isPerfAllowed(model) {
			return nil, fmt.Errorf("perf based metrics are not supported for processor model: 0x%X", model)
		}

		if err := b.perf.initResolver(b.perf.jsonPath); err != nil {
			return nil, fmt.Errorf("failed to init resolver: %w", err)
		}

		if err := b.perf.activate(events, cpus); err != nil {
			return nil, fmt.Errorf("failed to activate events: %w", err)
		}
		return b.perf.perfReaderWithStorage, nil
	}
	return nil, nil
}

// isPerfAllowed is helper function that returns true if the processor model supports hardware
// perf events specific to cstate residency metrics.
func isPerfAllowed(model int) bool {
	switch model {
	case cpumodel.INTEL_FAM6_SAPPHIRERAPIDS_X:
	case cpumodel.INTEL_FAM6_EMERALDRAPIDS_X:
		//TODO: Hybrid models are not supported right now
	//case cpumodel.INTEL_FAM6_ALDERLAKE:
	//case cpumodel.INTEL_FAM6_ALDERLAKE_L:
	//case cpumodel.INTEL_FAM6_RAPTORLAKE:
	//case cpumodel.INTEL_FAM6_RAPTORLAKE_P:
	//case cpumodel.INTEL_FAM6_RAPTORLAKE_S:
	//case cpumodel.INTEL_FAM6_METEORLAKE:
	//case cpumodel.INTEL_FAM6_METEORLAKE_L:
	//Above list should be updated in the future with new processors supporting the required events.
	default:
		return false
	}
	return true
}

// logTopologyDetails logs topology details such as CPU: vendor ID, family and model.
// It also logs core ID, package ID and die ID for every CPU ID.
func logTopologyDetails(t topologyReader) {
	var sb strings.Builder

	sb.WriteString("Topology details:\n")
	if vendorID, err := t.getCPUVendor(0); err != nil {
		sb.WriteString(fmt.Sprintf("  Error retrieving the CPU vendor ID: %v\n", err))
	} else {
		sb.WriteString(fmt.Sprintf("  CPU vendor ID: %s\n", vendorID))
	}

	if family, err := t.getCPUFamily(0); err != nil {
		sb.WriteString(fmt.Sprintf("  Error retrieving the CPU family: %v\n", err))
	} else {
		sb.WriteString(fmt.Sprintf("  CPU family: %s\n", family))
	}

	sb.WriteString(fmt.Sprintf("  CPU model: 0x%X\n", t.getCPUModel()))

	cpus := t.getCPUsNumber()
	sb.WriteString(fmt.Sprintf("  Number of CPUs: %d\n", cpus))
	for cpuID := 0; cpuID < cpus; cpuID++ {
		coreID, err := t.getCPUCoreID(cpuID)
		if err != nil {
			sb.WriteString(fmt.Sprintf("    Error retrieving the core ID: %v\n", err))
			continue
		}
		packageID, err := t.getCPUPackageID(cpuID)
		if err != nil {
			sb.WriteString(fmt.Sprintf("    Error retrieving the package ID: %v\n", err))
			continue
		}
		dieID, err := t.getCPUDieID(cpuID)
		if err != nil {
			sb.WriteString(fmt.Sprintf("    Error retrieving the die ID: %v\n", err))
			continue
		}

		sb.WriteString(fmt.Sprintf("    CPU ID: %4d, core ID: %4d, package ID: %2d, die ID: %2d\n", cpuID, coreID, packageID, dieID))
	}

	log.Debugf(sb.String())
}
