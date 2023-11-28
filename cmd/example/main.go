// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	powertelemetry "github.com/intel/powertelemetry"
	"github.com/intel/powertelemetry/internal/version"
)

const (
	interval = 5 * time.Second  // sample interval in seconds
	duration = 26 * time.Second // duration of each experiment in seconds
)

func main() {
	logger := simpleLogger{}

	// Print the current version of the application
	logger.Infof("Using: %s", version.GetFullVersion())

	// TODO: Add logic to parse CPU IDs (and package IDs?) from command line
	cpuID := 0
	packageID := 0
	dieID := 0
	includedCPUs := []int{cpuID}

	pt, err := powertelemetry.New(
		powertelemetry.WithLogger(&logger),
		// powertelemetry.WithExcludedCPUs(excludedCPUs),
		powertelemetry.WithIncludedCPUs(includedCPUs),
		powertelemetry.WithMsr(),
		powertelemetry.WithRapl(),
		powertelemetry.WithCoreFrequency(),
		powertelemetry.WithUncoreFrequency(),
		//powertelemetry.WithPerf(""),
	)

	var initErr *powertelemetry.MultiError
	if err != nil {
		if !errors.As(err, &initErr) {
			logger.Errorf("Failed to build powertelemetry instance: %v", err)
			os.Exit(1)
		}
		logger.Warn(err)
	}

	//
	// Per CPU ID metrics
	//
	logger.Info("=== Per CPU ID metrics ===")

	// CPU current frequency metric
	cpuFreq, err := pt.GetCPUFrequency(cpuID)
	if err != nil {
		logger.Errorf("Error getting current frequency for CPU ID %v: %v", cpuID, err)
	} else {
		logger.Infof("CPU ID: %v, CPU current frequency[MHz]: %0.1f", cpuID, cpuFreq)
	}

	// CPU temperature metric
	cpuTemp, err := pt.GetCPUTemperature(cpuID)
	if err != nil {
		logger.Errorf("Error getting temperature for CPU ID %v: %v", cpuID, err)
	} else {
		logger.Infof("CPU ID: %v, CPU temperature[°C]: %v", cpuID, cpuTemp)
	}

	//
	// CPU MSR time-based metrics:
	//
	// * CPU C0/C1/C3/C6/C7 state residency
	// * CPU busy frequency
	//
	logger.Info("=== CPU time-based metrics ===")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	getCPUMSRMetrics := func() (string, error) {
		return func(cpuID int) (string, error) {
			if err := pt.UpdatePerCPUMetrics(cpuID); err != nil {
				return "", fmt.Errorf("error gathering per CPU metrics for CPU ID %v: %w", cpuID, err)
			}
			c0State, err := pt.GetCPUC0StateResidency(cpuID)
			if err != nil {
				return "", fmt.Errorf("error getting CPU C0 state residency for CPU ID %v: %w", cpuID, err)
			}
			c1State, err := pt.GetCPUC1StateResidency(cpuID)
			if err != nil {
				return "", fmt.Errorf("error getting CPU C1 state residency for CPU ID %v: %w", cpuID, err)
			}
			c3State, err := pt.GetCPUC3StateResidency(cpuID)
			if err != nil {
				return "", fmt.Errorf("error getting CPU C3 state residency for CPU ID %v: %w", cpuID, err)
			}
			c6State, err := pt.GetCPUC6StateResidency(cpuID)
			if err != nil {
				return "", fmt.Errorf("error getting CPU C6 state residency for CPU ID %v: %w", cpuID, err)
			}
			c7State, err := pt.GetCPUC7StateResidency(cpuID)
			if err != nil {
				return "", fmt.Errorf("error getting CPU C7 state residency for CPU ID %v: %w", cpuID, err)
			}
			busyFreq, err := pt.GetCPUBusyFrequencyMhz(cpuID)
			if err != nil {
				return "", fmt.Errorf("error getting CPU busy frequency for CPU ID %v: %w", cpuID, err)
			}
			return fmt.Sprintf("CPU ID: %v, C0[%%]: %.4f, C1[%%]: %.4f, C3[%%]: %.4f, C6[%%]: %.4f, C7[%%]: %.4f, Busy freq[MHz]: %.4f",
				cpuID, c0State, c1State, c3State, c6State, c7State, busyFreq), nil
		}(cpuID)
	}

	err = printMetricsIteratively(ctx, logger, getCPUMSRMetrics)
	if err != nil {
		logger.Errorf("Error while getting CPU MSR metrics: %v", err)
	}

	//
	// Per package ID and die ID metrics
	//
	logger.Info("=== Per package ID metrics ===")

	// CPU base frequency metric
	cpuBaseFreq, err := pt.GetCPUBaseFrequency(packageID)
	if err != nil {
		logger.Errorf("Error getting CPU base frequency: %v", err)
	} else {
		logger.Infof("Package ID: %v, CPU base frequency[MHz]: %v", packageID, cpuBaseFreq)
	}

	// Package uncore frequency limits
	// Package customized uncore maximum frequency
	customizedMaxFreq, err := pt.GetCustomizedUncoreFrequencyMax(packageID, dieID)
	if err != nil {
		logger.Errorf("Error reading customized max frequency of packageID: %v, dieID: %v: %v", packageID, dieID, err)
	} else {
		logger.Infof("Package ID: %v, die ID: %v, customized uncore frequency max[MHz]: %v", packageID, dieID, customizedMaxFreq)
	}

	// Package customized uncore minimum frequency
	customizedMinFreq, err := pt.GetCustomizedUncoreFrequencyMin(packageID, dieID)
	if err != nil {
		logger.Errorf("Error reading customized min frequency of packageID: %v, dieID: %v: %v", packageID, dieID, err)
	} else {
		logger.Infof("Package ID: %v, die ID: %v, customized uncore frequency min[MHz]: %v", packageID, dieID, customizedMinFreq)
	}

	// Package initial uncore maximum frequency
	initialMaxFreq, err := pt.GetInitialUncoreFrequencyMax(packageID, dieID)
	if err != nil {
		logger.Errorf("Error reading initial max frequency of packageID: %v, dieID: %v: %v", packageID, dieID, err)
	} else {
		logger.Infof("Package ID: %v, die ID: %v, initial uncore frequency max[MHz]: %v", packageID, dieID, initialMaxFreq)
	}

	// Package initial uncore minimum frequency
	initialMinFreq, err := pt.GetInitialUncoreFrequencyMin(packageID, dieID)
	if err != nil {
		logger.Errorf("Error reading initial min frequency of packageID: %v, dieID: %v: %v", packageID, dieID, err)
	} else {
		logger.Infof("Package ID: %v, die ID: %v, initial uncore frequency min[MHz]: %v", packageID, dieID, initialMinFreq)
	}

	// Package current uncore frequency metric
	currentFreq, err := pt.GetCurrentUncoreFrequency(packageID, dieID)
	if err != nil {
		logger.Errorf("Error reading current frequency of packageID: %v, dieID: %v: %v", packageID, dieID, err)
	} else {
		logger.Infof("Package ID: %v, die ID: %v, current uncore frequency[MHz]: %v", packageID, dieID, currentFreq)
	}

	// Package thermal design power metric
	thermalDesignPower, err := pt.GetPackageThermalDesignPowerWatts(packageID)
	if err != nil {
		logger.Errorf("Error getting thermal design power for package ID %v: %v", packageID, err)
	} else {
		logger.Infof("Package ID: %v, thermal design power[W]: %v", packageID, thermalDesignPower)
	}

	maxTurboFreqList, err := pt.GetMaxTurboFreqList(packageID)
	if err != nil {
		logger.Errorf("Error getting max turbo frequency limit list: %v", err)
	} else {
		for _, v := range maxTurboFreqList {
			str := fmt.Sprintf("Package ID: %v, die ID: %v, max turbo frequency: %v MHz, active cores: %v", packageID, dieID, v.Value, v.ActiveCores)
			if v.Secondary {
				str += ", secondary"
			}
			logger.Info(str)
		}
	}

	//
	// Current power consumption metrics:
	//
	// * Package current power consumption
	// * Package DRAM current power consumption
	//
	logger.Info("=== Current Power Consumption ===")

	getPowerConsumptionMetrics := func() (string, error) {
		return func(packageID int) (string, error) {
			packageCurrPower, err := pt.GetCurrentPackagePowerConsumptionWatts(packageID)
			if err != nil {
				return "", fmt.Errorf("error getting package power consumption for package ID %v: %w", packageID, err)
			}
			dramCurrPower, err := pt.GetCurrentDramPowerConsumptionWatts(packageID)
			if err != nil {
				return "", fmt.Errorf("error getting dram power consumption for package ID %v: %w", packageID, err)
			}
			return fmt.Sprintf("PackageID: %v, package[W]: %.4f, dram[W]: %.4f", packageID, packageCurrPower, dramCurrPower), nil
		}(packageID)
	}

	err = printMetricsIteratively(ctx, logger, getPowerConsumptionMetrics)
	if err != nil {
		logger.Errorf("Error while getting power consumption metrics: %v", err)
	}

	// CPU flag support
	supported, err := pt.IsFlagSupported("msr")
	if err != nil {
		logger.Errorf("Error while checking if flag is supported by first CPU: %v", err)
	} else {
		logger.Infof("Is flag 'msr' supported for first CPU: %t", supported)
	}
}

func printMetricsIteratively(ctx context.Context, logger simpleLogger, getMetrics func() (string, error)) error {
	tInterval := time.NewTicker(interval)
	tDuration := time.NewTicker(duration)
	count := 0

	for {
		select {
		case <-ctx.Done():
			tInterval.Stop()
			tDuration.Stop()
			return ctx.Err()
		case <-tDuration.C:
			tInterval.Stop()
			tDuration.Stop()
			return nil
		case <-tInterval.C:
			count++
			if line, err := getMetrics(); err != nil {
				logger.Errorf("Error fetching metrics: %v", err)
			} else {
				logger.Infof("Sample %v: %q", count, line)
			}
		}
	}
}

type simpleLogger struct {
}

func (l *simpleLogger) Debugf(format string, args ...interface{}) {
	log.Printf("D! "+format, args...)
}

func (l *simpleLogger) Debug(args ...interface{}) {
	log.Print(append([]interface{}{"D! "}, args...)...)
}

func (l *simpleLogger) Infof(format string, args ...interface{}) {
	log.Printf("I! "+format, args...)
}

func (l *simpleLogger) Info(args ...interface{}) {
	log.Print(append([]interface{}{"I! "}, args...)...)
}

func (l *simpleLogger) Warnf(format string, args ...interface{}) {
	log.Printf("W! "+format, args...)
}

func (l *simpleLogger) Warn(args ...interface{}) {
	log.Print(append([]interface{}{"W! "}, args...)...)
}

func (l *simpleLogger) Errorf(format string, args ...interface{}) {
	log.Printf("E! "+format, args...)
}

func (l *simpleLogger) Error(args ...interface{}) {
	log.Print(append([]interface{}{"E! "}, args...)...)
}
