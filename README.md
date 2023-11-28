# Power Telemetry Library

`powertelemetry` is a Golang library that provides functionalities to get power management
related metrics for Intel processors.

## Metrics

Metric types can be distinguished based on the host topology attributes associated with them:

- *CPU metric*: Metric value related to a specific logical CPU (CPU ID).
- *Package metric*: Metric value related to a specific package ID (socket ID).
- *Die metric*: Metric value related to a specific die ID.

**The following metrics are supported by Power Telemetry Library:**

| Metric name                           | Type           | Description                                                                                                                                                                                                                                                      | Units           |
|---------------------------------------|----------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------|
| `CurrentPackagePowerConsumptionWatts` | Package     | Current power consumption of processor package.                                                                                                                                                                                                                  | Watts           |
| `CurrentDramPowerConsumptionWatts`    | Package     | Current power consumption of processor package DRAM subsystem.                                                                                                                                                                                                   | Watts           |
| `PackageThermalDesignPowerWatts`      | Package     | Maximum Thermal Design Power (TDP) available for processor package.                                                                                                                                                                                              | Watts           |
| `MaxTurboFreqList`                    | Package     | Maximum reachable turbo frequency for number of cores active.                                                                                                                                                                                                    | MHz             |
| `CurrentUncoreFrequency`              | Package/Die | Current uncore frequency for die in processor package. This value is available from `intel-uncore-frequency` module for kernel >= 5.18. For older kernel versions it needs to be accessed via MSR. In case of lack of loaded `msr`, value will not be collected. | MHz             |
| `InitialUncoreFrequencyMin`           | Package/Die | Initial minimum uncore frequency limit for die in processor package.                                                                                                                                                                                             | MHz             |
| `InitialUncoreFrequencyMax`           | Package/Die | Initial maximum uncore frequency limit for die in processor package.                                                                                                                                                                                             | MHz             |
| `CustomizedUncoreFrequencyMin`        | Package/Die | Customized minimum uncore frequency limit for die in processor package.                                                                                                                                                                                          | MHz             |
| `CustomizedUncoreFrequencyMax`        | Package/Die | Customized maximum uncore frequency limit for die in processor package.                                                                                                                                                                                          | MHz             |
| `CPUBaseFrequency`                    | Package     | CPU Base Frequency (maximum non-turbo frequency) for the processor package.                                                                                                                                                                                      | MHz             |
| `CPUFrequency`                        | CPU         | Current operational frequency of CPU Core.                                                                                                                                                                                                                       | MHz             |
| `CPUC0StateResidency`                 | CPU         | Percentage of time that CPU Core spent in C0 Core residency state.                                                                                                                                                                                               | %               |
| `CPUC1StateResidency`                 | CPU         | Percentage of time that CPU Core spent in C1 Core residency state.                                                                                                                                                                                               | %               |
| `CPUC3StateResidency`                 | CPU         | Percentage of time that CPU Core spent in C3 Core residency state.                                                                                                                                                                                               | %               |
| `CPUC6StateResidency`                 | CPU         | Percentage of time that CPU Core spent in C6 Core residency state.                                                                                                                                                                                               | %               |
| `CPUC7StateResidency`                 | CPU         | Percentage of time that CPU Core spent in C7 Core residency state.                                                                                                                                                                                               | %               |
| `CPUTemperature`                      | CPU         | Current temperature of CPU Core.                                                                                                                                                                                                                                 | degrees Celsius |
| `CPUBusyFrequencyMhz`                 | CPU         | CPU Core Busy Frequency measured as frequency adjusted to CPU Core busy cycles.                                                                                                                                                                                  | MHz             |
| `CPUC0SubstateC01Percent`             | CPU         | Percentage of time that CPU Core spent in C0.1 substate out of the total time in the C0 state.                                                                                                                                                                   | %               |
| `CPUC0SubstateC02Percent`             | CPU         | Percentage of time that CPU Core spent in C0.2 substate out of the total time in the C0 state.                                                                                                                                                                   | %               |
| `CPUC0SubstateC0WaitPercent`          | CPU         | Percentage of time that CPU Core spent in C0_Wait substate out of the total time in the C0 state.                                                                                                                                                                | %               |

> **Note**: Metrics that report processor C-state residencies or power consumption are calculated over elapsed intervals.

## SW Dependencies

### Kernel modules

The library is mostly based on Linux Kernel modules that expose specific metrics over
`sysfs` or `devfs` interfaces. The following dependencies are expected by
the library:

- `intel-rapl` kernel module which exposes Intel Runtime Power Limiting metrics over
  `sysfs` (`/sys/devices/virtual/powercap/intel-rapl`),
- `msr` kernel module that provides access to processor model specific
  registers over `devfs` (`/dev/cpu/cpu%d/msr`),
- `cpufreq` kernel module - which exposes per-CPU Frequency over `sysfs`
  (`/sys/devices/system/cpu/cpu%d/cpufreq/scaling_cur_freq`),
- `intel-uncore-frequency` kernel module which exposes Intel uncore frequency metrics
  over `sysfs` (`/sys/devices/system/cpu/intel_uncore_frequency`).

Make sure that required kernel modules are loaded and running. Modules might have to be manually enabled by using `modprobe`. Depending on the kernel version, run commands:

```sh
# rapl modules:
## kernel < 4.0
sudo modprobe intel_rapl
## kernel >= 4.0
sudo modprobe rapl
sudo modprobe intel_rapl_common
sudo modprobe intel_rapl_msr

# msr module:
sudo modprobe msr

# cpufreq module:
### integrated in kernel 

# intel-uncore-frequency module:
## only for kernel >= 5.6.0
sudo modprobe intel-uncore-frequency
```

### Kernel's perf interface

For perf-related metrics, when the application that uses this library is not running as root,
the following capability should be added to the application executable:

```sh
sudo setcap cap_sys_admin+ep <path_to_application>
```

Alternatively, `/proc/sys/kernel/perf_event_paranoid` has to be set to
value less than 1.

Depending on environment and configuration (number of monitored CPUs
and number of enabled metrics), it might be required to increase
the limit on the number of open file descriptors allowed.
This can be done for example by using `ulimit -n` command.

### Dependencies of metrics on system configuration

Details of these dependencies are discussed above:

| Metric name                           | Type           | Dependency                                     |
|---------------------------------------|----------------|------------------------------------------------|
| `CurrentPackagePowerConsumptionWatts` | Package        | `rapl` kernel module(s)                        |
| `CurrentDramPowerConsumptionWatts`    | Package        | `rapl` kernel module(s)                        |
| `PackageThermalDesignPowerWatts`      | Package        | `rapl` kernel module(s)                        |
| `MaxTurboFreqList`                    | Package        | `msr` kernel module                            |
| `CurrentUncoreFrequency`              | Package/Die    | `intel-uncore-frequency`/`msr` kernel modules* |
| `InitialUncoreFrequencyMin`           | Package/Die    | `intel-uncore-frequency` kernel module         |
| `InitialUncoreFrequencyMax`           | Package/Die    | `intel-uncore-frequency` kernel module         |
| `CustomizedUncoreFrequencyMin`        | Package/Die    | `intel-uncore-frequency` kernel module         |
| `CustomizedUncoreFrequencyMax`        | Package/Die    | `intel-uncore-frequency` kernel module         |
| `CPUBaseFrequency`                    | Package/Die    | `msr` kernel module                            |
| `CPUFrequency`                        | CPU            | `cpufreq` kernel module                        |
| `CPUC0StateResidency`                 | CPU            | `msr` kernel module                            |
| `CPUC1StateResidency`                 | CPU            | `msr` kernel module                            |
| `CPUC3StateResidency`                 | CPU            | `msr` kernel module                            |
| `CPUC6StateResidency`                 | CPU            | `msr` kernel module                            |
| `CPUC7StateResidency`                 | CPU            | `msr` kernel module                            |
| `CPUTemperature`                      | CPU            | `msr` kernel module                            |
| `CPUBusyFrequencyMhz`                 | CPU            | `msr` kernel module                            |
| `CPUC0SubstateC01Percent`             | CPU            | kernel's `perf` interface                      |
| `CPUC0SubstateC02Percent`             | CPU            | kernel's `perf` interface                      |
| `CPUC0SubstateC0WaitPercent`          | CPU            | kernel's `perf` interface                      |

*starting from kernel version 5.18, only the `intel-uncore-frequency` module
is required. For older kernel versions, the metric `CurrentUncoreFrequency`
requires the `msr` module to be enabled.

### Root privileges

**The application that uses this library may require
root privileges to read all the metrics**
(depending on OS type or configuration).

Alternatively, the following capabilities can be added to
the application executable:

```sh
# without perf-related metrics:
sudo setcap cap_sys_rawio,cap_dac_read_search+ep <path_to_application>

# with perf-related metrics:
sudo setcap cap_sys_rawio,cap_dac_read_search,cap_sys_admin+ep <path_to_application>
```

## HW Dependencies

Specific metrics require certain processor features to be present, otherwise
this library won't be able to read them. The user can detect supported
processor features by reading `/proc/cpuinfo` file.
The library assumes crucial properties are the same for all CPU cores in the system.

The following `processor` properties are examined in more detail
in this section:

- `vendor_id`
- `cpu family`
- `model`
- `flags`

The following processor properties are required by the library:

- Processor `vendor_id` must be `GenuineIntel` and `cpu family` must be `6` -
  since data used by the library are Intel-specific.
- The following processor flags shall be present:
  - `msr` shall be present for the library to read platform data from processor
      model specific registers and collect the following metrics:
    - `CPUC0StateResidency`
    - `CPUC1StateResidency`
    - `CPUC3StateResidency`
    - `CPUC6StateResidency`
    - `CPUC7StateResidency`
    - `CPUBusyFrequencyMhz`
    - `CPUTemperature`
    - `CPUBaseFrequency`
    - `MaxTurboFreqList`
    - `CurrentUncoreFrequency` (for kernel < 5.18)
  - `aperfmperf` shall be present to collect the following metrics:
    - `CPUC0StateResidency`
    - `CPUC1StateResidency`
    - `CPUBusyFrequencyMhz`
  - `dts` shall be present to collect:
    - `CPUTemperature`
- Please consult the table below which metrics among those listed are supported by the host's processor `model`:
  - `CPUC1StateResidency`
  - `CPUC3StateResidency`
  - `CPUC6StateResidency`
  - `CPUC7StateResidency`
  - `CPUTemperature`
  - `CPUBaseFrequency`
  - `CurrentUncoreFrequency`
  - `InitialUncoreFrequencyMin`
  - `InitialUncoreFrequencyMax`
  - `CustomizedUncoreFrequencyMin`
  - `CustomizedUncoreFrequencyMax`

      | Model number | Processor name                  | `CPUC1StateResidency`<br/>`CPUC6StateResidency`<br/>`CPUTemperature`<br/>`CPUBaseFrequency` | `CPUC3StateResidency` | `CPUC7StateResidency` | `CurrentUncoreFrequency`<br/>`InitialUncoreFrequencyMin`<br/>`InitialUncoreFrequencyMax`<br/>`CustomizedUncoreFrequencyMin`<br/>`CustomizedUncoreFrequencyMax` |
      |--------------|---------------------------------|:-------------------------------------------------------------------------------------------:|:---------------------:|:---------------------:|:--------------------------------------------------------------------------------------------------------------------------------------------------------------:|
      | 0x1E         | Intel Nehalem                   |                                              ✓                                              |           ✓           |                       |                                                                                                                                                                |
      | 0x1F         | Intel Nehalem-G                 |                                              ✓                                              |           ✓           |                       |                                                                                                                                                                |
      | 0x1A         | Intel Nehalem-EP                |                                              ✓                                              |           ✓           |                       |                                                                                                                                                                |
      | 0x2E         | Intel Nehalem-EX                |                                              ✓                                              |           ✓           |                       |                                                                                                                                                                |
      | 0x25         | Intel Westmere                  |                                              ✓                                              |           ✓           |                       |                                                                                                                                                                |
      | 0x2C         | Intel Westmere-EP               |                                              ✓                                              |           ✓           |                       |                                                                                                                                                                |
      | 0x2F         | Intel Westmere-EX               |                                              ✓                                              |           ✓           |                       |                                                                                                                                                                |
      | 0x2A         | Intel Sandybridge               |                                              ✓                                              |           ✓           |           ✓           |                                                                                                                                                                |
      | 0x2D         | Intel Sandybridge-X             |                                              ✓                                              |           ✓           |           ✓           |                                                                                                                                                                |
      | 0x3A         | Intel Ivybridge                 |                                              ✓                                              |           ✓           |           ✓           |                                                                                                                                                                |
      | 0x3E         | Intel Ivybridge-X               |                                              ✓                                              |           ✓           |           ✓           |                                                                                                                                                                |
      | 0x3C         | Intel Haswell                   |                                              ✓                                              |           ✓           |           ✓           |                                                                                                                                                                |
      | 0x3F         | Intel Haswell-X                 |                                              ✓                                              |           ✓           |           ✓           |                                                                                                                                                                |
      | 0x45         | Intel Haswell-L                 |                                              ✓                                              |           ✓           |           ✓           |                                                                                                                                                                |
      | 0x46         | Intel Haswell-G                 |                                              ✓                                              |           ✓           |           ✓           |                                                                                                                                                                |
      | 0x3D         | Intel Broadwell                 |                                              ✓                                              |           ✓           |           ✓           |                                                                                                                                                                |
      | 0x47         | Intel Broadwell-G               |                                              ✓                                              |           ✓           |           ✓           |                                                                               ✓                                                                                |
      | 0x4F         | Intel Broadwell-X               |                                              ✓                                              |           ✓           |                       |                                                                               ✓                                                                                |
      | 0x56         | Intel Broadwell-D               |                                              ✓                                              |           ✓           |                       |                                                                               ✓                                                                                |
      | 0x4E         | Intel Skylake-L                 |                                              ✓                                              |           ✓           |           ✓           |                                                                                                                                                                |
      | 0x5E         | Intel Skylake                   |                                              ✓                                              |           ✓           |           ✓           |                                                                                                                                                                |
      | 0x55         | Intel Skylake-X                 |                                              ✓                                              |                       |                       |                                                                               ✓                                                                                |
      | 0x8E         | Intel KabyLake-L                |                                              ✓                                              |           ✓           |           ✓           |                                                                                                                                                                |
      | 0x9E         | Intel KabyLake                  |                                              ✓                                              |           ✓           |           ✓           |                                                                                                                                                                |
      | 0xA5         | Intel CometLake                 |                                              ✓                                              |           ✓           |           ✓           |                                                                                                                                                                |
      | 0xA6         | Intel CometLake-L               |                                              ✓                                              |           ✓           |           ✓           |                                                                                                                                                                |
      | 0x66         | Intel CannonLake-L              |                                              ✓                                              |                       |           ✓           |                                                                                                                                                                |
      | 0x6A         | Intel IceLake-X                 |                                              ✓                                              |                       |                       |                                                                               ✓                                                                                |
      | 0x6C         | Intel IceLake-D                 |                                              ✓                                              |                       |                       |                                                                               ✓                                                                                |
      | 0x7D         | Intel IceLake                   |                                              ✓                                              |                       |                       |                                                                                                                                                                |
      | 0x7E         | Intel IceLake-L                 |                                              ✓                                              |                       |           ✓           |                                                                                                                                                                |
      | 0x9D         | Intel IceLake-NNPI              |                                              ✓                                              |                       |           ✓           |                                                                                                                                                                |
      | 0xA7         | Intel RocketLake                |                                              ✓                                              |                       |           ✓           |                                                                                                                                                                |
      | 0x8C         | Intel TigerLake-L               |                                              ✓                                              |                       |           ✓           |                                                                                                                                                                |
      | 0x8D         | Intel TigerLake                 |                                              ✓                                              |                       |           ✓           |                                                                                                                                                                |
      | 0x8F         | Intel Sapphire Rapids X         |                                              ✓                                              |                       |                       |                                                                               ✓                                                                                |
      | 0xCF         | Intel Emerald Rapids X          |                                              ✓                                              |                       |                       |                                                                               ✓                                                                                |
      | 0xAD         | Intel Granite Rapids X          |                                              ✓                                              |                       |                       |                                                                                                                                                                |
      | 0x8A         | Intel Lakefield                 |                                              ✓                                              |                       |           ✓           |                                                                                                                                                                |
      | 0x97         | Intel AlderLake                 |                                              ✓                                              |                       |           ✓           |                                                                               ✓                                                                                |
      | 0x9A         | Intel AlderLake-L               |                                              ✓                                              |                       |           ✓           |                                                                               ✓                                                                                |
      | 0xB7         | Intel RaptorLake                |                                              ✓                                              |                       |           ✓           |                                                                               ✓                                                                                |
      | 0xBA         | Intel RaptorLake-P              |                                              ✓                                              |                       |           ✓           |                                                                               ✓                                                                                |
      | 0xBF         | Intel RaptorLake-S              |                                              ✓                                              |                       |           ✓           |                                                                               ✓                                                                                |
      | 0xAC         | Intel MeteorLake                |                                              ✓                                              |                       |           ✓           |                                                                               ✓                                                                                |
      | 0xAA         | Intel MeteorLake-L              |                                              ✓                                              |                       |           ✓           |                                                                               ✓                                                                                |
      | 0xC6         | Intel ArrowLake                 |                                              ✓                                              |                       |           ✓           |                                                                                                                                                                |
      | 0xBD         | Intel LunarLake                 |                                              ✓                                              |                       |           ✓           |                                                                                                                                                                |
      | 0x37         | Intel Atom® Bay Trail           |                                              ✓                                              |                       |                       |                                                                                                                                                                |
      | 0x4D         | Intel Atom® Avaton              |                                              ✓                                              |                       |                       |                                                                                                                                                                |
      | 0x4A         | Intel Atom® Merrifield          |                                              ✓                                              |                       |                       |                                                                                                                                                                |
      | 0x5A         | Intel Atom® Moorefield          |                                              ✓                                              |                       |                       |                                                                                                                                                                |
      | 0x4C         | Intel Atom® Airmont             |                                              ✓                                              |           ✓           |                       |                                                                                                                                                                |
      | 0x5C         | Intel Atom® Apollo Lake         |                                              ✓                                              |           ✓           |           ✓           |                                                                                                                                                                |
      | 0x5F         | Intel Atom® Denverton           |                                              ✓                                              |                       |                       |                                                                                                                                                                |
      | 0x7A         | Intel Atom® Goldmont            |                                              ✓                                              |           ✓           |           ✓           |                                                                                                                                                                |
      | 0x86         | Intel Atom® Jacobsville         |                                              ✓                                              |                       |                       |                                                                                                                                                                |
      | 0x96         | Intel Atom® Elkhart Lake        |                                              ✓                                              |                       |           ✓           |                                                                                                                                                                |
      | 0x9C         | Intel Atom® Jasper Lake         |                                              ✓                                              |                       |           ✓           |                                                                                                                                                                |
      | 0xBE         | Intel AlderLake-N               |                                              ✓                                              |                       |           ✓           |                                                                                                                                                                |
      | 0xAF         | Intel Sierra Forest             |                                              ✓                                              |                       |                       |                                                                                                                                                                |
      | 0xB6         | Intel Grand Ridge               |                                              ✓                                              |                       |                       |                                                                                                                                                                |
      | 0x57         | Intel Xeon® PHI Knights Landing |                                              ✓                                              |                       |                       |                                                                                                                                                                |
      | 0x85         | Intel Xeon® PHI Knights Mill    |                                              ✓                                              |                       |                       |                                                                                                                                                                |

## How to use

### Initialization

`powertelemetry` library implements optional builder pattern. It allows the user to specify configuration parameters, and the dependencies to initalize needed to provide metrics information via `WithX` exported functions.

```go
ptel, err := New(opts...)
```

Supported options are:

- `WithCoreFrequency`: Option that enables access to metrics which rely on `cpufreq` kernel module.
- `WithMsr`: Option that enables access to metrics which rely on `msr` kernel module.
- `WithMsrTimeout`: Same as `WithMsr`, but it accepts an additional argument to specify the timeout for MSR reads.
- `WithIncludedCPUs/WithExcludedCPUs`: Option that allows to specify which logical CPU ID have access to `msr` and `cpufreq` kernel modules and `perf_events` kernel interface, by inclusion or exclusion. Notice that only one of these options can be used during instantiation. When omitted, all logical CPUs from host topology are accessible.
- `WithRapl`: Option that enables access to metrics which rely on `rapl` kernel module.
- `WithUncoreFrequency`: Option that enables access to metrics which rely on `intel-uncore-frequency` kernel module.
- `WithPerf`: Option that enables access to metrics which rely on `perf_events` kernel interface. It takes the path of a JSON file with perf event definitions specific for the host's CPU model. Files can be found in [`perfmon`](https://github.com/intel/perfmon) repository.
- `WithLogger`: The user can provide a custom logger.

Refer to [Dependencies of metrics on system configuration](#dependencies-of-metrics-on-system-configuration) section to check which options need to be enabled for each metric.

#### Example: Initialize `PowerTelemetry` using included CPUs option

This example uses options to enable all supported metrics. Additionally, uses `WithIncludedCPUs` to limit the CPU IDs that can be used to retrieve metrics which rely on `msr` and `cpufreq` kernel modules.

```go
pt, err := powertelemetry.New(
    powertelemetry.WithIncludedCPUs([]int{0, 1, 2, 3}),
    powertelemetry.WithMsr(),
    powertelemetry.WithCoreFrequency(),
    powertelemetry.WithRapl(),
    powertelemetry.WithUncoreFrequency(),
    powertelemetry.WithPerf("/path/to/events.json"),
)
```

#### Example: Initialize `PowerTelemetry` using excluded CPUs option

This example uses options to enable metrics which rely on `msr` and `cpufreq` kernel modules. Additionally, uses `WithExcludedCPUs` to exclude CPU IDs that can be used to retrieve these metrics.

```go
pt, err := powertelemetry.New(
    powertelemetry.WithExcludedCPUs([]int{0, 1, 2, 3}),
    powertelemetry.WithMsr(),
    powertelemetry.WithCoreFrequency(),
)
```

### Get Metric Values

`powertelemetry` provides exported methods to get metric values defined in the [Metrics](#metrics) section. These methods accept an argument depending on the metric type.

The exported method to get metric values have the following naming convention:

```go
// CPU metric type.
func (ptel *PowerTelemetry) Get<metric_name>(cpuID int) (<value_type>, error)

// Package metric type.
func (ptel *PowerTelemetry) Get<metric_name>(packageID int) (<value_type>, error)

// Package/die metric type.
func (ptel *PowerTelemetry) Get<metric_name>(packageID, dieID int) (<value_type>, error)
```

Where:

- `metric_name` corresponds to the metric name of the supported [Metrics](#metrics) section.
- `value_type` can be either `float64` or `uint64`, depending on the metric.

There are several types of metrics depending on how values are calculated:

- Instantaneous: The metric value corresponds to the time the specific instant in which the method is called.
- Elapsed interval: The metric value corresponds to a specific time interval.

### Instantaneous Metrics

Following are the metrics that provide instantaneous values:

- `CPUTemperature`
- `CPUFrequency`
- `CPUBaseFrequency`
- `CurrentUncoreFrequency`
- `InitialUncoreFrequencyMin`
- `InitialUncoreFrequencyMax`
- `CustomizedUncoreFrequencyMin`
- `CustomizedUncoreFrequencyMax`
- `MaxTurboFreqList`
- `PackageThermalDesignPowerWatts`

#### Example: Get the instantaneous value of CPU temperature metric

```go
// CPU temperature metric
cpuTemp, err := pt.GetCPUTemperature(cpuID)
if err != nil {
  // handle error
}
```

### Elapsed Interval Metrics

The following metrics depend on elapsed intervals:

- Metrics that rely on `msr`:
  - `CPUC0StateResidency`
  - `CPUC1StateResidency`
  - `CPUC3StateResidency`
  - `CPUC6StateResidency`
  - `CPUC7StateResidency`
  - `CPUBusyFrequencyMhz`
- Metrics that rely on `perf`:
  - `CPUC0SubstateC01Percent`
  - `CPUC0SubstateC02Percent`
  - `CPUC0SubstateC0WaitPercent`
- Metrics that rely on `rapl`:
  - `CurrentPackagePowerConsumptionWatts`
  - `CurrentDramPowerConsumptionWatts`

The elapsed time interval is automatically calculated between subsequent calls to retrieve metric values. It is recommended to use a scheduler to consistently retrieve metrics over a fixed time interval.

### Metrics relying on `rapl`

The following example shows how to retrieve the value of power related metrics based on `rapl` kernel module.

#### Example: Get metrics which rely on `rapl` kernel module

```go
// First read of metrics at init (t0).
ptel, err := ptel.New(WithRapl())
if err != nil {
  // handle error
}

// Method call at t1. Metric value corresponds to time interval t1-t0.
powerInterval1, err := pt.GetCurrentPackagePowerConsumptionWatts(packageID)
if err != nil {
  // handle error
}

// Method call at t2. Metric value corresponds to time interval t2-t1.
powerInterval2, err := pt.GetCurrentPackagePowerConsumptionWatts(packageID)
if err != nil {
  // handle error
}
```

> **Note**: The first metric reading operation happens at initialization, when `WithRapl`option is present.

### Metrics relying on `msr`

C-state residency metrics need an additional method call to `UpdatePerCPUMetrics` that reads all required offsets of the corresponding MSR registers prior to providing their values.

#### Example: Get time-elapsed metrics which rely on `msr` kernel module

```go
// First offset reading of MSR registers (call to UpdatePerCPUMetrics) at init (t0).
ptel, err := ptel.New(WithMsr())
if err != nil {
  // handle error
}

cpuID := 0

// Method call at t1. Elapsed time is calculated from current and previous call to UpdatePerCPUMetrics, t1-t0.
if err := ptel.UpdatePerCPUMetrics(cpuID); err != nil {
  // handle error
  return
}

// Get CPUC0StateResidency corresponding to previous elapsed interval.
c0State, err := ptel.GetCPUC0StateResidency(cpuID)
if err != nil {
  // handle error
}

// Get CPUC1StateResidency corresponding to previous elapsed interval.
c1State, err := ptel.GetCPUC1StateResidency(cpuID)
if err != nil {
  // handle error
}
```

> **Note**: The first reading operations of the MSR register happen at initialization, when `WithMsr` or `WithMsrTimeout` options are present.

### Metrics relying on `perf`

C0-substate metrics need an additional `ReadPerfEvents` method call that reads all required perf events, per-CPU, prior to providing their values.

When an instance of `PowerTelemetry` has been successfully initialized with the option `WithPerf`, perf events are activated. This means multiple file descriptors remain open. Therefore, if the user no longer needs to get `perf` specific metrics these resources need to be released, via `DeactivatePerfEvents` method call.

> **Note**: Event activation happens at initialization, when `WithPerf` option is added.

#### Example: Get time-elapsed metrics which rely on `perf` kernel interface

```go
// Read events related to perf-related metrics.
// Elapsed time is calculated from current and previous call to ReadPerfEvents.
if err := ptel.ReadPerfEvents(); err != nil {
  // handle error
  return
}

// Get GetCPUC0SubstateC01Percent corresponding to previous elapsed interval.
c0SubstateC01, err := ptel.GetCPUC0SubstateC01Percent(cpuID)
if err != nil {
  // handle error
}

// Get GetCPUC0SubstateC02Percent corresponding to previous elapsed interval.
c0SubstateC02, err := ptel.GetCPUC0SubstateC02Percent(cpuID)
if err != nil {
  // handle error
}

// Get GetCPUC0SubstateC0WaitPercent corresponding to previous elapsed interval.
c0SubstateC0Wait, err := ptel.GetCPUC0SubstateC0WaitPercent(cpuID)
if err != nil {
  // handle error
}

// Release resources. Close file descriptors.
err := ptel.DeactivatePerfEvents()
if err != nil {
  // handle error
}
```

### Error Handling

This library exposes several types of errors, providing the user the flexibility to handle them differently:

- `MultiError`: Holds a slice of error descriptions. It is used to mark errors that happened during the initialization of `PowerTelemetry` dependencies.
- `ModuleNotInitializedError`: Used to indicate that a dependency has not been initialized, and the user tried to access it.
- `MetricNotSupportedError`: Used to indicate that a metric is not supported by the host's CPU model.

#### Usage of `MultiError`

When creating a new `PowerTelemetry` instance, the `New` function returns a `MultiError` if any of the dependencies, requested via options, failed to initialize.

```go
ptel, err := powertelemetry.New(
  powertelemetry.WithMsr(),
  powertelemetry.WithRapl(),
  powertelemetry.WithCoreFrequency(),
  powertelemetry.WithUncoreFrequency(),
  powertelemetry.WithPerf("/path/to/events.json"),
)

var initErr *powertelemetry.MultiError
if err != nil {
  if !errors.As(err, &initErr) {
    logger.Errorf("Failed to build powertelemetry instance: %v", err)
    os.Exit(1)
  }
  logger.Warn(err)
}
```

Typical scenarios that return a `MultiError` type would be:

- Requesting a dependency, but the corresponding kernel module was not loaded previously.
- Provide invalid paths to `WithX` options that allow to specify custom path used to initialize the dependencies.
- Provide invalid JSON file for perf event definitions via `WithPerf`.

#### Usage of `ModuleNotInitializedError`

Calls to a [get metric method](#get-metric-values) return an error of type `ModuleNotInitializedError` if the dependency to which it relies on, has not been initialized via the corresponding `WithX` option, or it failed to initialize.

This might be used to:

- Prevent subsequent calls to metrics relying on the same dependency.
- Prevent subsequent calls to the same metric getter, when looping through multiple logical CPU IDs or package/die IDs.

##### Example: `ModuleNotInitializedError` error handling for CPU frequency metric

```go
// CPU current frequency metric
cpuID := 0
cpuFreq, err := ptel.GetCPUFrequency(cpuID)

var moduleErr *powertelemetry.ModuleNotInitializedError
if err != nil {
  if !errors.As(err, &moduleErr) {
    // Handle module not initialized error
  }
  // Handle other error types
}
```

#### Usage of `MetricNotSupportedError`

As mentioned in [HW Dependencies](#hw-dependencies) section, specific metrics require certain processor features to be present, or specific processor models.

`powertelemetry` library provides exported functions that allow the user to check if a metric is supported by the CPU. If not supported, a `MetricNotSupportedError` is returned, for the user to handle it.

The exported functions have the following naming convention:

```go
func CheckIf<metric_name>Supported(cpuModel int) error
```

This might be used to disable metric requests based on CPU model compatibility.

##### Example: `MetricNotSupportedError` error handling for CPU C1 state residency metric

```go
var notSupportedErr *powertelemetry.MetricNotSupportedError

err := ptel.CheckIfCPUC1StateResidencySupported(cpuModel)
if err != nil && errors.As(err, &notSupportedErr)
  // handle not supported metric error
}
```
