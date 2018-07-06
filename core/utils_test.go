// -*- Mode: Go; indent-tabs-mode: t -*-
// Flashback
// Copyright 2018 Canonical Ltd.  All rights reserved.

package core_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/CanonicalLtd/flashback/core"
	check "gopkg.in/check.v1"
)

func TestCore(t *testing.T) { check.TestingT(t) }

type SuiteTest struct {
	Func   string
	Input  string
	Input2 string
	Output string
	Error  bool
}

type coreSuite struct{}

var _ = check.Suite(&coreSuite{})

func (s *coreSuite) TestUtils(c *check.C) {
	tests := []SuiteTest{
		{"RootDeviceNameFromPath", "/dev/sdd1", "", "sdd", false},
		{"RootDeviceNameFromPath", "/dev/sdd", "", "sdd", false},
		{"RootDeviceNameFromPath", "/dev/mmcblk1p2", "", "mmcblk1", false},
		{"RootDeviceNameFromPath", "/dev/mmcblk2", "", "mmcblk2", false},
		{"RootDeviceNameFromPath", "/dev/mmcblk3p4", "", "mmcblk3", false},
		{"DevicePathFromDevice", "sdd1", "", "/dev/sdd1", false},
		{"DevicePathFromDevice", "sdd", "", "/dev/sdd", false},
		{"DevicePathFromDevice", "mmcblk3p4", "", "/dev/mmcblk3p4", false},
		{"DevicePathFromNumber", "/dev/sdd1", "3", "/dev/sdd3", false},
		{"DevicePathFromNumber", "/dev/mmcblk1p4", "3", "/dev/mmcblk1p3", false},
		{"DeviceNameFromPath", "/dev/sdd3", "", "sdd3", false},
		{"DeviceNameFromPath", "/dev/mmcblk1p4", "", "mmcblk1p4", false},
		{"DeviceNumberFromPath", "/dev/sdd3", "", "3", false},
		{"DeviceNumberFromPath", "/dev/mmcblk1p4", "", "4", false},
	}

	for _, t := range tests {
		var s string

		switch t.Func {
		case "RootDeviceNameFromPath":
			s = core.RootDeviceNameFromPath(t.Input)
		case "DevicePathFromDevice":
			s = core.DevicePathFromDevice(t.Input)
		case "DeviceNameFromPath":
			s = core.DeviceNameFromPath(t.Input)
		case "DevicePathFromNumber":
			i, _ := strconv.Atoi(t.Input2)
			s = core.DevicePathFromNumber(t.Input, i)
		case "DeviceNumberFromPath":
			i, err := core.DeviceNumberFromPath(t.Input)
			c.Assert(err, check.IsNil)
			s = fmt.Sprint(i)
		}
		c.Assert(s, check.Equals, t.Output)
	}
}
