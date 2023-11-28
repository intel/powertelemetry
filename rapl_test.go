// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//go:build linux && amd64

package powertelemetry

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestDomainTypeToString(t *testing.T) {
	t.Run("Package", func(t *testing.T) {
		packageType := domainType(0)
		require.Equal(t, "package", packageType.String())
	})

	t.Run("Dram", func(t *testing.T) {
		dramType := domainType(1)
		require.Equal(t, "dram", dramType.String())
	})

	t.Run("Invalid", func(t *testing.T) {
		invalidType := domainType(2)
		require.Equal(t, "", invalidType.String())
	})
}

func TestAttrTypeToString(t *testing.T) {
	t.Run("CurrentEnergy", func(t *testing.T) {
		currEnergyType := attrType(0)
		require.Equal(t, "currEnergy", currEnergyType.String())
	})

	t.Run("MaximumEnergy", func(t *testing.T) {
		maxEnergyType := attrType(1)
		require.Equal(t, "maxEnergy", maxEnergyType.String())
	})

	t.Run("MaximumPower", func(t *testing.T) {
		maxPowerType := attrType(2)
		require.Equal(t, "maxPower", maxPowerType.String())
	})

	t.Run("Invalid", func(t *testing.T) {
		invalidType := attrType(3)
		require.Equal(t, "", invalidType.String())
	})
}

func TestZoneGetters(t *testing.T) {
	zoneName := "package-0"
	zonePath := "testdata/intel-rapl/intel-rapl:0"

	dramZone := &zone{
		name:     "dram",
		path:     "testdata/intel-rapl/intel-rapl:0/intel-rapl:0:0",
		subzones: make([]powerZone, 0),
	}
	subZones := []powerZone{dramZone}

	sample := attrSample{
		value:     100000,
		timestamp: time.Now(),
	}

	z := &zone{
		name:     zoneName,
		path:     zonePath,
		energy:   sample,
		subzones: subZones,
	}

	require.Equal(t, zoneName, z.getName())
	require.Equal(t, zonePath, z.getPath())
	require.Equal(t, sample, z.getEnergySample())
	require.Equal(t, subZones, z.getSubzones())
	require.Equal(t, dramZone, z.getDomainSubzone(dramDomain.String()))
}

func TestZoneSetters(t *testing.T) {
	z := &zone{
		name:     "package-0",
		path:     "testdata/intel-rapl/intel-rapl:0",
		energy:   attrSample{},
		subzones: make([]powerZone, 0),
	}

	// set energy sample to package domain zone
	s := attrSample{
		value:     100000,
		timestamp: time.Now(),
	}
	z.setEnergySample(s)

	// set dram subzone as child of package zone
	dramZone := &zone{
		name:     "dram",
		path:     "testdata/intel-rapl/intel-rapl:0/intel-rapl:0:0",
		subzones: make([]powerZone, 0),
	}
	z.addSubzone(dramZone)

	require.Equal(t, s, z.getEnergySample())
	require.Equal(t, dramZone, z.getDomainSubzone(dramDomain.String()))
}

func (s *raplTimeSensitiveTestSuite) TestZoneReadAttribute() {
	testCases := []struct {
		name      string
		path      string
		attr      string
		sampleExp attrSample
		err       error
	}{
		{
			name:      "Unsupported",
			path:      makeTestDataPath("testdata/intel-rapl/intel-rapl:0"),
			attr:      "minEnergy",
			sampleExp: attrSample{},
			err:       errors.New("unsupported attribute \"minEnergy\""),
		},
		{
			name:      "MissingFile",
			path:      makeTestDataPath("testdata/intel-rapl/intel-rapl:1/intel-rapl:1:0"),
			attr:      currEnergyAttr.String(),
			sampleExp: attrSample{},
			err:       errors.New(`error reading file "` + makeTestDataPath("testdata/intel-rapl/intel-rapl:1/intel-rapl:1:0/energy_uj") + `"`),
		},
		{
			name:      "FileContentNonNumeric",
			path:      makeTestDataPath("testdata/intel-rapl/intel-rapl:1"),
			attr:      maxEnergyAttr.String(),
			sampleExp: attrSample{},
			err:       errors.New("error converting attribute file content to float64"),
		},
		{
			name: "CurrEnergy",
			path: makeTestDataPath("testdata/intel-rapl/intel-rapl:0"),
			attr: currEnergyAttr.String(),
			sampleExp: attrSample{
				value:     206999074695,
				timestamp: fakeClock.Now(),
			},
			err: nil,
		},
		{
			name: "MaxEnergy",
			path: makeTestDataPath("testdata/intel-rapl/intel-rapl:0/intel-rapl:0:1"),
			attr: maxEnergyAttr.String(),
			sampleExp: attrSample{
				value:     65712999613,
				timestamp: fakeClock.Now(),
			},
			err: nil,
		},
		{
			name: "MaxPowerConstraint",
			path: makeTestDataPath("testdata/intel-rapl/intel-rapl:0"),
			attr: maxPowerConstraintAttr.String(),
			sampleExp: attrSample{
				value:     250000000,
				timestamp: fakeClock.Now(),
			},
			err: nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			z := &zone{
				name: tc.name,
				path: tc.path,
			}

			sampleOut, err := z.readAttribute(tc.attr)
			s.Require().Equal(tc.sampleExp, sampleOut)
			if tc.err != nil {
				s.Require().ErrorContains(err, tc.err.Error())
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func TestPackageZoneGetPackageID(t *testing.T) {
	testCases := []struct {
		name         string
		zoneName     string
		zonePath     string
		packageIDExp int
		err          error
	}{
		{
			name:         "InvalidName",
			zoneName:     "package-socket",
			zonePath:     "intel-rapl:0",
			packageIDExp: 0,
			err:          errors.New("invalid package domain name for zone at path \"intel-rapl:0\""),
		},
		{
			name:         "InvalidPath",
			zoneName:     "package-0",
			zonePath:     "rapl:0",
			packageIDExp: 0,
			err:          errors.New("invalid package domain zone path \"rapl:0\""),
		},
		{
			name:         "PackageIDMismatch",
			zoneName:     "package-1",
			zonePath:     "intel-rapl:0",
			packageIDExp: 0,
			err:          errors.New("package ID mismatch between zone path \"intel-rapl:0\" and zone name \"package-1\""),
		},
		{
			name:         "PackageNameWithLeadingZeroes",
			zoneName:     "package-01",
			zonePath:     "intel-rapl:1",
			packageIDExp: 0,
			err:          errors.New("invalid package domain name for zone at path \"intel-rapl:1\""),
		},
		{
			name:         "PackageZonePathWithLeadingZeroes",
			zoneName:     "package-1",
			zonePath:     "intel-rapl:01",
			packageIDExp: 0,
			err:          errors.New("package ID mismatch between zone path \"intel-rapl:01\" and zone name \"package-1\""),
		},
		{
			name:         "PackageID_1",
			zoneName:     "package-1",
			zonePath:     "intel-rapl:1",
			packageIDExp: 1,
			err:          nil,
		},
		{
			name:         "PackageID_10",
			zoneName:     "package-10",
			zonePath:     "intel-rapl:10",
			packageIDExp: 10,
			err:          nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			z := &packageZone{
				&zone{
					name: tc.zoneName,
					path: tc.zonePath,
				},
			}

			packageIDOut, err := z.getPackageID()
			require.Equal(t, tc.packageIDExp, packageIDOut)
			if err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIsRaplLoaded(t *testing.T) {
	testCases := []struct {
		name     string
		filePath string
		expected bool
		err      error
	}{
		{
			name:     "EmptyFilename",
			filePath: "",
			expected: false,
			err:      errors.New("file path is empty"),
		},
		{
			name:     "FileNotExist",
			filePath: "dummy_proc_modules_file",
			expected: false,
			err:      errors.New("file \"dummy_proc_modules_file\" does not exist"),
		},
		{
			name:     "PathIsADir",
			filePath: "testdata",
			expected: false,
			err:      errors.New("could not read file \"testdata\": read testdata: is a directory"),
		},
		{
			name:     "Symlink",
			filePath: "testdata/symlink",
			expected: false,
			err:      errors.New("file \"testdata/symlink\" is a symlink"),
		},
		{
			name:     "NotLoaded",
			filePath: "testdata/proc_modules_rapl_not_loaded",
			expected: false,
			err:      nil,
		},
		{
			name:     "Loaded",
			filePath: "testdata/proc_modules_rapl_loaded",
			expected: true,
			err:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := &raplData{}
			out, err := r.isRaplLoaded(tc.filePath)
			require.Equal(t, tc.expected, out)
			if tc.err != nil {
				require.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type raplTimeSensitiveTestSuite struct {
	suite.Suite
}

func (s *raplTimeSensitiveTestSuite) SetupTest() {
	setFakeClock()
	fakeClock.Set(time.Now())
}

func (s *raplTimeSensitiveTestSuite) TearDownTest() {
	unsetFakeClock()
}

func TestRaplTimeSensitive(t *testing.T) {
	suite.Run(t, new(raplTimeSensitiveTestSuite))
}

func (s *raplTimeSensitiveTestSuite) TestInitZoneMap() {
	testCases := []struct {
		name      string
		raplPath  string
		raplZones map[int]powerZone
		err       error
	}{
		{
			name:      "RaplPathEmpty",
			raplPath:  "",
			raplZones: nil,
			err:       errors.New("base path of rapl control zone cannot be empty"),
		},
		{
			name:      "RaplPathNotExist",
			raplPath:  "/dummy/path",
			raplZones: nil,
			err:       errors.New("file \"/dummy/path\" does not exist"),
		},
		{
			name:      "RaplPathInvalid",
			raplPath:  makeTestDataPath("testdata/"),
			raplZones: nil,
			err:       errors.New(`no package zones found for base path "` + makeTestDataPath("testdata/") + `"`),
		},
		{
			name:      "RaplPathIsNotADir",
			raplPath:  makeTestDataPath("testdata/intel-rapl/intel-rapl:0/name"),
			raplZones: nil,
			err:       errors.New(`error reading path "` + makeTestDataPath("testdata/intel-rapl/intel-rapl:0/name") + `"`),
		},
		{
			name:      "PackageDomainNameFileNotExist",
			raplPath:  makeTestDataPath("testdata/intel-rapl-package-domain-name-not-exist"),
			raplZones: nil,
			err: errors.New(`file "` +
				makeTestDataPath("testdata/intel-rapl-package-domain-name-not-exist/intel-rapl:0/name") + `" does not exist`),
		},
		{
			name:      "PackageDomainNameFileEmpty",
			raplPath:  makeTestDataPath("testdata/intel-rapl-domain-name-empty"),
			raplZones: nil,
			err:       errors.New("zone domain cannot be empty"),
		},
		{
			name:      "DramDomainNameFileNotExist",
			raplPath:  makeTestDataPath("testdata/intel-rapl-dram-domain-name-not-exist"),
			raplZones: nil,
			err: errors.New(`file "` +
				makeTestDataPath("testdata/intel-rapl-dram-domain-name-not-exist/intel-rapl:0/intel-rapl:0:0/name") + `" does not exist`),
		},
		{
			name:      "MismatchPackageDomainID",
			raplPath:  makeTestDataPath("testdata/intel-rapl-invalid-package-domain-name-id"),
			raplZones: nil,
			err: errors.New(`package ID mismatch between zone path "` +
				makeTestDataPath("testdata/intel-rapl-invalid-package-domain-name-id/intel-rapl:1") + `" and zone name "package-0"`),
		},
		{
			name:      "PackageCurrentEnergyAttributeFileNotExist",
			raplPath:  makeTestDataPath("testdata/intel-rapl-package-curr-energy-attr-file-not-exist"),
			raplZones: nil,
			err: errors.New(`error reading file "` +
				makeTestDataPath("testdata/intel-rapl-package-curr-energy-attr-file-not-exist/intel-rapl:0/energy_uj") + `"`),
		},
		{
			name:      "DramCurrentEnergyAttributeFileNotExist",
			raplPath:  makeTestDataPath("testdata/intel-rapl-dram-curr-energy-attr-file-not-exist"),
			raplZones: nil,
			err: errors.New(`error reading file "` +
				makeTestDataPath("testdata/intel-rapl-dram-curr-energy-attr-file-not-exist/intel-rapl:0/intel-rapl:0:1/energy_uj") + `"`),
		},
		{
			name:     "RaplPathValid",
			raplPath: makeTestDataPath("testdata/intel-rapl"),
			raplZones: map[int]powerZone{
				0: &zone{
					name: "package-0",
					path: makeTestDataPath("testdata/intel-rapl/intel-rapl:0"),
					energy: attrSample{
						value:     206999074695,
						timestamp: fakeClock.Now(),
					},
					subzones: []powerZone{
						&zone{
							name:     "domain",
							path:     makeTestDataPath("testdata/intel-rapl/intel-rapl:0/intel-rapl:0:0"),
							subzones: make([]powerZone, 0),
						},
						&zone{
							name: "dram",
							path: makeTestDataPath("testdata/intel-rapl/intel-rapl:0/intel-rapl:0:1"),
							energy: attrSample{
								value:     64155753419,
								timestamp: fakeClock.Now(),
							},
							subzones: make([]powerZone, 0),
						},
					},
				},
				1: &zone{
					name: "package-1",
					path: makeTestDataPath("testdata/intel-rapl/intel-rapl:1"),
					energy: attrSample{
						value:     206999075695,
						timestamp: fakeClock.Now(),
					},
					subzones: []powerZone{
						&zone{
							name:     "socket",
							path:     makeTestDataPath("testdata/intel-rapl/intel-rapl:1/intel-rapl:1:0"),
							subzones: make([]powerZone, 0),
						},
					},
				},
				2: &zone{
					name: "package-2",
					path: makeTestDataPath("testdata/intel-rapl/intel-rapl:2"),
					energy: attrSample{
						value:     205999075695,
						timestamp: fakeClock.Now(),
					},
					subzones: []powerZone{
						&zone{
							name: "dram",
							path: makeTestDataPath("testdata/intel-rapl/intel-rapl:2/intel-rapl:2:0"),
							energy: attrSample{
								value:     66155553419,
								timestamp: fakeClock.Now(),
							},
							subzones: make([]powerZone, 0),
						},
					},
				},
				3: &zone{
					name: "package-3",
					path: makeTestDataPath("testdata/intel-rapl/intel-rapl:3"),
					energy: attrSample{
						value:     205888075695,
						timestamp: fakeClock.Now(),
					},
					subzones: make([]powerZone, 0),
				},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			rapl := &raplData{
				basePath: tc.raplPath,
			}

			err := rapl.initZoneMap()
			s.Require().Equal(tc.raplZones, rapl.zones)
			if tc.err != nil {
				s.Require().ErrorContains(err, tc.err.Error())
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *raplTimeSensitiveTestSuite) TestGetEnergyAttributeWithTimestamp() {
	testCase := []struct {
		name       string
		packageID  int
		domain     string
		energyAttr string
		sample     attrSample
		err        error
	}{
		{
			name:       "InvalidPackageID",
			packageID:  3,
			domain:     "package",
			energyAttr: "currEnergy",
			err:        errors.New("could not find zone for package ID: 3"),
		},
		{
			name:       "InvalidDomainType",
			packageID:  0,
			domain:     "invalid",
			energyAttr: "currEnergy",
			err:        errors.New("unsupported rapl domain \"invalid\""),
		},
		{
			name:       "DramDomainNotExist",
			packageID:  1,
			domain:     "dram",
			energyAttr: "currEnergy",
			err:        errors.New("could not find dram subzone for package ID: 1"),
		},
		{
			name:       "InvalidEnergyAttribute",
			packageID:  0,
			domain:     "package",
			energyAttr: "invalid",
			err:        errors.New("unsupported attribute \"invalid\""),
		},
		{
			name:       "EnergyAttributeFileNotExist",
			packageID:  2,
			domain:     "package",
			energyAttr: "maxEnergy",
			err:        errors.New(`file "` + makeTestDataPath("testdata/intel-rapl/intel-rapl:2/max_energy_range_uj") + `" does not exist`),
		},
		{
			name:       "EnergyAttributeFileNonNumeric",
			packageID:  1,
			domain:     "package",
			energyAttr: "maxEnergy",
			err:        errors.New("error reading energy attribute \"maxEnergy\": error converting attribute file content to float64"),
		},
		{
			name:       "EnergyAttributeFileEmpty",
			packageID:  2,
			domain:     "dram",
			energyAttr: "maxEnergy",
			err:        errors.New("error reading energy attribute \"maxEnergy\": error converting attribute file content to float64"),
		},
		{
			name:       "PackageMaxEnergyAttribute",
			packageID:  0,
			domain:     "package",
			energyAttr: "maxEnergy",
			sample: attrSample{
				value:     262143328850.0,
				timestamp: fakeClock.Now(),
			},
			err: nil,
		},
		{
			name:       "DramCurrEnergyAttribute",
			packageID:  0,
			domain:     "dram",
			energyAttr: "currEnergy",
			sample: attrSample{
				value:     64155753419.0,
				timestamp: fakeClock.Now(),
			},
			err: nil,
		},
	}

	r := &raplData{
		zones: map[int]powerZone{
			0: &zone{
				name: "package-0",
				path: makeTestDataPath("testdata/intel-rapl/intel-rapl:0"),
				subzones: []powerZone{
					&zone{
						name:     "dram",
						path:     makeTestDataPath("testdata/intel-rapl/intel-rapl:0/intel-rapl:0:1"),
						subzones: make([]powerZone, 0),
					},
				},
			},
			1: &zone{
				name:     "package-1",
				path:     makeTestDataPath("testdata/intel-rapl/intel-rapl:1"),
				subzones: []powerZone{},
			},
			2: &zone{
				name: "package-2",
				path: makeTestDataPath("testdata/intel-rapl/intel-rapl:2"),
				subzones: []powerZone{
					&zone{
						name:     "dram",
						path:     makeTestDataPath("testdata/intel-rapl/intel-rapl:2/intel-rapl:2:0"),
						subzones: make([]powerZone, 0),
					},
				},
			},
		},
	}

	for _, tc := range testCase {
		s.Run(tc.name, func() {
			outSample, err := r.getEnergyAttributeWithTimestamp(tc.packageID, tc.domain, tc.energyAttr)
			if tc.err != nil {
				s.Require().ErrorContains(err, tc.err.Error())
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.sample, outSample)
			}
		})
	}
}

func (s *raplTimeSensitiveTestSuite) TestGetLastMeasuredEnergyAttribute() {
	testCases := []struct {
		name      string
		packageID int
		domain    string
		sample    attrSample
		err       error
	}{
		{
			name:      "InvalidPackageID",
			packageID: 3,
			domain:    "package",
			err:       errors.New("could not find zone for package ID: 3"),
		},
		{
			name:      "InvalidDomainType",
			packageID: 0,
			domain:    "socket",
			err:       errors.New("unsupported rapl domain \"socket\""),
		},
		{
			name:      "DramSubzoneNotExist",
			packageID: 1,
			domain:    "dram",
			err:       errors.New("could not find dram subzone for package ID: 1"),
		},
		{
			name:      "PackageCurrEnergy",
			packageID: 0,
			domain:    "package",
			sample: attrSample{
				value:     206999074695.0,
				timestamp: fakeClock.Now(),
			},
			err: nil,
		},
		{
			name:      "DramCurrEnergy",
			packageID: 0,
			domain:    "dram",
			sample: attrSample{
				value:     64155753419.0,
				timestamp: fakeClock.Now(),
			},
			err: nil,
		},
	}

	r := &raplData{
		zones: map[int]powerZone{
			0: &zone{
				name: "package-0",
				path: "testdata/intel-rapl/intel-rapl:0",
				energy: attrSample{
					value:     206999074695.0,
					timestamp: fakeClock.Now(),
				},
				subzones: []powerZone{
					&zone{
						name: "dram",
						path: "testdata/intel-rapl/intel-rapl:0/intel-rapl:0:1",
						energy: attrSample{
							value:     64155753419.0,
							timestamp: fakeClock.Now(),
						},
						subzones: make([]powerZone, 0),
					},
				},
			},
			1: &zone{
				name:     "package-1",
				path:     "testdata/intel-rapl/intel-rapl:1",
				subzones: []powerZone{},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			outSample, err := r.getLastMeasuredEnergyAttribute(tc.packageID, tc.domain)
			s.Require().Equal(tc.sample, outSample)
			if tc.err != nil {
				s.Require().ErrorContains(err, tc.err.Error())
			}
		})
	}
}

func (s *raplTimeSensitiveTestSuite) TestSetLastMeasuredEnergyAttribute() {
	testCases := []struct {
		name      string
		packageID int
		domain    string
		sample    attrSample
		err       error
	}{
		{
			name:      "InvalidPackageID",
			packageID: 5,
			domain:    "package",
			err:       errors.New("could not find zone for package ID: 5"),
		},
		{
			name:      "InvalidDomainType",
			packageID: 0,
			domain:    "socket",
			err:       errors.New("unsupported rapl domain \"socket\""),
		},
		{
			name:      "DramSubzoneNotExist",
			packageID: 1,
			domain:    "dram",
			err:       errors.New("could not find dram subzone for package ID: 1"),
		},
		{
			name:      "PackageCurrEnergy",
			packageID: 0,
			domain:    "package",
			sample: attrSample{
				value:     206999074695.0,
				timestamp: fakeClock.Now(),
			},
			err: nil,
		},
		{
			name:      "DramCurrEnergy",
			packageID: 0,
			domain:    "dram",
			sample: attrSample{
				value:     64155753419.0,
				timestamp: fakeClock.Now(),
			},
			err: nil,
		},
	}

	r := &raplData{
		zones: map[int]powerZone{
			0: &zone{
				name: "package-0",
				path: "testdata/intel-rapl/intel-rapl:0",
				subzones: []powerZone{
					&zone{
						name:     "dram",
						path:     "testdata/intel-rapl/intel-rapl:0/intel-rapl:0:1",
						subzones: make([]powerZone, 0),
					},
				},
			},
			1: &zone{
				name:     "package-1",
				path:     "testdata/intel-rapl/intel-rapl:1",
				subzones: []powerZone{},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := r.setLastMeasuredEnergyAttribute(tc.packageID, tc.domain, tc.sample)
			if tc.err != nil {
				s.Require().ErrorContains(err, tc.err.Error())
			} else {
				outSample, err := r.getLastMeasuredEnergyAttribute(tc.packageID, tc.domain)
				s.Require().NoError(err)
				s.Require().Equal(tc.sample, outSample)
			}
		})
	}
}

// zoneMock represents a mock for raplData type. Implements raplReader interface.
type zoneMock struct {
	mock.Mock
}

func (m *zoneMock) getName() string {
	args := m.Called()
	return args.String(0)
}

func (m *zoneMock) getPath() string {
	args := m.Called()
	return args.String(0)
}

func (m *zoneMock) addSubzone(subzone powerZone) {
	m.Called(subzone)
}

func (m *zoneMock) getDomainSubzone(domain string) powerZone {
	args := m.Called(domain)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(powerZone)
}

func (m *zoneMock) getSubzones() []powerZone {
	args := m.Called()
	return args.Get(0).([]powerZone)
}

func (m *zoneMock) getEnergySample() attrSample {
	args := m.Called()
	return args.Get(0).(attrSample)
}

func (m *zoneMock) setEnergySample(e attrSample) {
	m.Called(e)
}

func (m *zoneMock) readAttribute(attribute string) (attrSample, error) {
	args := m.Called(attribute)
	return args.Get(0).(attrSample), args.Error(1)
}

func (s *raplTimeSensitiveTestSuite) TestGetCurrentPowerConsumptionWatt() {
	s.Run("InvalidPackageID", func() {
		packageID := 1
		domain := packageDomain.String()
		expPower := 0.0
		errMsg := fmt.Sprintf("error getting last measured current energy attribute for %q domain: could not find zone for package ID: %v", domain, packageID)

		m := &zoneMock{}
		r := &raplData{
			zones: map[int]powerZone{
				0: &zone{},
			},
		}

		outPower, err := r.getCurrentPowerConsumptionWatts(packageID, domain)
		s.Require().Equal(expPower, outPower)
		s.Require().ErrorContains(err, errMsg)
		m.AssertExpectations(s.T())
	})

	s.Run("InvalidDomain", func() {
		packageID := 0
		domain := "socket"
		expPower := 0.0
		errMsg := fmt.Sprintf("error getting last measured current energy attribute for %q domain: unsupported rapl domain %q", domain, domain)

		m := &zoneMock{}
		r := &raplData{
			zones: map[int]powerZone{
				0: m,
			},
		}

		outPower, err := r.getCurrentPowerConsumptionWatts(packageID, domain)
		s.Require().Equal(expPower, outPower)
		s.Require().ErrorContains(err, errMsg)
		m.AssertExpectations(s.T())
	})

	s.Run("PackageCurrentEnergyAttrError", func() {
		packageID := 0
		pkg := packageDomain.String()
		energyAttr := currEnergyAttr.String()
		errMsg := fmt.Sprintf("error reading energy attribute %q", energyAttr)
		expPower := 0.0

		m := &zoneMock{}
		m.On("getEnergySample").Return(attrSample{1000, fakeClock.Now()}, nil).Once()
		m.On("readAttribute", energyAttr).Return(attrSample{}, errors.New(errMsg)).Once()
		r := &raplData{
			zones: map[int]powerZone{
				0: m,
			},
		}

		outPower, err := r.getCurrentPowerConsumptionWatts(packageID, pkg)
		s.Require().Equal(expPower, outPower)
		s.Require().ErrorContains(err, errMsg)
		m.AssertExpectations(s.T())
	})

	s.Run("PackageWithoutResetCount", func() {
		packageID := 0
		pkg := packageDomain.String()
		s1 := attrSample{4000000, fakeClock.Now()}
		s2 := attrSample{5000000, fakeClock.Now().Add(time.Second)}
		expPower := 1.0

		m := &zoneMock{}
		m.On("getEnergySample").Return(s1, nil).Once()
		m.On("readAttribute", currEnergyAttr.String()).Return(s2, nil).Once()
		m.On("setEnergySample", s2).Once()
		r := &raplData{
			zones: map[int]powerZone{
				0: m,
			},
		}

		outPower, err := r.getCurrentPowerConsumptionWatts(packageID, pkg)
		s.Require().Equal(expPower, outPower)
		s.Require().NoError(err)
		m.AssertExpectations(s.T())
	})

	s.Run("DramSetLastEnergyAttrError", func() {
		packageID := 0
		dram := dramDomain.String()
		s1 := attrSample{4000000, fakeClock.Now()}
		s2 := attrSample{5000000, fakeClock.Now().Add(time.Second)}
		errMsg := fmt.Sprintf("could not find dram subzone for package ID: %v", packageID)
		expPower := 0.0

		m := &zoneMock{}
		m.On("getDomainSubzone", dram).Return(m).Once()
		m.On("getEnergySample").Return(s1, nil).Once()
		m.On("getDomainSubzone", dram).Return(m).Once()
		m.On("readAttribute", currEnergyAttr.String()).Return(s2, nil).Once()
		m.On("getDomainSubzone", dram).Return(nil).Once()
		r := &raplData{
			zones: map[int]powerZone{
				0: m,
			},
		}

		outPower, err := r.getCurrentPowerConsumptionWatts(packageID, dram)
		s.Require().Equal(expPower, outPower)
		s.Require().ErrorContains(err, errMsg)
		m.AssertExpectations(s.T())
	})

	s.Run("PackageWithResetCountError", func() {
		packageID := 0
		pkg := packageDomain.String()
		expPower := 0.0
		energyAttr := maxEnergyAttr.String()
		errMsg := fmt.Sprintf("error reading energy attribute %q", energyAttr)

		m := &zoneMock{}
		m.On("getEnergySample").Return(attrSample{4000000, fakeClock.Now()}, nil).Once()
		m.On("readAttribute", currEnergyAttr.String()).Return(attrSample{1000000, fakeClock.Now().Add(time.Second)}, nil).Once()
		m.On("readAttribute", energyAttr).Return(attrSample{}, errors.New(errMsg)).Once()
		r := &raplData{
			zones: map[int]powerZone{
				0: m,
			},
		}

		outPower, err := r.getCurrentPowerConsumptionWatts(packageID, pkg)
		s.Require().Equal(expPower, outPower)
		s.Require().ErrorContains(err, errMsg)
		m.AssertExpectations(s.T())
	})

	s.Run("PackageWithResetCount", func() {
		packageID := 0
		domain := packageDomain.String()
		s1 := attrSample{4000000, fakeClock.Now()}
		s2 := attrSample{1000000, fakeClock.Now().Add(time.Second)}
		sMax := attrSample{4000000, time.Time{}}
		expPower := 1.0

		m := &zoneMock{}
		m.On("getEnergySample").Return(s1, nil).Once()
		m.On("readAttribute", currEnergyAttr.String()).Return(s2, nil).Once()
		m.On("readAttribute", maxEnergyAttr.String()).Return(sMax, nil).Once()
		m.On("setEnergySample", s2).Once()
		r := &raplData{
			zones: map[int]powerZone{
				0: m,
			},
		}

		outPower, err := r.getCurrentPowerConsumptionWatts(packageID, domain)
		s.Require().Equal(expPower, outPower)
		s.Require().NoError(err)
		m.AssertExpectations(s.T())
	})

	s.Run("DramWithoutResetCount", func() {
		packageID := 0
		dram := dramDomain.String()
		s1 := attrSample{3000000, fakeClock.Now()}
		s2 := attrSample{5000000, fakeClock.Now().Add(time.Second)}
		expPower := 2.0

		m := &zoneMock{}
		m.On("getDomainSubzone", dram).Return(m)
		m.On("getEnergySample").Return(s1, nil).Once()
		m.On("readAttribute", currEnergyAttr.String()).Return(s2, nil).Once()
		m.On("setEnergySample", s2).Once()
		r := &raplData{
			zones: map[int]powerZone{
				0: m,
			},
		}

		outPower, err := r.getCurrentPowerConsumptionWatts(packageID, dram)
		s.Require().Equal(expPower, outPower)
		s.Require().NoError(err)
		m.AssertExpectations(s.T())
	})

	s.Run("DramWithResetCount", func() {
		packageID := 0
		dram := dramDomain.String()
		s1 := attrSample{3000000, fakeClock.Now()}
		s2 := attrSample{1000000, fakeClock.Now().Add(time.Second)}
		sMax := attrSample{4000000, time.Time{}}
		expPower := 2.0

		m := &zoneMock{}
		m.On("getDomainSubzone", dram).Return(m)
		m.On("getEnergySample").Return(s1, nil).Once()
		m.On("readAttribute", currEnergyAttr.String()).Return(s2, nil).Once()
		m.On("readAttribute", maxEnergyAttr.String()).Return(sMax, nil).Once()
		m.On("setEnergySample", s2).Once()
		r := &raplData{
			zones: map[int]powerZone{
				0: m,
			},
		}

		outPower, err := r.getCurrentPowerConsumptionWatts(packageID, dram)
		s.Require().Equal(expPower, outPower)
		s.Require().NoError(err)
		m.AssertExpectations(s.T())
	})
}

func TestGetMaxPowerConstraintWatts(t *testing.T) {
	testCases := []struct {
		name      string
		packageID int
		power     float64
		err       error
	}{
		{
			name:      "InvalidPackageID",
			packageID: 4,
			power:     0.0,
			err:       errors.New("could not find zone for package ID: 4"),
		},
		{
			name:      "AttributeFileNotExist",
			packageID: 3,
			power:     0.0,
			err:       errors.New(`error reading file "` + makeTestDataPath("testdata/intel-rapl/intel-rapl:3/constraint_0_max_power_uw") + `"`),
		},
		{
			name:      "AttributeFileEmpty",
			packageID: 2,
			power:     0.0,
			err:       errors.New("error converting attribute file content to float64"),
		},
		{
			name:      "AttributeFileNonNumeric",
			packageID: 1,
			power:     0.0,
			err:       errors.New("error converting attribute file content to float64"),
		},
		{
			name:      "Valid",
			packageID: 0,
			power:     250.0,
			err:       nil,
		},
	}

	r := &raplData{
		basePath: makeTestDataPath("testdata/intel-rapl"),
	}
	require.NoError(t, r.initZoneMap())

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outPower, err := r.getMaxPowerConstraintWatts(tc.packageID)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.power, outPower)
			}
		})
	}
}

func TestGetPackageIDs(t *testing.T) {
	testCases := []struct {
		name       string
		raplZones  map[int]powerZone
		packageIDs []int
	}{
		{
			name:       "ZonesMapIsNil",
			raplZones:  nil,
			packageIDs: []int{},
		},
		{
			name:       "ZonesMapIsEmpty",
			raplZones:  map[int]powerZone{},
			packageIDs: []int{},
		},
		{
			name: "ZonesMapIsUnordered",
			raplZones: map[int]powerZone{
				1: &zone{},
				0: &zone{},
				4: &zone{},
				3: &zone{},
			},
			packageIDs: []int{0, 1, 3, 4},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := &raplData{
				zones: tc.raplZones,
			}
			require.Equal(t, tc.packageIDs, r.getPackageIDs())
		})
	}
}
