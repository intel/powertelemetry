// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"syscall"

	ev "github.com/intel/iaevents"
	"golang.org/x/sync/errgroup"
)

// File path which monitors the maximum number of file-handles that the Linux
// kernel can allocate.
const fileMaxPath = "/proc/sys/fs/file-max"

// fileInfoProvider reads contents of files and provides the maximum number of
// file descriptors that a process may allocate.
// TODO: Consider to move rlimit into a new single method interface.
// TODO: Move this interface to a separate file.
type fileInfoProvider interface {
	// readFile reads the contents of a file.
	readFile(path string) ([]byte, error)

	// rlimit returns the maximum number of file descriptors that a process may allocate.
	rlimit() (uint64, error)
}

// fsHelper implements fileInfoProvider interface.
type fsHelper struct{}

// readFile reads the contents of the given file path.
func (*fsHelper) readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// rlimit returns the maximum number of file descriptors that a process may allocate.
// It makes a syscall to get RLIMIT_NOFILE property.
func (*fsHelper) rlimit() (uint64, error) {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	return rLimit.Cur, err
}

// getMaxFd is a helper function that takes a fileInfoProvider interface and returns
// the maximum number of file-handles that the Linux kernel can allocate.
func getMaxFd(fp fileInfoProvider) (uint64, error) {
	buf, err := fp.readFile(fileMaxPath)
	if err != nil {
		return 0, fmt.Errorf("could not read file %q: %w", fileMaxPath, err)
	}

	maxFd, err := strconv.ParseUint(string(bytes.TrimRight(buf, "\n")), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse file content %v to uint64: %w", maxFd, err)
	}

	return maxFd, nil
}

// checkFileDescriptors is a helper function that takes the number of estimated file descriptors
// needed and a fileInfoProvider, and returns nil if the number of estimated file descriptors does
// not exceed the maximum number of file-handles that the Linux kernel can allocate. Otherwise, returns
// an error.
// TODO: Add information about max number of file handles into README.md.
func checkFileDescriptors(fd uint64, reader fileInfoProvider) error {
	maxFd, err := getMaxFd(reader)
	if err != nil {
		return fmt.Errorf("error retrieving kernel max file descriptors: %w", err)
	}

	if fd > maxFd {
		return fmt.Errorf("required file descriptors %d, exceeds the maximum number of available file descriptors %d", fd, maxFd)
	}

	limit, err := reader.rlimit()
	if err != nil {
		return fmt.Errorf("error retrieving process max file descriptors: %w", err)
	}

	if fd > limit {
		return fmt.Errorf("required file descriptors %d, exceeds the maximum number of available file descriptors that a process may allocate %d", fd, limit)
	}
	return nil
}

// multiply is a helper function that calculates the product of two uint64 values.
// If overflow occurs, it returns an error.
func multiply(a, b uint64) (uint64, error) {
	bigA := new(big.Int).SetUint64(a)
	bigB := new(big.Int).SetUint64(b)

	res := new(big.Int).Mul(bigA, bigB)
	if !res.IsUint64() {
		return 0, fmt.Errorf("value could not be represented as uint64: %v", res)
	}
	return res.Uint64(), nil
}

// c0StateType is an enum type to identify event names corresponding to C0 substate metrics.
type c0StateType int

// c0StateType enum defines supported event names for C0 substate metrics.
const (
	c01 c0StateType = iota
	c02
	c0Wait
	thread
)

// Helper function to return a string representation of c0StateType.
func (t c0StateType) String() string {
	switch t {
	case c01:
		return "CPU_CLK_UNHALTED.C01"
	case c02:
		return "CPU_CLK_UNHALTED.C02"
	case c0Wait:
		return "CPU_CLK_UNHALTED.C0_WAIT"
	case thread:
		return "CPU_CLK_UNHALTED.THREAD"
	default:
		return ""
	}
}

// coreMetric represents the values of a core event read at a specific time instant.
type coreMetric struct {
	name  string
	cpuID int

	values ev.CounterValue
	scaled uint64
}

// eventsResolver resolves event names, from a core event group, to custom events which
// be activated.
type eventsResolver interface {
	resolveEvents(events []string) ([]ev.CustomizableEvent, error)
}

// eventsResolverImpl implements eventsResolver interface.
type eventsResolverImpl struct {
	reader      ev.Reader
	transformer ev.Transformer
}

// resolveEvents takes a core event group with event names and resolves them into
// custom events that can be activated.
func (r *eventsResolverImpl) resolveEvents(events []string) ([]ev.CustomizableEvent, error) {
	if len(events) == 0 {
		return nil, errors.New("event group cannot be empty")
	}

	customEvents := make([]ev.CustomizableEvent, len(events))
	for i, event := range events {
		var err error
		customEvents[i], err = r.resolveEvent(event)
		if err != nil {
			return nil, fmt.Errorf("error resolving event %q: %w", event, err)
		}
	}
	return customEvents, nil
}

// resolveEvent takes an event name string and returns a custom event that be activated.
func (r *eventsResolverImpl) resolveEvent(name string) (ev.CustomizableEvent, error) {
	if r.transformer == nil {
		return ev.CustomizableEvent{}, errors.New("transformer is nil")
	}

	perfEvents, err := r.transformer.Transform(r.reader, ev.NewNameMatcher(name))
	if err != nil {
		return ev.CustomizableEvent{}, err
	}

	if len(perfEvents) == 0 {
		return ev.CustomizableEvent{}, errors.New("event could not be resolved")
	}

	return ev.CustomizableEvent{
		Event: perfEvents[0],
	}, nil
}

// placementMaker takes a slice of cores and an event, which is the leader of
// the event group, and returns core placements needed for activation of each
// event of the group.
type placementMaker interface {
	makeCorePlacement(cpuIDs []int, factory ev.PlacementFactory) ([]ev.PlacementProvider, error)
}

// placementMakerImpl implements placementMaker interface.
type placementMakerImpl struct{}

// makeCorePlacement takes a slice of cores and makes core placements for the given
// PlacementProvider.
func (*placementMakerImpl) makeCorePlacement(cpuIDs []int, factory ev.PlacementFactory) ([]ev.PlacementProvider, error) {
	var cpuPlacements []ev.PlacementProvider
	var err error

	switch len(cpuIDs) {
	case 0:
		return nil, errors.New("no CPU IDs were provided")
	case 1:
		cpuPlacements, err = ev.NewCorePlacements(factory, cpuIDs[0])
		if err != nil {
			return nil, fmt.Errorf("failed to create single core placement: %w", err)
		}
	default:
		cpuPlacements, err = ev.NewCorePlacements(factory, cpuIDs[0], cpuIDs[1:]...)
		if err != nil {
			return nil, fmt.Errorf("failed to create multiple core placements: %w", err)
		}
	}
	return cpuPlacements, nil
}

// eventGroupActivator activates custom core events using the given core PlacementProvider.
type eventGroupActivator interface {
	activateEventsAsGroup(p ev.PlacementProvider, events []ev.CustomizableEvent) ([]*ev.ActiveEvent, error)
}

// eventGroupActivatorImpl implements eventGroupActivator interface.
type eventGroupActivatorImpl struct{}

// activateEventsAsGroup takes a core PlacementProvider and a slice of custom events, and
// returns a slice of events which have been successfully activated.
func (*eventGroupActivatorImpl) activateEventsAsGroup(p ev.PlacementProvider, events []ev.CustomizableEvent) ([]*ev.ActiveEvent, error) {
	activeEventGroup, err := ev.ActivateGroup(p, ev.NewEventTargetProcess(-1, 0), events)
	return activeEventGroup.Events(), err
}

// eventsActivator activates a group of core events.
type eventsActivator interface {
	activateEvents(customEvents []ev.CustomizableEvent, cores []int) ([]*ev.ActiveEvent, error)
}

// eventsActivatorImpl implements eventsActivator interface.
type eventsActivatorImpl struct {
	placementMaker placementMaker
	perfActivator  eventGroupActivator
}

// activateGroup takes a group of core events and activates them.
func (a *eventsActivatorImpl) activateEvents(customEvents []ev.CustomizableEvent, cores []int) ([]*ev.ActiveEvent, error) {
	if len(customEvents) == 0 {
		return nil, errors.New("no custom events provided")
	}

	if len(cores) == 0 {
		return nil, errors.New("no cores provided")
	}

	leader := customEvents[0]
	placements, err := a.placementMaker.makeCorePlacement(cores, leader.Event)
	if err != nil {
		return nil, fmt.Errorf("failed to make core placements: %w", err)
	}

	activeEvents := make([]*ev.ActiveEvent, 0)
	for _, placement := range placements {
		events, err := a.perfActivator.activateEventsAsGroup(placement, customEvents)
		if err != nil {
			return activeEvents, fmt.Errorf("failed to activate events as a group: %w", err)
		}
		activeEvents = append(activeEvents, events...)
	}
	return activeEvents, nil
}

// valuesReader reads values of an active core event.
type valuesReader interface {
	readValue(event *ev.ActiveEvent) (ev.CounterValue, error)
}

// valuesReaderImpl implements valuesReader interface.
type valuesReaderImpl struct{}

// readValue takes an active event and returns its values.
// It is a wrapper of ReadValue method of an ev.ActiveEvent value type.
func (*valuesReaderImpl) readValue(event *ev.ActiveEvent) (ev.CounterValue, error) {
	return event.ReadValue()
}

// eventsReader reads the values of a group of active core events.
type eventsReader interface {
	readEvents(events []*ev.ActiveEvent) ([]coreMetric, error)
}

// eventsReaderImpl implements eventsReader interface.
type eventsReaderImpl struct {
	eventReader valuesReader
}

// readEvents takes a group of active core events and returns a slice of coreMetrics.
// Each coreMetric has read values specific for an event name and core.
// TODO: Rework implementation to accept context propagated from top of the call stack.
func (r *eventsReaderImpl) readEvents(events []*ev.ActiveEvent) ([]coreMetric, error) {
	if len(events) == 0 {
		return nil, errors.New("no active events provided")
	}

	metrics := make([]coreMetric, len(events))
	errGroup := errgroup.Group{}
	for i, event := range events {
		if event == nil || event.PerfEvent == nil {
			return nil, errors.New("invalid active event")
		}

		index := i
		activeEvent := event

		errGroup.Go(func() error {
			values, err := r.eventReader.readValue(activeEvent)
			if err != nil {
				return fmt.Errorf("failed to read values for event %q: %w", activeEvent, err)
			}

			cpu, _ := activeEvent.PMUPlacement()
			metrics[index] = coreMetric{
				values: values,
				cpuID:  cpu,
				name:   activeEvent.PerfEvent.Name,
			}
			return nil
		})
	}

	if err := errGroup.Wait(); err != nil {
		return nil, err
	}
	return metrics, nil
}

// eventDeactivator deactivates an active core event.
type eventDeactivator interface {
	deactivateEvent(event *ev.ActiveEvent) error
}

// eventDeactivatorImpl implements eventDeactivator interface.
type eventDeactivatorImpl struct{}

// deactivateEvent takes an active core event and deactivates it. If the event could not
// be deactivated successfully an error is returned. This method is a wrapper of Deactivate
// method of ev.ActiveEvent value type.
func (*eventDeactivatorImpl) deactivateEvent(event *ev.ActiveEvent) error {
	return event.Deactivate()
}

// eventsDeactivator deactivates a group of active core events.
type eventsDeactivator interface {
	deactivateEvents(events []*ev.ActiveEvent) ([]*ev.ActiveEvent, error)
}

// eventsDeactivatorImpl implements eventsDeactivator interface.
type eventsDeactivatorImpl struct {
	perfDeactivator eventDeactivator
}

// deactivateEvents takes a slice of active core events and deactivates them.
func (d *eventsDeactivatorImpl) deactivateEvents(events []*ev.ActiveEvent) ([]*ev.ActiveEvent, error) {
	var err error
	failedToDeactivate := make([]string, 0)
	activeEvents := make([]*ev.ActiveEvent, 0)

	for _, event := range events {
		if event == nil || event.PerfEvent == nil {
			continue
		}

		if err := d.perfDeactivator.deactivateEvent(event); err != nil {
			failedToDeactivate = append(failedToDeactivate, event.PerfEvent.Name)
			activeEvents = append(activeEvents, event)
		}
	}

	if len(failedToDeactivate) != 0 {
		err = fmt.Errorf("failed to deactivate events: %q", strings.Join(failedToDeactivate, ", "))
	}
	return activeEvents, err
}

// perfReader activates, reads and deactivates groups of core events accessible via `perf_events`
// kernel interface.
type perfReader interface {
	initResolver(jsonFile string) error

	activate(events []string, cores []int) error

	read() ([]coreMetric, error)

	deactivate() error
}

// perf implements perfReader interface. It keeps track of the current active events.
type perf struct {
	resolver       eventsResolver
	activator      eventsActivator
	deactivator    eventsDeactivator
	valuesReader   eventsReader
	fileInfoReader fileInfoProvider

	activeEvents []*ev.ActiveEvent
}

// newPerf takes a path string, corresponding to a JSON file which comprises processor model
// specific events.
func newPerf() perfReader {
	return &perf{
		activator: &eventsActivatorImpl{
			placementMaker: &placementMakerImpl{},
			perfActivator:  &eventGroupActivatorImpl{},
		},
		deactivator:    &eventsDeactivatorImpl{&eventDeactivatorImpl{}},
		valuesReader:   &eventsReaderImpl{&valuesReaderImpl{}},
		fileInfoReader: &fsHelper{},
	}
}

func (p *perf) initResolver(jsonFile string) error {
	reader := ev.NewFilesReader()
	if err := reader.AddFiles(jsonFile); err != nil {
		return fmt.Errorf("error adding file to reader: %w", err)
	}

	p.resolver = &eventsResolverImpl{
		reader:      reader,
		transformer: ev.NewPerfTransformer(),
	}
	return nil
}

// activate takes a slice of core event names and cores. It resolves the given event
// names into perf events and activates them. If number of file descriptors needed to
// read the events it returns an error.
// TODO: Do not receive events from arguments.
func (p *perf) activate(events []string, cores []int) error {
	// resolve
	customEvents, err := p.resolver.resolveEvents(events)
	if err != nil {
		return fmt.Errorf("error resolving event: %w", err)
	}

	// calculate file descriptors needed to access all events
	numEvents := uint64(len(customEvents))
	numCores := uint64(len(cores))
	fd, err := multiply(numEvents, numCores)
	if err != nil {
		return err
	}

	// check maximum allowed number of file descriptors
	err = checkFileDescriptors(fd, p.fileInfoReader)
	if err != nil {
		return fmt.Errorf("error checking available file descriptors: %w", err)
	}

	// activate
	p.activeEvents, err = p.activator.activateEvents(customEvents, cores)
	if err != nil {
		return fmt.Errorf("error during event activation: %w", err)
	}
	return nil
}

// deactivate deactivates all active events. If an event or events could not
// be successfully deactivated, an error is returned.
func (p *perf) deactivate() error {
	var err error
	p.activeEvents, err = p.deactivator.deactivateEvents(p.activeEvents)
	return err
}

// read performs a single read of all active events and returns a slice with the metrics for each one.
// Events need to be activated previously by calling resolve method.
// TODO: Rework implementation to accept context propagated from top of the call stack.
func (p *perf) read() ([]coreMetric, error) {
	return p.valuesReader.readEvents(p.activeEvents)
}

// perfReaderWithStorage decorates perfReader with the ability to store core event read
// values and to retrieve all metrics that belong to a specific CPU ID.
type perfReaderWithStorage interface {
	perfReader

	update() error

	getCoreMetrics(cpuID int) []coreMetric
}

// perfWithStorage implements perfReaderWithStorage interface. The content of metrics field
// are the core event values read from the last call to read method.
type perfWithStorage struct {
	// TODO: Evaluate implications of either embedding perf or perfReader
	perfReader

	metrics []coreMetric
}

// update reads values for active core events specified by the receiver. It updates the metrics
// field with the latest values returned by read method and calculates scaled value of a metric.
func (p *perfWithStorage) update() error {
	var err error
	p.metrics, err = p.read()
	if err != nil {
		return err
	}

	for i := range p.metrics {
		p.metrics[i].scaled, err = scaleMetricValues(p.metrics[i].values)
		if err != nil {
			return err
		}
	}
	return nil
}

// scaleMetricValues calculates scaled value from metric values. Scaled value is equal to
// raw * enabled / running. If running value is equal to 0, then the raw value will be returned.
func scaleMetricValues(values ev.CounterValue) (uint64, error) {
	enabledBig := new(big.Int).SetUint64(values.Enabled)
	runningBig := new(big.Int).SetUint64(values.Running)
	rawBig := new(big.Int).SetUint64(values.Raw)

	if values.Enabled != values.Running && values.Running != uint64(0) {
		product := new(big.Int).Mul(rawBig, enabledBig)
		scaled := new(big.Int).Div(product, runningBig)

		if !scaled.IsUint64() {
			return 0, fmt.Errorf("scaled value could not be represented as uint64: %v", scaled)
		}
		return scaled.Uint64(), nil
	}
	return rawBig.Uint64(), nil
}

// getCoreMetrics takes a CPU ID as argument and returns all core metrics specific to this core
// stored in metrics field.
func (p *perfWithStorage) getCoreMetrics(cpuID int) []coreMetric {
	metrics := make([]coreMetric, 0)
	for _, metric := range p.metrics {
		if metric.cpuID == cpuID {
			metrics = append(metrics, metric)
		}
	}

	return metrics
}
