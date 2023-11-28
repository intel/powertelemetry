// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	ev "github.com/intel/iaevents"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPerf_C0StateType_String(t *testing.T) {
	t.Run("C01", func(t *testing.T) {
		c0State := c0StateType(0)
		require.Equal(t, "CPU_CLK_UNHALTED.C01", c0State.String())
	})

	t.Run("C02", func(t *testing.T) {
		c0State := c0StateType(1)
		require.Equal(t, "CPU_CLK_UNHALTED.C02", c0State.String())
	})

	t.Run("C0_Wait", func(t *testing.T) {
		c0State := c0StateType(2)
		require.Equal(t, "CPU_CLK_UNHALTED.C0_WAIT", c0State.String())
	})

	t.Run("Thread", func(t *testing.T) {
		c0State := c0StateType(3)
		require.Equal(t, "CPU_CLK_UNHALTED.THREAD", c0State.String())
	})

	t.Run("Invalid", func(t *testing.T) {
		c0State := c0StateType(4)
		require.Equal(t, "", c0State.String())
	})
}

type mockTransformer struct {
	mock.Mock
}

func (m *mockTransformer) Transform(reader ev.Reader, matcher ev.Matcher) ([]*ev.PerfEvent, error) {
	args := m.Called(reader, matcher)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ev.PerfEvent), args.Error(1)
}

func TestPerf_EventsResolver_ResolveEvent(t *testing.T) {
	mError := "error mock"
	mEventName := "Event.Mock"
	mTransformer := &mockTransformer{}
	mResolver := &eventsResolverImpl{
		transformer: mTransformer,
	}

	t.Run("TransformerIsNil", func(t *testing.T) {
		mResolver := &eventsResolverImpl{}

		_, err := mResolver.resolveEvent(mEventName)
		require.Error(t, err)
		require.ErrorContains(t, err, "transformer is nil")
	})

	t.Run("TransformReturnsError", func(t *testing.T) {
		mTransformer.On("Transform", nil, ev.NewNameMatcher(mEventName)).Return(nil, errors.New(mError)).Once()

		_, err := mResolver.resolveEvent(mEventName)
		require.Error(t, err)
		require.ErrorContains(t, err, mError)
		mTransformer.AssertExpectations(t)
	})

	t.Run("NoTransformedEvents", func(t *testing.T) {
		mCustom := ev.CustomizableEvent{}
		mTransformer.On("Transform", nil, ev.NewNameMatcher(mEventName)).Return([]*ev.PerfEvent{}, nil).Once()

		res, err := mResolver.resolveEvent(mEventName)
		require.Error(t, err)
		require.ErrorContains(t, err, "event could not be resolved")
		require.Equal(t, mCustom, res)
		mTransformer.AssertExpectations(t)
	})

	t.Run("EventSuccessfullyTransformed", func(t *testing.T) {
		mPerfEvent := &ev.PerfEvent{
			Name: mEventName,
		}
		mCustomEvent := ev.CustomizableEvent{
			Event: mPerfEvent,
		}

		mTransformer.On("Transform", nil, ev.NewNameMatcher(mEventName)).Return([]*ev.PerfEvent{mPerfEvent}, nil).Once()
		res, err := mResolver.resolveEvent(mEventName)

		require.NoError(t, err)
		require.Equal(t, res, mCustomEvent)
	})
}

func TestPerf_EventsResolver_ResolveEvents(t *testing.T) {
	t.Run("NoEvents", func(t *testing.T) {
		mTransformer := &mockTransformer{}
		var mResolver eventsResolver = &eventsResolverImpl{
			transformer: mTransformer,
		}

		customEvents, err := mResolver.resolveEvents(nil)
		require.Nil(t, customEvents)
		require.ErrorContains(t, err, "event group cannot be empty")
		mTransformer.AssertExpectations(t)
	})

	t.Run("FailedToResolveEvent", func(t *testing.T) {
		events := []string{
			"Event.Mock.1",
			"Event.Mock.2",
			"Event.Mock.3",
		}

		matcher1 := ev.NewNameMatcher(events[0])
		matcher2 := ev.NewNameMatcher(events[1])

		perfEvent := ev.PerfEvent{Name: events[0]}

		mTransformer := &mockTransformer{}
		mTransformer.On("Transform", nil, matcher1).Return([]*ev.PerfEvent{&perfEvent}, nil).Once()
		mTransformer.On("Transform", nil, matcher2).Return(nil, errors.New("mock error")).Once()

		var mResolver eventsResolver = &eventsResolverImpl{
			transformer: mTransformer,
		}

		customEvents, err := mResolver.resolveEvents(events)
		require.Nil(t, customEvents)
		require.ErrorContains(t, err, fmt.Sprintf("error resolving event %q", events[1]))
		mTransformer.AssertExpectations(t)
	})

	t.Run("EventsResolved", func(t *testing.T) {
		events := []string{
			"Event.Mock.1",
			"Event.Mock.2",
			"Event.Mock.3",
		}

		mTransformer := &mockTransformer{}
		customEventsExp := []ev.CustomizableEvent{}
		for _, event := range events {
			matcher := ev.NewNameMatcher(event)
			perfEvent := &ev.PerfEvent{Name: event}
			mTransformer.On("Transform", nil, matcher).Return([]*ev.PerfEvent{perfEvent}, nil).Once()

			customEventsExp = append(customEventsExp, ev.CustomizableEvent{Event: perfEvent})
		}

		var mResolver eventsResolver = &eventsResolverImpl{
			transformer: mTransformer,
		}

		customEvents, err := mResolver.resolveEvents(events)
		require.Equal(t, customEventsExp, customEvents)
		require.NoError(t, err)
		mTransformer.AssertExpectations(t)
	})
}

type mockValuesReader struct {
	mock.Mock
}

func (m *mockValuesReader) readValue(event *ev.ActiveEvent) (ev.CounterValue, error) {
	args := m.Called(event)
	return args.Get(0).(ev.CounterValue), args.Error(1)
}

type eventWithValues struct {
	event  *ev.ActiveEvent
	values ev.CounterValue
}

func TestPerf_EventsReader_ReadEvents(t *testing.T) {
	t.Run("NoActiveEvents", func(t *testing.T) {
		mReader := &mockValuesReader{}
		var mEventsReader eventsReader = &eventsReaderImpl{mReader}

		metrics, err := mEventsReader.readEvents(nil)
		require.Nil(t, metrics)
		require.ErrorContains(t, err, "no active events provided")
		mReader.AssertExpectations(t)
	})

	t.Run("InvalidActiveEvent", func(t *testing.T) {
		var mEventsReader eventsReader = &eventsReaderImpl{}
		activeEvents := []*ev.ActiveEvent{nil}

		metrics, err := mEventsReader.readEvents(activeEvents)
		require.Nil(t, metrics)
		require.ErrorContains(t, err, "invalid active event")

		activeEvents = []*ev.ActiveEvent{{PerfEvent: nil}}

		metrics, err = mEventsReader.readEvents(activeEvents)
		require.Nil(t, metrics)
		require.ErrorContains(t, err, "invalid active event")
	})

	t.Run("FailedToReadValue", func(t *testing.T) {
		mReader := &mockValuesReader{}
		var mEventsReader eventsReader = &eventsReaderImpl{mReader}

		mErr := errors.New("mock error")

		events := []string{
			"Event.Mock.1",
			"Event.Mock.2",
			"Event.Mock.3",
		}

		activeEvents := []*ev.ActiveEvent{}
		for _, event := range events {
			activeEvent := &ev.ActiveEvent{
				PerfEvent: &ev.PerfEvent{
					Name: event,
				},
			}
			activeEvents = append(activeEvents, activeEvent)
		}

		mReader.On("readValue", activeEvents[0]).Return(ev.CounterValue{}, nil).Once()
		mReader.On("readValue", activeEvents[1]).Return(ev.CounterValue{}, mErr).Once()
		mReader.On("readValue", activeEvents[2]).Return(ev.CounterValue{}, nil).Once()

		metrics, err := mEventsReader.readEvents(activeEvents)
		require.Nil(t, metrics)
		require.ErrorContains(t, err, fmt.Sprintf("failed to read values for event %q", activeEvents[1]))
		mReader.AssertExpectations(t)
	})

	t.Run("EventValuesRead", func(t *testing.T) {
		mReader := &mockValuesReader{}
		var mEventsReader eventsReader = &eventsReaderImpl{mReader}

		metricsExp := []coreMetric{}
		mEvents := []eventWithValues{
			{
				event: &ev.ActiveEvent{PerfEvent: &ev.PerfEvent{Name: "Event.1"}},
				values: ev.CounterValue{
					Raw:     123456789,
					Enabled: 1289175421,
					Running: 2374652324,
				},
			},
			{
				event: &ev.ActiveEvent{PerfEvent: &ev.PerfEvent{Name: "Event.2"}},
				values: ev.CounterValue{
					Raw:     987654321,
					Enabled: 4217641289,
					Running: 4901621382,
				},
			},
		}

		activeEvents := []*ev.ActiveEvent{}
		for _, activeEv := range mEvents {
			activeEvents = append(activeEvents, activeEv.event)

			cpu, _ := activeEv.event.PMUPlacement()
			metric := coreMetric{
				name:  activeEv.event.PerfEvent.Name,
				cpuID: cpu,

				values: activeEv.values,
			}
			metricsExp = append(metricsExp, metric)

			mReader.On("readValue", activeEv.event).Return(activeEv.values, nil).Once()
		}

		metricsOut, err := mEventsReader.readEvents(activeEvents)
		require.NoError(t, err)
		require.Equal(t, metricsExp, metricsOut)
		mReader.AssertExpectations(t)
	})
}

type mockPlacementMaker struct {
	mock.Mock
}

func (m *mockPlacementMaker) makeCorePlacement(cpuIDs []int, factory ev.PlacementFactory) ([]ev.PlacementProvider, error) {
	args := m.Called(cpuIDs, factory)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ev.PlacementProvider), args.Error(1)
}

type mockEventsActivator struct {
	mock.Mock
}

func (m *mockEventsActivator) activateEventsAsGroup(p ev.PlacementProvider, events []ev.CustomizableEvent) ([]*ev.ActiveEvent, error) {
	args := m.Called(p, events)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ev.ActiveEvent), args.Error(1)
}

func TestPerf_EventsActivator_ActivateEvents(t *testing.T) {
	var mEventsActivator eventsActivator

	mPlacementMaker := &mockPlacementMaker{}
	mActivator := &mockEventsActivator{}
	mEventsActivator = &eventsActivatorImpl{
		placementMaker: mPlacementMaker,
		perfActivator:  mActivator,
	}
	mError := errors.New("mock error")

	leader := &ev.PerfEvent{Name: "Event.Mock.1"}
	event := &ev.PerfEvent{Name: "Event.Mock.2"}

	customEvents := []ev.CustomizableEvent{
		{Event: leader},
		{Event: event},
	}
	placements := []ev.PlacementProvider{&ev.Placement{}, &ev.Placement{}}

	cores := []int{0, 1, 2, 3}

	activeEvents := []*ev.ActiveEvent{}
	for _, customEvent := range customEvents {
		activeEvent := &ev.ActiveEvent{PerfEvent: customEvent.Event}
		activeEvents = append(activeEvents, activeEvent)
	}

	t.Run("NoCustomEvents", func(t *testing.T) {
		activeEvents, err := mEventsActivator.activateEvents(nil, cores)
		require.Nil(t, activeEvents)
		require.ErrorContains(t, err, "no custom events provided")
	})

	t.Run("NoCores", func(t *testing.T) {
		activeEvents, err := mEventsActivator.activateEvents(customEvents, nil)
		require.Nil(t, activeEvents)
		require.ErrorContains(t, err, "no cores provided")
	})

	t.Run("FailedToMakePlacement", func(t *testing.T) {
		mPlacementMaker.On("makeCorePlacement", cores, leader).Return(nil, mError).Once()

		activeEvents, err := mEventsActivator.activateEvents(customEvents, cores)
		require.Nil(t, activeEvents)
		require.ErrorContains(t, err, fmt.Sprintf("failed to make core placements: %s", mError.Error()))
		mPlacementMaker.AssertExpectations(t)
	})

	t.Run("FailedToActivateEvents", func(t *testing.T) {
		activeEventsExp := []*ev.ActiveEvent{
			{PerfEvent: customEvents[0].Event},
			{PerfEvent: customEvents[1].Event},
		}
		mPlacementMaker.On("makeCorePlacement", cores, leader).Return(placements, nil).Once()
		mActivator.On("activateEventsAsGroup", placements[0], customEvents).Return(activeEventsExp, nil).Once()
		mActivator.On("activateEventsAsGroup", placements[1], customEvents).Return(nil, mError).Once()

		activeEventsOut, err := mEventsActivator.activateEvents(customEvents, cores)
		require.Equal(t, activeEventsExp, activeEventsOut)
		require.ErrorContains(t, err, "failed to activate events as a group")
		mPlacementMaker.AssertExpectations(t)
		mActivator.AssertExpectations(t)
	})

	t.Run("EventsActivated", func(t *testing.T) {
		activeEventsExp := []*ev.ActiveEvent{}

		mPlacementMaker.On("makeCorePlacement", cores, leader).Return(placements, nil).Once()
		for _, placement := range placements {
			mActivator.On("activateEventsAsGroup", placement, customEvents).Return(activeEvents, nil).Once()
			activeEventsExp = append(activeEventsExp, activeEvents...)
		}

		activeEventsOut, err := mEventsActivator.activateEvents(customEvents, cores)
		require.Equal(t, activeEventsExp, activeEventsOut)
		require.NoError(t, err)
		mPlacementMaker.AssertExpectations(t)
		mActivator.AssertExpectations(t)
	})
}

type mockPlacementFactory struct {
	err error
}

func (m *mockPlacementFactory) NewPlacements(_ string, cpu int, cpus ...int) ([]ev.PlacementProvider, error) {
	if m.err != nil {
		return nil, m.err
	}

	placements := make([]ev.PlacementProvider, 0)
	placements = append(placements, &ev.Placement{
		CPU:     cpu,
		PMUType: 4,
	})

	for _, cpu := range cpus {
		placements = append(placements, &ev.Placement{
			CPU:     cpu,
			PMUType: 4,
		})
	}
	return placements, nil
}

func TestPerf_PlacementMaker_MakeCorePlacement(t *testing.T) {
	mockError := errors.New("mock error")

	testCases := []struct {
		name      string
		cpuIDs    []int
		perfEvent ev.PlacementFactory
		expected  []ev.PlacementProvider
		err       error
	}{
		{
			name:      "NoCPUIDsProvided",
			cpuIDs:    nil,
			perfEvent: &ev.PerfEvent{},
			expected:  nil,
			err:       errors.New("no CPU IDs were provided"),
		},
		{
			name:      "SingleCorePlacementFailed",
			cpuIDs:    []int{0},
			perfEvent: &mockPlacementFactory{mockError},
			expected:  nil,
			err:       errors.New("failed to create single core placement"),
		},
		{
			name:      "MultipleCorePlacementFailed",
			cpuIDs:    []int{0, 1},
			perfEvent: &mockPlacementFactory{mockError},
			expected:  nil,
			err:       errors.New("failed to create multiple core placements"),
		},
		{
			name:      "SingleCorePlacement",
			cpuIDs:    []int{0},
			perfEvent: &mockPlacementFactory{nil},
			expected: []ev.PlacementProvider{
				&ev.Placement{CPU: 0, PMUType: 4},
			},
			err: nil,
		},
		{
			name:      "MultipleCorePlacements",
			cpuIDs:    []int{0, 1, 2, 3},
			perfEvent: &mockPlacementFactory{nil},
			expected: []ev.PlacementProvider{
				&ev.Placement{CPU: 0, PMUType: 4},
				&ev.Placement{CPU: 1, PMUType: 4},
				&ev.Placement{CPU: 2, PMUType: 4},
				&ev.Placement{CPU: 3, PMUType: 4},
			},
			err: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			maker := &placementMakerImpl{}
			providers, err := maker.makeCorePlacement(tc.cpuIDs, tc.perfEvent)

			require.Equal(t, tc.expected, providers)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type mockEventDeactivator struct {
	mock.Mock
}

func (m *mockEventDeactivator) deactivateEvent(event *ev.ActiveEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func TestPerf_EventsDeactivator_DeactivateEvents(t *testing.T) {
	events := []*ev.ActiveEvent{
		{PerfEvent: &ev.PerfEvent{Name: "Event.Mock.1", Uncore: false}},
		{PerfEvent: &ev.PerfEvent{Name: "Event.Mock.2", Uncore: false}},
		{PerfEvent: &ev.PerfEvent{Name: "Event.Mock.3", Uncore: false}},
		{PerfEvent: nil},
		nil,
	}

	t.Run("FailToDeactivate", func(t *testing.T) {
		mEventDeactivator := &mockEventDeactivator{}
		mEventDeactivator.On("deactivateEvent", events[0]).Return(nil).Once()
		mEventDeactivator.On("deactivateEvent", events[1]).Return(errors.New("mock error")).Once()
		mEventDeactivator.On("deactivateEvent", events[2]).Return(nil).Once()

		activeEventsExp := []*ev.ActiveEvent{events[1]}

		var mEventsDeactivator eventsDeactivator = &eventsDeactivatorImpl{
			perfDeactivator: mEventDeactivator,
		}

		activeEventsOut, err := mEventsDeactivator.deactivateEvents(events)
		require.Equal(t, activeEventsExp, activeEventsOut)
		require.ErrorContains(t, err, "failed to deactivate events")
		mEventDeactivator.AssertExpectations(t)
	})

	t.Run("EventsDeactivated", func(t *testing.T) {
		mEventDeactivator := &mockEventDeactivator{}
		mEventDeactivator.On("deactivateEvent", events[0]).Return(nil).Once()
		mEventDeactivator.On("deactivateEvent", events[1]).Return(nil).Once()
		mEventDeactivator.On("deactivateEvent", events[2]).Return(nil).Once()

		activeEventsExp := []*ev.ActiveEvent{}

		var mEventsDeactivator eventsDeactivator = &eventsDeactivatorImpl{
			perfDeactivator: mEventDeactivator,
		}

		activeEventsOut, err := mEventsDeactivator.deactivateEvents(events)
		require.Empty(t, activeEventsOut)
		require.Equal(t, activeEventsExp, activeEventsOut)
		require.NoError(t, err)
		mEventDeactivator.AssertExpectations(t)
	})
}

type mockFileInfoProvider struct {
	mock.Mock
}

func (m *mockFileInfoProvider) readFile(name string) ([]byte, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockFileInfoProvider) rlimit() (uint64, error) {
	args := m.Called()
	return args.Get(0).(uint64), args.Error(1)
}

func TestPerf_Helper_Multiply(t *testing.T) {
	t.Run("Overflow", func(t *testing.T) {
		maxUint64 := uint64(math.MaxUint64)
		b := uint64(10000)

		res, err := multiply(maxUint64, b)
		require.Equal(t, uint64(0), res)
		require.ErrorContains(t, err, "could not be represented as uint64")
	})

	t.Run("Valid", func(t *testing.T) {
		a := uint64(math.MaxUint64 >> 2)
		b := uint64(2)

		_, err := multiply(a, b)
		require.NoError(t, err)
	})
}

func TestPerf_Helper_GetMaxFd(t *testing.T) {
	t.Run("ReadFileMaxError", func(t *testing.T) {
		mError := errors.New("mock error")

		mFileInfoProvider := &mockFileInfoProvider{}
		mFileInfoProvider.On("readFile", fileMaxPath).Return([]byte{}, mError).Once()

		_, err := getMaxFd(mFileInfoProvider)
		require.ErrorContains(t, err, mError.Error())
		mFileInfoProvider.AssertExpectations(t)
	})

	t.Run("FileContentError", func(t *testing.T) {
		fileContent := []byte("invalid")

		mFileInfoProvider := &mockFileInfoProvider{}
		mFileInfoProvider.On("readFile", fileMaxPath).Return(fileContent, nil).Once()

		fd, err := getMaxFd(mFileInfoProvider)
		require.Equal(t, uint64(0), fd)
		require.ErrorContains(t, err, "could not parse file content")
		mFileInfoProvider.AssertExpectations(t)
	})

	t.Run("Valid", func(t *testing.T) {
		fdExp := uint64(25)
		fileContent := []byte(fmt.Sprintf("%d\n", fdExp))

		mFileInfoProvider := &mockFileInfoProvider{}
		mFileInfoProvider.On("readFile", fileMaxPath).Return(fileContent, nil).Once()

		fd, err := getMaxFd(mFileInfoProvider)
		require.Equal(t, fdExp, fd)
		require.NoError(t, err)
		mFileInfoProvider.AssertExpectations(t)
	})
}

func TestPerf_Helper_CheckFileDescriptor(t *testing.T) {
	t.Run("ReadHardLimitError", func(t *testing.T) {
		fd := uint64(25)
		fdMax := uint64(0)
		fileContent := []byte(fmt.Sprintf("%d\n", fdMax))

		mFileInfoProvider := &mockFileInfoProvider{}
		mFileInfoProvider.On("readFile", fileMaxPath).Return(fileContent, errors.New("mock error")).Once()

		err := checkFileDescriptors(fd, mFileInfoProvider)
		require.ErrorContains(t, err, "error retrieving kernel max file descriptors")
		mFileInfoProvider.AssertExpectations(t)
	})

	t.Run("HardLimitExceeded", func(t *testing.T) {
		fd := uint64(100)
		fdMax := uint64(25)
		fileContent := []byte(fmt.Sprintf("%d\n", fdMax))

		mFileInfoProvider := &mockFileInfoProvider{}
		mFileInfoProvider.On("readFile", fileMaxPath).Return(fileContent, nil).Once()

		err := checkFileDescriptors(fd, mFileInfoProvider)
		require.ErrorContains(t, err, fmt.Sprintf("required file descriptors %d, exceeds the maximum number of "+
			"available file descriptors %d", fd, fdMax))
		mFileInfoProvider.AssertExpectations(t)
	})

	t.Run("ReadSoftLimitError", func(t *testing.T) {
		fd := uint64(100)
		fdMax := uint64(125)
		fileContent := []byte(fmt.Sprintf("%d\n", fdMax))

		mError := errors.New("mock error")
		rlimit := uint64(0)

		mFileInfoProvider := &mockFileInfoProvider{}
		mFileInfoProvider.On("readFile", fileMaxPath).Return(fileContent, nil).Once()
		mFileInfoProvider.On("rlimit").Return(rlimit, mError).Once()

		err := checkFileDescriptors(fd, mFileInfoProvider)
		require.ErrorContains(t, err, "error retrieving process max file descriptors")
		mFileInfoProvider.AssertExpectations(t)
	})

	t.Run("SoftLimitExceeded", func(t *testing.T) {
		fd := uint64(25)
		fdMax := uint64(100)
		fileContent := []byte(fmt.Sprintf("%d\n", fdMax))
		rlimit := uint64(20)

		mFileInfoProvider := &mockFileInfoProvider{}
		mFileInfoProvider.On("readFile", fileMaxPath).Return(fileContent, nil).Once()
		mFileInfoProvider.On("rlimit").Return(rlimit, nil).Once()

		err := checkFileDescriptors(fd, mFileInfoProvider)
		require.Error(t, err, fmt.Sprintf("required file descriptors %d, exceeds the maximum number of"+
			"available file descriptors that a process may allocate %d", fd, rlimit))
		mFileInfoProvider.AssertExpectations(t)
	})

	t.Run("Valid", func(t *testing.T) {
		fd := uint64(25)
		fdMax := uint64(100)
		fileContent := []byte(fmt.Sprintf("%d\n", fdMax))
		rlimit := uint64(50)

		mFileInfoProvider := &mockFileInfoProvider{}
		mFileInfoProvider.On("readFile", fileMaxPath).Return(fileContent, nil).Once()
		mFileInfoProvider.On("rlimit").Return(rlimit, nil).Once()

		err := checkFileDescriptors(fd, mFileInfoProvider)
		require.NoError(t, err)
		mFileInfoProvider.AssertExpectations(t)
	})
}

func TestPerf_Perf_InitResolver(t *testing.T) {
	t.Run("WithValidJSONFile", func(t *testing.T) {
		perf := newPerf()
		require.NotNil(t, perf)
		require.NoError(t, perf.initResolver("testdata/sapphirerapids_core.json"))
	})

	t.Run("WithInvalidJSONFile", func(t *testing.T) {
		perf := newPerf()
		require.NotNil(t, perf)
		require.ErrorContains(t, perf.initResolver("dummy.json"), "error adding file to reader")
	})
}

type mockResolver struct {
	mock.Mock
}

func (m *mockResolver) resolveEvents(events []string) ([]ev.CustomizableEvent, error) {
	args := m.Called(events)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ev.CustomizableEvent), args.Error(1)
}

type mockActivator struct {
	mock.Mock
}

func (m *mockActivator) activateEvents(customEvents []ev.CustomizableEvent, cores []int) ([]*ev.ActiveEvent, error) {
	args := m.Called(customEvents, cores)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ev.ActiveEvent), args.Error(1)
}

func TestPerf_Perf_Activate(t *testing.T) {
	events := []string{
		"Event.Mock.1",
		"Event.Mock.2",
	}

	cores := []int{0, 1, 2, 3}

	customEvents := []ev.CustomizableEvent{}
	activeEvents := []*ev.ActiveEvent{}
	for _, event := range events {
		perfEvent := &ev.PerfEvent{Name: event, Uncore: false, PMUName: "cpu", PMUTypes: []ev.NamedPMUType{{Name: "cpu", PMUType: 4}}}
		customEvents = append(customEvents, ev.CustomizableEvent{
			Event: perfEvent,
		})

		activeEvents = append(activeEvents, &ev.ActiveEvent{
			PerfEvent: perfEvent,
		})
	}

	mError := errors.New("mock error")

	t.Run("FailedToResolve", func(t *testing.T) {
		mResolver := &mockResolver{}
		mResolver.On("resolveEvents", events).Return(nil, mError).Once()

		perf := &perf{
			resolver: mResolver,
		}

		require.ErrorContains(t, perf.activate(events, cores), "error resolving event")
		mResolver.AssertExpectations(t)
	})

	t.Run("FailedToCheckFileDescriptors", func(t *testing.T) {
		mResolver := &mockResolver{}
		mResolver.On("resolveEvents", events).Return(customEvents, nil).Once()

		mFileInfoProvider := &mockFileInfoProvider{}
		mFileInfoProvider.On("readFile", fileMaxPath).Return(nil, mError).Once()

		perf := &perf{
			resolver:       mResolver,
			fileInfoReader: mFileInfoProvider,
		}

		require.ErrorContains(t, perf.activate(events, cores), "error checking available file descriptors")
		mResolver.AssertExpectations(t)
		mFileInfoProvider.AssertExpectations(t)
	})

	t.Run("FailedToActivateEvents", func(t *testing.T) {
		mResolver := &mockResolver{}
		mResolver.On("resolveEvents", events).Return(customEvents, nil).Once()

		fdMax := uint64(10)
		fileContent := []byte(fmt.Sprintf("%d\n", fdMax))
		rlimit := uint64(10)

		mFileInfoProvider := &mockFileInfoProvider{}
		mFileInfoProvider.On("readFile", fileMaxPath).Return(fileContent, nil).Once()
		mFileInfoProvider.On("rlimit").Return(rlimit, nil).Once()

		// Only two events were activated
		activeEvents := []*ev.ActiveEvent{
			{PerfEvent: customEvents[0].Event},
			{PerfEvent: customEvents[1].Event},
		}
		mActivator := &mockActivator{}
		mActivator.On("activateEvents", customEvents, cores).Return(activeEvents, mError)

		perf := &perf{
			resolver:       mResolver,
			activator:      mActivator,
			fileInfoReader: mFileInfoProvider,
		}

		require.ErrorContains(t, perf.activate(events, cores), "error during event activation")
		require.Equal(t, activeEvents, perf.activeEvents)
		mResolver.AssertExpectations(t)
		mFileInfoProvider.AssertExpectations(t)
		mActivator.AssertExpectations(t)
	})

	t.Run("EventsActivated", func(t *testing.T) {
		mResolver := &mockResolver{}
		mResolver.On("resolveEvents", events).Return(customEvents, nil).Once()

		fdMax := uint64(10)
		fileContent := []byte(fmt.Sprintf("%d\n", fdMax))
		rlimit := uint64(10)

		mFileInfoProvider := &mockFileInfoProvider{}
		mFileInfoProvider.On("readFile", fileMaxPath).Return(fileContent, nil).Once()
		mFileInfoProvider.On("rlimit").Return(rlimit, nil).Once()

		mActivator := &mockActivator{}
		mActivator.On("activateEvents", customEvents, cores).Return(activeEvents, nil)

		perf := &perf{
			resolver:       mResolver,
			activator:      mActivator,
			fileInfoReader: mFileInfoProvider,
		}

		require.NoError(t, perf.activate(events, cores))
		require.Equal(t, activeEvents, perf.activeEvents)
		mResolver.AssertExpectations(t)
		mFileInfoProvider.AssertExpectations(t)
		mActivator.AssertExpectations(t)
	})
}

type mockEventsReader struct {
	mock.Mock
}

func (m *mockEventsReader) readEvents(events []*ev.ActiveEvent) ([]coreMetric, error) {
	args := m.Called(events)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]coreMetric), args.Error(1)
}

func TestPerf_Perf_Read(t *testing.T) {
	setFakeClock()
	fakeClock.Set(time.Now())
	defer unsetFakeClock()

	mEvents := []eventWithValues{
		{
			event: &ev.ActiveEvent{PerfEvent: &ev.PerfEvent{Name: "Event.1"}},
			values: ev.CounterValue{
				Raw:     123456789,
				Enabled: 1289175421,
				Running: 2374652324,
			},
		},
		{
			event: &ev.ActiveEvent{PerfEvent: &ev.PerfEvent{Name: "Event.2"}},
			values: ev.CounterValue{
				Raw:     987654321,
				Enabled: 4217641289,
				Running: 4901621382,
			},
		},
	}

	activeEvents := []*ev.ActiveEvent{}
	metricsExp := []coreMetric{}
	for _, mEvent := range mEvents {
		activeEvents = append(activeEvents, mEvent.event)

		cpu, _ := mEvent.event.PMUPlacement()
		metric := coreMetric{
			name:  mEvent.event.PerfEvent.Name,
			cpuID: cpu,

			values: mEvent.values,
		}
		metricsExp = append(metricsExp, metric)
	}

	t.Run("FailedToRead", func(t *testing.T) {
		mReader := &mockEventsReader{}
		mReader.On("readEvents", activeEvents).Return(nil, errors.New("mock error"))

		perf := &perf{
			valuesReader: mReader,
			activeEvents: activeEvents,
		}

		metrics, err := perf.read()
		require.Nil(t, metrics)
		require.ErrorContains(t, err, "mock error")
		mReader.AssertExpectations(t)
	})

	t.Run("EventValuesRead", func(t *testing.T) {
		mReader := &mockEventsReader{}
		mReader.On("readEvents", activeEvents).Return(metricsExp, nil)

		perf := &perf{
			valuesReader: mReader,
			activeEvents: activeEvents,
		}

		metricsOut, err := perf.read()
		require.Equal(t, metricsExp, metricsOut)
		require.NoError(t, err)
		mReader.AssertExpectations(t)
	})
}

type mockEventsDeactivator struct {
	mock.Mock
}

func (m *mockEventsDeactivator) deactivateEvents(events []*ev.ActiveEvent) ([]*ev.ActiveEvent, error) {
	args := m.Called(events)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ev.ActiveEvent), args.Error(1)
}

func TestPerf_Perf_Deactivate(t *testing.T) {
	events := []*ev.ActiveEvent{
		{PerfEvent: &ev.PerfEvent{Name: "Event.Mock.1", Uncore: false}},
		{PerfEvent: &ev.PerfEvent{Name: "Event.Mock.2", Uncore: false}},
		{PerfEvent: &ev.PerfEvent{Name: "Event.Mock.3", Uncore: false}},
		{PerfEvent: nil},
		nil,
	}

	t.Run("FailedToDeactivate", func(t *testing.T) {
		activeEventsExp := []*ev.ActiveEvent{events[1]}
		mError := fmt.Errorf("failed to deactivate events: %s", events[1].PerfEvent.Name)

		mDeactivator := &mockEventsDeactivator{}
		mDeactivator.On("deactivateEvents", events).Return(activeEventsExp, mError).Once()

		perf := &perf{
			deactivator: mDeactivator,

			activeEvents: events,
		}

		err := perf.deactivate()
		require.Equal(t, activeEventsExp, perf.activeEvents)
		require.ErrorContains(t, err, mError.Error())
		mDeactivator.AssertExpectations(t)
	})

	t.Run("EventsDeactivated", func(t *testing.T) {
		activeEventsExp := []*ev.ActiveEvent{}

		mDeactivator := &mockEventsDeactivator{}
		mDeactivator.On("deactivateEvents", events).Return(activeEventsExp, nil).Once()

		perf := &perf{
			deactivator: mDeactivator,

			activeEvents: events,
		}

		err := perf.deactivate()
		require.Equal(t, activeEventsExp, perf.activeEvents)
		require.NoError(t, err)
		mDeactivator.AssertExpectations(t)
	})
}

func TestPerf_PerfWithStorage_Update(t *testing.T) {
	t.Run("FailedToRead", func(t *testing.T) {
		mEventName := "Event.1"
		mActiveEvents := []*ev.ActiveEvent{{PerfEvent: &ev.PerfEvent{Name: mEventName}}}
		mReader := &mockEventsReader{}
		mReader.On("readEvents", mActiveEvents).Return(nil, errors.New("mock error")).Once()

		perfWithStorage := &perfWithStorage{
			perfReader: &perf{
				valuesReader: mReader,
				activeEvents: mActiveEvents,
			},
			metrics: make([]coreMetric, 0),
		}

		err := perfWithStorage.update()
		require.ErrorContains(t, err, "mock error")
		mReader.AssertExpectations(t)
	})

	t.Run("FailedToScale", func(t *testing.T) {
		mEventName := "Event.1"
		mActiveEvents := []*ev.ActiveEvent{{PerfEvent: &ev.PerfEvent{Name: mEventName}}}
		mMetrics := []coreMetric{
			{
				name:  mEventName,
				cpuID: 0,
				values: ev.CounterValue{
					Raw:     500,
					Enabled: math.MaxUint64,
					Running: 1,
				},
			},
		}

		metricsExp := []coreMetric{
			mMetrics[0],
		}
		metricsExp[0].scaled = 0

		mReader := &mockEventsReader{}
		mReader.On("readEvents", mActiveEvents).Return(mMetrics, nil).Once()

		perfWithStorage := &perfWithStorage{
			perfReader: &perf{
				valuesReader: mReader,
				activeEvents: mActiveEvents,
			},
			metrics: make([]coreMetric, 0),
		}

		err := perfWithStorage.update()
		require.Equal(t, metricsExp, perfWithStorage.metrics)
		require.ErrorContains(t, err, "scaled value could not be represented as uint64")
		mReader.AssertExpectations(t)
	})

	t.Run("UpdatedWithoutScaling", func(t *testing.T) {
		mEventName1 := "Event.1"
		mEventName2 := "Event.2"
		mActiveEvents := []*ev.ActiveEvent{
			{PerfEvent: &ev.PerfEvent{Name: mEventName1}},
			{PerfEvent: &ev.PerfEvent{Name: mEventName2}},
		}
		mMetrics := []coreMetric{
			{
				name:  "Event.1",
				cpuID: 0,
				values: ev.CounterValue{
					Raw:     881235,
					Enabled: 881235,
					Running: 881235,
				},
			},
			{
				name:  "Event.2",
				cpuID: 0,
				values: ev.CounterValue{
					Raw:     123456,
					Enabled: 123456,
					Running: 0,
				},
			},
		}

		metricsExp := []coreMetric{
			mMetrics[0],
			mMetrics[1],
		}
		metricsExp[0].scaled = mMetrics[0].values.Raw
		metricsExp[1].scaled = mMetrics[1].values.Raw

		mReader := &mockEventsReader{}
		mReader.On("readEvents", mActiveEvents).Return(mMetrics, nil).Once()

		perfWithStorage := &perfWithStorage{
			perfReader: &perf{
				valuesReader: mReader,
				activeEvents: mActiveEvents,
			},
			metrics: make([]coreMetric, 0),
		}

		err := perfWithStorage.update()
		require.Equal(t, metricsExp, perfWithStorage.metrics)
		require.NoError(t, err)
		mReader.AssertExpectations(t)
	})

	t.Run("UpdatedWithScaling", func(t *testing.T) {
		mEventName1 := "Event.1"
		mEventName2 := "Event.2"
		mActiveEvents := []*ev.ActiveEvent{
			{PerfEvent: &ev.PerfEvent{Name: mEventName1}},
			{PerfEvent: &ev.PerfEvent{Name: mEventName2}},
		}
		mMetrics := []coreMetric{
			{
				name:  mEventName1,
				cpuID: 0,
				values: ev.CounterValue{
					Raw:     123456789,
					Enabled: 1289175421,
					Running: 2374652324,
				},
			},
			{
				name:  mEventName2,
				cpuID: 0,
				values: ev.CounterValue{
					Raw:     987654321,
					Enabled: 4217641289,
					Running: 4901621382,
				},
			},
		}

		metricsExp := []coreMetric{
			mMetrics[0],
			mMetrics[1],
		}
		metricsExp[0].scaled = 67023478
		metricsExp[1].scaled = 849835456

		mReader := &mockEventsReader{}
		mReader.On("readEvents", mActiveEvents).Return(mMetrics, nil).Once()

		perfWithStorage := &perfWithStorage{
			perfReader: &perf{
				valuesReader: mReader,
				activeEvents: mActiveEvents,
			},
			metrics: make([]coreMetric, 0),
		}

		err := perfWithStorage.update()
		require.Equal(t, metricsExp, perfWithStorage.metrics)
		require.NoError(t, err)
		mReader.AssertExpectations(t)
	})
}

func TestPerf_PerfWithStorage_GetCoreMetrics(t *testing.T) {
	metrics := []coreMetric{
		{name: "Mock.Event.1", cpuID: 1},
		{name: "Mock.Event.1", cpuID: 2},
		{name: "Mock.Event.2", cpuID: 1},
		{name: "Mock.Event.3", cpuID: 2},
	}

	var perf perfReaderWithStorage = &perfWithStorage{
		metrics: metrics,
	}

	t.Run("CoreNotFound", func(t *testing.T) {
		metricsOut := perf.getCoreMetrics(0)
		require.Equal(t, []coreMetric{}, metricsOut)
	})

	t.Run("CoreMetricsFound", func(t *testing.T) {
		metricsExp := []coreMetric{metrics[0], metrics[2]}
		metricsOut := perf.getCoreMetrics(1)
		require.Equal(t, metricsExp, metricsOut)
	})
}
